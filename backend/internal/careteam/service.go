package careteam

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
)

// Service defines the business logic interface for care team management
type Service interface {
	AssignRole(ctx context.Context, r *http.Request, encounterID string, req AssignRoleRequest, currentUserID string) (*CareTeamAssignment, error)
	TransferRole(ctx context.Context, r *http.Request, assignmentID string, req TransferRoleRequest, currentUserID string) (*CareTeamAssignment, error)
	ListCareTeam(ctx context.Context, encounterID string, activeOnly bool) ([]CareTeamAssignment, error)
	GetHandoffs(ctx context.Context, encounterID string, limit, offset int) ([]HandoffNote, int64, error)
	GetCareTeamWithDetails(ctx context.Context, encounterID string) (*CareTeamResponse, error)
}

// service handles care team business logic
type service struct {
	repo Repository
	db   *database.DB
}

// NewService creates a new care team service
func NewService(repo Repository, db *database.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

// AssignRole assigns a user to a role in an encounter's care team
func (s *service) AssignRole(ctx context.Context, r *http.Request, encounterID string, req AssignRoleRequest, currentUserID string) (*CareTeamAssignment, error) {
	// Check if there's already an active assignment for this role
	existingAssignment, err := s.repo.GetActiveAssignmentByRole(ctx, encounterID, req.RoleType)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing assignment: %w", err)
	}

	if existingAssignment != nil {
		return nil, fmt.Errorf("role %s is already assigned to user %s; use transfer to reassign", req.RoleType, existingAssignment.UserID)
	}

	// Set default StartsAt to now if not provided
	startsAt := time.Now().UTC()
	if req.StartsAt != nil {
		startsAt = req.StartsAt.UTC()
	}

	// Create new assignment
	assignment := &CareTeamAssignment{
		EncounterID: encounterID,
		UserID:      req.UserID,
		RoleType:    req.RoleType,
		StartsAt:    startsAt,
		CreatedBy:   currentUserID,
	}

	if err := s.repo.CreateAssignment(ctx, assignment); err != nil {
		return nil, err
	}

	// Audit log
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "care_team_assignment",
		EntityID:   assignment.ID,
		Action:     "CREATE",
		ByUserID:   currentUserID,
		After:      assignment,
	})

	return assignment, nil
}

// TransferRole transfers a role from one user to another with handoff documentation
func (s *service) TransferRole(ctx context.Context, r *http.Request, assignmentID string, req TransferRoleRequest, currentUserID string) (*CareTeamAssignment, error) {
	// Get the current assignment
	currentAssignment, err := s.repo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	if !currentAssignment.IsActive() {
		return nil, fmt.Errorf("assignment is already ended")
	}

	// Validate handoff note for critical roles
	if CriticalRoles[currentAssignment.RoleType] && req.HandoffNote == "" {
		return nil, fmt.Errorf("handoff note is required for critical role: %s", currentAssignment.RoleType)
	}

	// Set default EndsAt to now if not provided
	endsAt := time.Now().UTC()
	if req.EndsAt != nil {
		endsAt = req.EndsAt.UTC()
	}

	// Create handoff note
	var handoffNoteID *string
	if req.HandoffNote != "" {
		handoffNote := &HandoffNote{
			EncounterID:          currentAssignment.EncounterID,
			FromUserID:           currentAssignment.UserID,
			ToUserID:             req.ToUserID,
			RoleType:             currentAssignment.RoleType,
			Note:                 req.HandoffNote,
			StructuredFieldsJSON: req.StructuredFields,
		}

		if err := s.repo.CreateHandoffNote(ctx, handoffNote); err != nil {
			return nil, fmt.Errorf("failed to create handoff note: %w", err)
		}

		handoffNoteID = &handoffNote.ID

		// Audit log handoff note
		audit.Log(ctx, s.db, r, audit.Entry{
			EntityType: "handoff_note",
			EntityID:   handoffNote.ID,
			Action:     "CREATE",
			ByUserID:   currentUserID,
			After:      handoffNote,
		})
	}

	// End the current assignment
	if err := s.repo.EndAssignment(ctx, assignmentID, endsAt); err != nil {
		return nil, err
	}

	// Audit log the end of assignment
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "care_team_assignment",
		EntityID:   assignmentID,
		Action:     "UPDATE",
		ByUserID:   currentUserID,
		Before:     currentAssignment,
		After:      map[string]interface{}{"ends_at": endsAt},
		Reason:     &req.HandoffNote,
	})

	// Create new assignment
	newAssignment := &CareTeamAssignment{
		EncounterID:   currentAssignment.EncounterID,
		UserID:        req.ToUserID,
		RoleType:      currentAssignment.RoleType,
		StartsAt:      endsAt,
		CreatedBy:     currentUserID,
		HandoffNoteID: handoffNoteID,
	}

	if err := s.repo.CreateAssignment(ctx, newAssignment); err != nil {
		return nil, fmt.Errorf("failed to create new assignment: %w", err)
	}

	// Audit log new assignment
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "care_team_assignment",
		EntityID:   newAssignment.ID,
		Action:     "CREATE",
		ByUserID:   currentUserID,
		After:      newAssignment,
		Reason:     &req.HandoffNote,
	})

	return newAssignment, nil
}

// ListCareTeam returns the current active care team for an encounter
func (s *service) ListCareTeam(ctx context.Context, encounterID string, activeOnly bool) ([]CareTeamAssignment, error) {
	if activeOnly {
		return s.repo.GetActiveAssignments(ctx, encounterID)
	}
	return s.repo.GetAssignmentHistory(ctx, encounterID)
}

// GetHandoffs returns handoff notes for an encounter with pagination
func (s *service) GetHandoffs(ctx context.Context, encounterID string, limit, offset int) ([]HandoffNote, int64, error) {
	return s.repo.GetHandoffNotesPaginated(ctx, encounterID, limit, offset)
}

// GetCareTeamWithDetails returns the current care team with user details populated
// This is a convenience method that joins with the users table
func (s *service) GetCareTeamWithDetails(ctx context.Context, encounterID string) (*CareTeamResponse, error) {
	assignments, err := s.repo.GetActiveAssignments(ctx, encounterID)
	if err != nil {
		return nil, err
	}

	members := make([]CareTeamMember, 0, len(assignments))
	for _, assignment := range assignments {
		// Fetch user details
		var user models.User
		if err := s.db.WithContext(ctx).Where("id = ?", assignment.UserID).First(&user).Error; err != nil {
			// Log warning but continue (don't fail entire request if one user lookup fails)
			continue
		}

		members = append(members, CareTeamMember{
			Assignment: assignment,
			UserName:   user.Name,
			UserEmail:  user.Email,
		})
	}

	return &CareTeamResponse{
		EncounterID: encounterID,
		Members:     members,
	}, nil
}
