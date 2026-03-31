package flow

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
)

// Service defines the business logic interface for flow tracking
type Service interface {
	RecordTransition(ctx context.Context, r *http.Request, encounterID string, req CreateTransitionRequest, currentUserID string) (*FlowStateTransition, error)
	OverrideTransition(ctx context.Context, r *http.Request, encounterID string, req OverrideTransitionRequest, currentUserID string, userRole models.Role) (*FlowStateTransition, error)
	GetTimeline(ctx context.Context, encounterID string) (*FlowTimelineResponse, error)
	GetTimelinePaginated(ctx context.Context, encounterID string, limit, offset int) (*FlowTimelineResponse, error)
	GetTimelineWithActors(ctx context.Context, encounterID string) (*FlowTimelineDetailResponse, error)
	GetCurrentState(ctx context.Context, encounterID string) (*FlowState, error)
}

// service handles flow tracking business logic
type service struct {
	repo Repository
	db   *database.DB
}

// NewService creates a new flow service
func NewService(repo Repository, db *database.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

// RecordTransition records a state transition with validation
func (s *service) RecordTransition(ctx context.Context, r *http.Request, encounterID string, req CreateTransitionRequest, currentUserID string) (*FlowStateTransition, error) {
	// Get current state
	currentTransition, err := s.repo.GetCurrentState(ctx, encounterID)
	if err != nil {
		return nil, err
	}

	var fromState *FlowState
	if currentTransition != nil {
		fromState = &currentTransition.ToState
	}

	// Validate transition
	if fromState != nil && !IsValidTransition(*fromState, req.ToState) {
		return nil, fmt.Errorf("invalid transition from %s to %s; use override endpoint if this is intentional", *fromState, req.ToState)
	}

	// Set default TransitionedAt to now if not provided
	transitionedAt := time.Now().UTC()
	if req.TransitionedAt != nil {
		transitionedAt = req.TransitionedAt.UTC()
	}

	// Create transition
	transition := &FlowStateTransition{
		EncounterID:    encounterID,
		FromState:      fromState,
		ToState:        req.ToState,
		TransitionedAt: transitionedAt,
		ActorType:      ActorTypeUser,
		ActorUserID:    &currentUserID,
		Reason:         req.Reason,
		IsOverride:     false,
	}

	if err := s.repo.CreateTransition(ctx, transition); err != nil {
		return nil, err
	}

	// Audit log
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "flow_state_transition",
		EntityID:   transition.ID,
		Action:     "CREATE",
		ByUserID:   currentUserID,
		After:      transition,
		Reason:     req.Reason,
	})

	return transition, nil
}

// OverrideTransition records a privileged state transition that bypasses validation
func (s *service) OverrideTransition(ctx context.Context, r *http.Request, encounterID string, req OverrideTransitionRequest, currentUserID string, userRole models.Role) (*FlowStateTransition, error) {
	// Only admin and operations roles can override
	if userRole != models.RoleAdmin && userRole != models.RoleOperations {
		return nil, fmt.Errorf("insufficient permissions to override flow transitions; requires admin or operations role")
	}

	// Reason is mandatory for overrides
	if req.Reason == "" {
		return nil, fmt.Errorf("reason is required for flow transition overrides")
	}

	// Get current state if not explicitly provided
	var fromState *FlowState
	if req.FromState != nil {
		fromState = req.FromState
	} else {
		currentTransition, err := s.repo.GetCurrentState(ctx, encounterID)
		if err != nil {
			return nil, err
		}
		if currentTransition != nil {
			fromState = &currentTransition.ToState
		}
	}

	// Set default TransitionedAt to now if not provided
	transitionedAt := time.Now().UTC()
	if req.TransitionedAt != nil {
		transitionedAt = req.TransitionedAt.UTC()
	}

	// Create override transition
	transition := &FlowStateTransition{
		EncounterID:    encounterID,
		FromState:      fromState,
		ToState:        req.ToState,
		TransitionedAt: transitionedAt,
		ActorType:      ActorTypeUser,
		ActorUserID:    &currentUserID,
		Reason:         &req.Reason,
		IsOverride:     true,
	}

	if err := s.repo.CreateTransition(ctx, transition); err != nil {
		return nil, err
	}

	// Audit log with special note for override
	reason := fmt.Sprintf("OVERRIDE: %s", req.Reason)
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "flow_state_transition",
		EntityID:   transition.ID,
		Action:     "OVERRIDE",
		ByUserID:   currentUserID,
		After:      transition,
		Reason:     &reason,
		Source:     "user_action",
	})

	return transition, nil
}

// GetTimeline returns the flow timeline for an encounter
func (s *service) GetTimeline(ctx context.Context, encounterID string) (*FlowTimelineResponse, error) {
	transitions, err := s.repo.GetTimeline(ctx, encounterID)
	if err != nil {
		return nil, err
	}

	var currentState *FlowState
	if len(transitions) > 0 {
		currentState = &transitions[len(transitions)-1].ToState
	}

	return &FlowTimelineResponse{
		EncounterID:  encounterID,
		CurrentState: currentState,
		Transitions:  transitions,
		Total:        int64(len(transitions)),
	}, nil
}

// GetTimelinePaginated returns the flow timeline with pagination
func (s *service) GetTimelinePaginated(ctx context.Context, encounterID string, limit, offset int) (*FlowTimelineResponse, error) {
	transitions, total, err := s.repo.GetTimelinePaginated(ctx, encounterID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Get current state separately
	var currentState *FlowState
	currentTransition, err := s.repo.GetCurrentState(ctx, encounterID)
	if err != nil {
		return nil, err
	}
	if currentTransition != nil {
		currentState = &currentTransition.ToState
	}

	return &FlowTimelineResponse{
		EncounterID:  encounterID,
		CurrentState: currentState,
		Transitions:  transitions,
		Total:        total,
	}, nil
}

// GetTimelineWithActors returns the timeline with actor details populated
func (s *service) GetTimelineWithActors(ctx context.Context, encounterID string) (*FlowTimelineDetailResponse, error) {
	transitions, err := s.repo.GetTimeline(ctx, encounterID)
	if err != nil {
		return nil, err
	}

	var currentState *FlowState
	if len(transitions) > 0 {
		currentState = &transitions[len(transitions)-1].ToState
	}

	// Populate actor details
	transitionsWithActors := make([]TransitionWithActor, 0, len(transitions))
	for _, t := range transitions {
		twa := TransitionWithActor{
			FlowStateTransition: t,
		}

		// Fetch user details if actor is user
		if t.ActorType == ActorTypeUser && t.ActorUserID != nil {
			var user models.User
			if err := s.db.WithContext(ctx).Where("id = ?", *t.ActorUserID).First(&user).Error; err == nil {
				twa.ActorName = &user.Name
				twa.ActorEmail = &user.Email
			}
		}

		transitionsWithActors = append(transitionsWithActors, twa)
	}

	return &FlowTimelineDetailResponse{
		EncounterID:  encounterID,
		CurrentState: currentState,
		Transitions:  transitionsWithActors,
		Total:        int64(len(transitionsWithActors)),
	}, nil
}

// GetCurrentState returns the current flow state for an encounter
func (s *service) GetCurrentState(ctx context.Context, encounterID string) (*FlowState, error) {
	transition, err := s.repo.GetCurrentState(ctx, encounterID)
	if err != nil {
		return nil, err
	}
	if transition == nil {
		return nil, nil
	}
	return &transition.ToState, nil
}
