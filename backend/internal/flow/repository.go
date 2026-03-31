package flow

import (
	"context"
	"fmt"
	"time"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// Repository defines the data access interface for flow state transitions
type Repository interface {
	CreateTransition(ctx context.Context, transition *FlowStateTransition) error
	GetTimeline(ctx context.Context, encounterID string) ([]FlowStateTransition, error)
	GetTimelinePaginated(ctx context.Context, encounterID string, limit, offset int) ([]FlowStateTransition, int64, error)
	GetCurrentState(ctx context.Context, encounterID string) (*FlowStateTransition, error)
	GetTransitionByID(ctx context.Context, transitionID string) (*FlowStateTransition, error)
	GetTransitionsSince(ctx context.Context, encounterID string, since time.Time) ([]FlowStateTransition, error)
	GetTransitionsByState(ctx context.Context, encounterID string, state FlowState) ([]FlowStateTransition, error)
}

// repository handles data access for flow state transitions
type repository struct {
	db *database.DB
}

// NewRepository creates a new flow repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

// CreateTransition creates a new flow state transition
func (r *repository) CreateTransition(ctx context.Context, transition *FlowStateTransition) error {
	if err := r.db.WithContext(ctx).Create(transition).Error; err != nil {
		return fmt.Errorf("failed to create flow transition: %w", err)
	}
	return nil
}

// GetTimeline returns all transitions for an encounter, ordered chronologically
func (r *repository) GetTimeline(ctx context.Context, encounterID string) ([]FlowStateTransition, error) {
	var transitions []FlowStateTransition
	err := r.db.WithContext(ctx).
		Where("encounter_id = ?", encounterID).
		Order("transitioned_at ASC, created_at ASC").
		Find(&transitions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get flow timeline: %w", err)
	}
	return transitions, nil
}

// GetTimelinePaginated returns transitions with pagination
func (r *repository) GetTimelinePaginated(ctx context.Context, encounterID string, limit, offset int) ([]FlowStateTransition, int64, error) {
	var transitions []FlowStateTransition
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&FlowStateTransition{}).
		Where("encounter_id = ?", encounterID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count flow transitions: %w", err)
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Where("encounter_id = ?", encounterID).
		Order("transitioned_at DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transitions).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get flow timeline: %w", err)
	}

	return transitions, total, nil
}

// GetCurrentState returns the most recent state for an encounter
func (r *repository) GetCurrentState(ctx context.Context, encounterID string) (*FlowStateTransition, error) {
	var transition FlowStateTransition
	err := r.db.WithContext(ctx).
		Where("encounter_id = ?", encounterID).
		Order("transitioned_at DESC, created_at DESC").
		First(&transition).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // No transitions yet
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current state: %w", err)
	}
	return &transition, nil
}

// GetTransitionByID retrieves a specific transition by ID
func (r *repository) GetTransitionByID(ctx context.Context, transitionID string) (*FlowStateTransition, error) {
	var transition FlowStateTransition
	err := r.db.WithContext(ctx).
		Where("id = ?", transitionID).
		First(&transition).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("transition not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transition: %w", err)
	}
	return &transition, nil
}

// GetTransitionsSince returns transitions after a specific timestamp
func (r *repository) GetTransitionsSince(ctx context.Context, encounterID string, since time.Time) ([]FlowStateTransition, error) {
	var transitions []FlowStateTransition
	err := r.db.WithContext(ctx).
		Where("encounter_id = ? AND transitioned_at > ?", encounterID, since).
		Order("transitioned_at ASC, created_at ASC").
		Find(&transitions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get transitions since %v: %w", since, err)
	}
	return transitions, nil
}

// GetTransitionsByState returns transitions to a specific state
func (r *repository) GetTransitionsByState(ctx context.Context, encounterID string, state FlowState) ([]FlowStateTransition, error) {
	var transitions []FlowStateTransition
	err := r.db.WithContext(ctx).
		Where("encounter_id = ? AND to_state = ?", encounterID, state).
		Order("transitioned_at ASC, created_at ASC").
		Find(&transitions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get transitions by state: %w", err)
	}
	return transitions, nil
}
