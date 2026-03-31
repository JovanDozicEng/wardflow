package task

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
)

// Service defines the business logic interface for task management
type Service interface {
	CreateTask(ctx context.Context, r *http.Request, req CreateTaskRequest, currentUserID string) (*Task, error)
	UpdateTask(ctx context.Context, r *http.Request, taskID string, req UpdateTaskRequest, currentUserID string) (*Task, error)
	AssignTask(ctx context.Context, r *http.Request, taskID string, req AssignTaskRequest, currentUserID string, currentUserRole models.Role) (*Task, error)
	CompleteTask(ctx context.Context, r *http.Request, taskID string, req CompleteTaskRequest, currentUserID string, currentUserRole models.Role) (*Task, error)
	ListTasks(ctx context.Context, filter ListTasksFilter) ([]Task, int64, error)
	GetTaskHistory(ctx context.Context, taskID string, limit, offset int) ([]TaskAssignmentEvent, int64, error)
	GetTaskByID(ctx context.Context, taskID string) (*Task, error)
	GetTasksWithOwnerDetails(ctx context.Context, filter ListTasksFilter) ([]TaskWithOwner, int64, error)
}

// service handles task business logic
type service struct {
	repo Repository
	db   *database.DB
}

// NewService creates a new task service
func NewService(repo Repository, db *database.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

// CreateTask creates a new task with optional initial assignment
func (s *service) CreateTask(ctx context.Context, r *http.Request, req CreateTaskRequest, currentUserID string) (*Task, error) {
	// Set default priority if not provided
	priority := TaskPriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}

	task := &Task{
		ScopeType: req.ScopeType,
		ScopeID:   req.ScopeID,
		Title:     req.Title,
		Details:   req.Details,
		Status:    TaskStatusOpen,
		Priority:  priority,
		SLADueAt:  req.SLADueAt,
		CreatedBy: currentUserID,
	}

	// Handle initial assignment if provided
	if req.AssignTo != nil {
		task.CurrentOwnerID = req.AssignTo
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}

	// Record initial assignment event if assigned
	if req.AssignTo != nil {
		event := &TaskAssignmentEvent{
			TaskID:      task.ID,
			FromOwnerID: nil, // initial assignment
			ToOwnerID:   req.AssignTo,
			AssignedAt:  time.Now().UTC(),
			AssignedBy:  currentUserID,
		}
		if err := s.repo.CreateAssignmentEvent(ctx, event); err != nil {
			// Log but don't fail the task creation
			fmt.Printf("warning: failed to create initial assignment event: %v\n", err)
		}
	}

	// Audit log
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "task",
		EntityID:   task.ID,
		Action:     "CREATE",
		ByUserID:   currentUserID,
		After:      task,
	})

	return task, nil
}

// UpdateTask updates task fields
func (s *service) UpdateTask(ctx context.Context, r *http.Request, taskID string, req UpdateTaskRequest, currentUserID string) (*Task, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Store before state for audit
	before := *task

	// Update fields
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Details != nil {
		task.Details = req.Details
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.SLADueAt != nil {
		task.SLADueAt = req.SLADueAt
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	// Audit log
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "task",
		EntityID:   task.ID,
		Action:     "UPDATE",
		ByUserID:   currentUserID,
		Before:     before,
		After:      task,
	})

	return task, nil
}

