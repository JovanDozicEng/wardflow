package task

import (
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusOpen       TaskStatus = "open"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusEscalated  TaskStatus = "escalated"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// ScopeType represents what the task is scoped to
type ScopeType string

const (
	ScopeTypeEncounter ScopeType = "encounter"
	ScopeTypePatient   ScopeType = "patient"
	ScopeTypeUnit      ScopeType = "unit"
)

// Task represents a clinical task
type Task struct {
	ID             string        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ScopeType      ScopeType     `json:"scopeType" gorm:"type:varchar(20);not null;index:idx_task_scope"`
	ScopeID        string        `json:"scopeId" gorm:"type:varchar(100);not null;index:idx_task_scope"`
	Title          string        `json:"title" gorm:"type:text;not null"`
	Details        *string       `json:"details,omitempty" gorm:"type:text"`
	Status         TaskStatus    `json:"status" gorm:"type:varchar(50);not null;index:idx_task_status;default:'open'"`
	Priority       TaskPriority  `json:"priority" gorm:"type:varchar(20);not null;default:'medium'"`
	SLADueAt       *time.Time    `json:"slaDueAt,omitempty" gorm:"index:idx_task_sla"`
	CurrentOwnerID *string       `json:"currentOwnerId,omitempty" gorm:"type:uuid;index:idx_task_owner"`
	CreatedBy      string        `json:"createdBy" gorm:"type:uuid;not null"`
	CreatedAt      time.Time     `json:"createdAt" gorm:"index:idx_task_created"`
	CompletedBy    *string       `json:"completedBy,omitempty" gorm:"type:uuid"`
	CompletedAt    *time.Time    `json:"completedAt,omitempty"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (Task) TableName() string {
	return "tasks"
}

// IsOverdue returns true if the task has an SLA and it has passed
func (t *Task) IsOverdue() bool {
	if t.SLADueAt == nil {
		return false
	}
	if t.Status == TaskStatusCompleted || t.Status == TaskStatusCancelled {
		return false
	}
	return time.Now().UTC().After(*t.SLADueAt)
}

// TaskAssignmentEvent represents an immutable assignment history event
// This table is append-only; events are never updated or deleted
type TaskAssignmentEvent struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskID      string     `json:"taskId" gorm:"type:uuid;not null;index:idx_task_assignment_task"`
	FromOwnerID *string    `json:"fromOwnerId,omitempty" gorm:"type:uuid"` // null for initial assignment
	ToOwnerID   *string    `json:"toOwnerId,omitempty" gorm:"type:uuid"`   // null for unassignment
	AssignedAt  time.Time  `json:"assignedAt" gorm:"not null;index:idx_task_assignment_at"`
	AssignedBy  string     `json:"assignedBy" gorm:"type:uuid;not null"`
	Reason      *string    `json:"reason,omitempty" gorm:"type:text"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// TableName returns the table name for GORM
func (TaskAssignmentEvent) TableName() string {
	return "task_assignment_events"
}

// --- Request/Response DTOs ---

// CreateTaskRequest is the request body for creating a task
type CreateTaskRequest struct {
	ScopeType ScopeType     `json:"scopeType" binding:"required"`
	ScopeID   string        `json:"scopeId" binding:"required"`
	Title     string        `json:"title" binding:"required"`
	Details   *string       `json:"details,omitempty"`
	Priority  *TaskPriority `json:"priority,omitempty"` // defaults to medium
	SLADueAt  *time.Time    `json:"slaDueAt,omitempty"`
	AssignTo  *string       `json:"assignTo,omitempty"` // optional initial assignment
}

// UpdateTaskRequest is the request body for updating a task
type UpdateTaskRequest struct {
	Title    *string       `json:"title,omitempty"`
	Details  *string       `json:"details,omitempty"`
	Status   *TaskStatus   `json:"status,omitempty"`
	Priority *TaskPriority `json:"priority,omitempty"`
	SLADueAt *time.Time    `json:"slaDueAt,omitempty"`
}

// AssignTaskRequest is the request body for assigning a task
type AssignTaskRequest struct {
	ToOwnerID *string `json:"toOwnerId" binding:"required"` // null to unassign
	Reason    *string `json:"reason,omitempty"`
}

// CompleteTaskRequest is the request body for completing a task
type CompleteTaskRequest struct {
	CompletionNotes *string `json:"completionNotes,omitempty"`
}

// ListTasksFilter holds filters for listing tasks
type ListTasksFilter struct {
	ScopeType      *ScopeType
	ScopeID        *string
	Status         *TaskStatus
	Priority       *TaskPriority
	OwnerID        *string
	Overdue        *bool
	CreatedAfter   *time.Time
	CreatedBefore  *time.Time
	Limit          int
	Offset         int
}

// ListTasksResponse contains paginated tasks
type ListTasksResponse struct {
	Tasks  []Task `json:"tasks"`
	Total  int64  `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// TaskHistoryResponse contains assignment history for a task
type TaskHistoryResponse struct {
	TaskID  string                 `json:"taskId"`
	Events  []TaskAssignmentEvent  `json:"events"`
	Total   int64                  `json:"total"`
}

// TaskWithOwner includes owner details
type TaskWithOwner struct {
	Task
	OwnerName  *string `json:"ownerName,omitempty"`
	OwnerEmail *string `json:"ownerEmail,omitempty"`
}

// ListTasksDetailResponse includes owner details
type ListTasksDetailResponse struct {
	Tasks  []TaskWithOwner `json:"tasks"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}
