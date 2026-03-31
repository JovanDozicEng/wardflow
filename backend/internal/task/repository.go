package task

import (
	"context"
	"fmt"
	"time"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// Repository defines the data access interface for tasks
type Repository interface {
	Create(ctx context.Context, task *Task) error
	Update(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, taskID string) (*Task, error)
	List(ctx context.Context, filter ListTasksFilter) ([]Task, int64, error)
	CreateAssignmentEvent(ctx context.Context, event *TaskAssignmentEvent) error
	GetAssignmentHistory(ctx context.Context, taskID string) ([]TaskAssignmentEvent, error)
	GetAssignmentHistoryPaginated(ctx context.Context, taskID string, limit, offset int) ([]TaskAssignmentEvent, int64, error)
	Delete(ctx context.Context, taskID string) error
}

// repository handles data access for tasks
type repository struct {
	db *database.DB
}

// NewRepository creates a new task repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

// Create creates a new task
func (r *repository) Create(ctx context.Context, task *Task) error {
	if err := r.db.WithContext(ctx).Create(task).Error; err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

// Update updates a task
func (r *repository) Update(ctx context.Context, task *Task) error {
	if err := r.db.WithContext(ctx).Save(task).Error; err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

// GetByID retrieves a task by ID
func (r *repository) GetByID(ctx context.Context, taskID string) (*Task, error) {
	var task Task
	err := r.db.WithContext(ctx).
		Where("id = ?", taskID).
		First(&task).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &task, nil
}

// List retrieves tasks with filters and pagination
func (r *repository) List(ctx context.Context, filter ListTasksFilter) ([]Task, int64, error) {
	query := r.db.WithContext(ctx).Model(&Task{})

	// Apply filters
	if filter.ScopeType != nil {
		query = query.Where("scope_type = ?", *filter.ScopeType)
	}
	if filter.ScopeID != nil {
		query = query.Where("scope_id = ?", *filter.ScopeID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Priority != nil {
		query = query.Where("priority = ?", *filter.Priority)
	}
	if filter.OwnerID != nil {
		query = query.Where("current_owner_id = ?", *filter.OwnerID)
	}
	if filter.Overdue != nil && *filter.Overdue {
		query = query.Where("sla_due_at IS NOT NULL AND sla_due_at < ? AND status NOT IN (?, ?)",
			time.Now().UTC(), TaskStatusCompleted, TaskStatusCancelled)
	}
	if filter.CreatedAfter != nil {
		query = query.Where("created_at > ?", *filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		query = query.Where("created_at < ?", *filter.CreatedBefore)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Get paginated results
	var tasks []Task
	err := query.
		Order("created_at DESC").
		Limit(filter.Limit).
		Offset(filter.Offset).
		Find(&tasks).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, total, nil
}

// CreateAssignmentEvent creates a new task assignment event
func (r *repository) CreateAssignmentEvent(ctx context.Context, event *TaskAssignmentEvent) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("failed to create assignment event: %w", err)
	}
	return nil
}

// GetAssignmentHistory retrieves assignment history for a task
func (r *repository) GetAssignmentHistory(ctx context.Context, taskID string) ([]TaskAssignmentEvent, error) {
	var events []TaskAssignmentEvent
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("assigned_at DESC, created_at DESC").
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get assignment history: %w", err)
	}
	return events, nil
}

// GetAssignmentHistoryPaginated retrieves assignment history with pagination
func (r *repository) GetAssignmentHistoryPaginated(ctx context.Context, taskID string, limit, offset int) ([]TaskAssignmentEvent, int64, error) {
	var events []TaskAssignmentEvent
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).
		Model(&TaskAssignmentEvent{}).
		Where("task_id = ?", taskID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count assignment events: %w", err)
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Order("assigned_at DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get assignment history: %w", err)
	}

	return events, total, nil
}

// Delete soft deletes a task (if we add soft delete support later)
// For now, tasks are never deleted, only cancelled
func (r *repository) Delete(ctx context.Context, taskID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", taskID).
		Delete(&Task{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete task: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