// AssignTask assigns a task to a user and records the assignment event
func (s *service) AssignTask(ctx context.Context, r *http.Request, taskID string, req AssignTaskRequest, currentUserID string, currentUserRole models.Role) (*Task, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// RBAC: Check if user can reassign tasks
	// For now, allow admin, charge_nurse, and operations to reassign any task
	// Others can only assign unassigned tasks or take tasks assigned to them
	canReassignAny := currentUserRole == models.RoleAdmin ||
		currentUserRole == models.RoleChargeNurse ||
		currentUserRole == models.RoleOperations

	if !canReassignAny {
		// Check if task is currently assigned to someone else
		if task.CurrentOwnerID != nil && *task.CurrentOwnerID != currentUserID {
			return nil, fmt.Errorf("insufficient permissions to reassign task owned by another user")
		}
	}

	// Store before state
	before := *task

	// Record assignment event
	event := &TaskAssignmentEvent{
		TaskID:      taskID,
		FromOwnerID: task.CurrentOwnerID,
		ToOwnerID:   req.ToOwnerID,
		AssignedAt:  time.Now().UTC(),
		AssignedBy:  currentUserID,
		Reason:      req.Reason,
	}

	if err := s.repo.CreateAssignmentEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create assignment event: %w", err)
	}

	// Update task owner
	task.CurrentOwnerID = req.ToOwnerID
	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	// Audit log
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "task",
		EntityID:   task.ID,
		Action:     "ASSIGN",
		ByUserID:   currentUserID,
		Before:     before,
		After:      task,
		Reason:     req.Reason,
	})

	return task, nil
}

// CompleteTask marks a task as completed
func (s *service) CompleteTask(ctx context.Context, r *http.Request, taskID string, req CompleteTaskRequest, currentUserID string, currentUserRole models.Role) (*Task, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// RBAC: Check if user can complete this task
	// Admin and charge nurse can complete any task
	// Others can only complete tasks assigned to them or unassigned tasks
	canCompleteAny := currentUserRole == models.RoleAdmin || currentUserRole == models.RoleChargeNurse

	if !canCompleteAny {
		if task.CurrentOwnerID != nil && *task.CurrentOwnerID != currentUserID {
			return nil, fmt.Errorf("cannot complete task assigned to another user")
		}
	}

	// Store before state
	before := *task

	// Mark as completed
	now := time.Now().UTC()
	task.Status = TaskStatusCompleted
	task.CompletedBy = &currentUserID
	task.CompletedAt = &now

	// Append completion notes to details if provided
	if req.CompletionNotes != nil {
		completionNote := fmt.Sprintf("\n\n[Completed: %s]", *req.CompletionNotes)
		if task.Details == nil {
			task.Details = &completionNote
		} else {
			updatedDetails := *task.Details + completionNote
			task.Details = &updatedDetails
		}
	}

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}

	// Audit log
	reason := "Task completed"
	if req.CompletionNotes != nil {
		reason = *req.CompletionNotes
	}
	audit.Log(ctx, s.db, r, audit.Entry{
		EntityType: "task",
		EntityID:   task.ID,
		Action:     "COMPLETE",
		ByUserID:   currentUserID,
		Before:     before,
		After:      task,
		Reason:     &reason,
	})

	return task, nil
}

// ListTasks retrieves tasks with filters
func (s *service) ListTasks(ctx context.Context, filter ListTasksFilter) ([]Task, int64, error) {
	return s.repo.List(ctx, filter)
}

// GetTaskHistory retrieves assignment history for a task
func (s *service) GetTaskHistory(ctx context.Context, taskID string, limit, offset int) ([]TaskAssignmentEvent, int64, error) {
	return s.repo.GetAssignmentHistoryPaginated(ctx, taskID, limit, offset)
}

// GetTaskByID retrieves a task by ID
func (s *service) GetTaskByID(ctx context.Context, taskID string) (*Task, error) {
	return s.repo.GetByID(ctx, taskID)
}

// GetTasksWithOwnerDetails retrieves tasks with owner details populated
func (s *service) GetTasksWithOwnerDetails(ctx context.Context, filter ListTasksFilter) ([]TaskWithOwner, int64, error) {
	tasks, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	tasksWithOwner := make([]TaskWithOwner, 0, len(tasks))
	for _, t := range tasks {
		two := TaskWithOwner{Task: t}

		// Fetch owner details if task is assigned
		if t.CurrentOwnerID != nil {
			var user models.User
			if err := s.db.WithContext(ctx).Where("id = ?", *t.CurrentOwnerID).First(&user).Error; err == nil {
				two.OwnerName = &user.Name
				two.OwnerEmail = &user.Email
			}
		}

		tasksWithOwner = append(tasksWithOwner, two)
	}

	return tasksWithOwner, total, nil
}
