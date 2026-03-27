package careteam

import (
	"context"
	"fmt"
	"time"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// Repository handles data access for care team assignments
type Repository struct {
	db *database.DB
}

// NewRepository creates a new care team repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// CreateAssignment creates a new care team assignment
func (r *Repository) CreateAssignment(ctx context.Context, assignment *CareTeamAssignment) error {
	if err := r.db.WithContext(ctx).Create(assignment).Error; err != nil {
		return fmt.Errorf("failed to create care team assignment: %w", err)
	}
	return nil
}

// EndAssignment ends an active assignment by setting EndsAt timestamp
func (r *Repository) EndAssignment(ctx context.Context, assignmentID string, endsAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&CareTeamAssignment{}).
		Where("id = ? AND ends_at IS NULL", assignmentID).
		Update("ends_at", endsAt)

	if result.Error != nil {
		return fmt.Errorf("failed to end assignment: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("assignment not found or already ended")
	}
	return nil
}

// GetActiveAssignments returns all active assignments for an encounter
func (r *Repository) GetActiveAssignments(ctx context.Context, encounterID string) ([]CareTeamAssignment, error) {
	var assignments []CareTeamAssignment
	err := r.db.WithContext(ctx).
		Where("encounter_id = ? AND ends_at IS NULL", encounterID).
		Order("created_at ASC").
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get active assignments: %w", err)
	}
	return assignments, nil
}

// GetActiveAssignmentByRole returns the current active assignment for a specific role
func (r *Repository) GetActiveAssignmentByRole(ctx context.Context, encounterID string, roleType RoleType) (*CareTeamAssignment, error) {
	var assignment CareTeamAssignment
	err := r.db.WithContext(ctx).
		Where("encounter_id = ? AND role_type = ? AND ends_at IS NULL", encounterID, roleType).
		First(&assignment).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active assignment by role: %w", err)
	}
	return &assignment, nil
}

// GetAssignmentHistory returns all assignments (active and ended) for an encounter
func (r *Repository) GetAssignmentHistory(ctx context.Context, encounterID string) ([]CareTeamAssignment, error) {
	var assignments []CareTeamAssignment
	err := r.db.WithContext(ctx).
		Preload("HandoffNote").
		Where("encounter_id = ?", encounterID).
		Order("starts_at DESC").
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get assignment history: %w", err)
	}
	return assignments, nil
}

// GetAssignmentByID retrieves a specific assignment by ID
func (r *Repository) GetAssignmentByID(ctx context.Context, assignmentID string) (*CareTeamAssignment, error) {
	var assignment CareTeamAssignment
	err := r.db.WithContext(ctx).
		Preload("HandoffNote").
		Where("id = ?", assignmentID).
		First(&assignment).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("assignment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment: %w", err)
	}
	return &assignment, nil
}

// CreateHandoffNote creates a new handoff note
func (r *Repository) CreateHandoffNote(ctx context.Context, note *HandoffNote) error {
	if err := r.db.WithContext(ctx).Create(note).Error; err != nil {
		return fmt.Errorf("failed to create handoff note: %w", err)
	}
	return nil
}

// GetHandoffNotes returns all handoff notes for an encounter
func (r *Repository) GetHandoffNotes(ctx context.Context, encounterID string) ([]HandoffNote, error) {
	var notes []HandoffNote
	err := r.db.WithContext(ctx).
		Where("encounter_id = ?", encounterID).
		Order("created_at DESC").
		Find(&notes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get handoff notes: %w", err)
	}
	return notes, nil
}

// GetHandoffNotesPaginated returns handoff notes with pagination
func (r *Repository) GetHandoffNotesPaginated(ctx context.Context, encounterID string, limit, offset int) ([]HandoffNote, int64, error) {
	var notes []HandoffNote
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&HandoffNote{}).
		Where("encounter_id = ?", encounterID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count handoff notes: %w", err)
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Where("encounter_id = ?", encounterID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notes).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get handoff notes: %w", err)
	}

	return notes, total, nil
}
