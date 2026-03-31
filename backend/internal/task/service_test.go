package task

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
)

// mockRepository is a mock implementation of Repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, task *Task) error {
	args := m.Called(ctx, task)
	// Set ID for created task
	if args.Error(0) == nil && task.ID == "" {
		task.ID = "task-123"
	}
	return args.Error(0)
}

func (m *mockRepository) Update(ctx context.Context, task *Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockRepository) GetByID(ctx context.Context, taskID string) (*Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Task), args.Error(1)
}

func (m *mockRepository) List(ctx context.Context, filter ListTasksFilter) ([]Task, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]Task), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) CreateAssignmentEvent(ctx context.Context, event *TaskAssignmentEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockRepository) GetAssignmentHistory(ctx context.Context, taskID string) ([]TaskAssignmentEvent, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TaskAssignmentEvent), args.Error(1)
}

func (m *mockRepository) GetAssignmentHistoryPaginated(ctx context.Context, taskID string, limit, offset int) ([]TaskAssignmentEvent, int64, error) {
	args := m.Called(ctx, taskID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]TaskAssignmentEvent), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) Delete(ctx context.Context, taskID string) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func TestCreateTask_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	req := CreateTaskRequest{
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Check vitals",
	}

	repo.On("Create", ctx, mock.MatchedBy(func(t *Task) bool {
		return t.ScopeType == ScopeTypeEncounter &&
			t.ScopeID == "encounter-1" &&
			t.Title == "Check vitals" &&
			t.Status == TaskStatusOpen &&
			t.Priority == TaskPriorityMedium &&
			t.CreatedBy == "user-1"
	})).Return(nil)

	result, err := svc.CreateTask(ctx, r, req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Check vitals", result.Title)
	assert.Equal(t, TaskStatusOpen, result.Status)
	assert.Equal(t, TaskPriorityMedium, result.Priority)
	repo.AssertExpectations(t)
}

func TestCreateTask_WithAssignment(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	assignTo := "user-2"
	priority := TaskPriorityHigh
	req := CreateTaskRequest{
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Urgent task",
		Priority:  &priority,
		AssignTo:  &assignTo,
	}

	repo.On("Create", ctx, mock.MatchedBy(func(t *Task) bool {
		return t.CurrentOwnerID != nil && *t.CurrentOwnerID == "user-2" &&
			t.Priority == TaskPriorityHigh
	})).Return(nil)

	repo.On("CreateAssignmentEvent", ctx, mock.MatchedBy(func(e *TaskAssignmentEvent) bool {
		return e.ToOwnerID != nil && *e.ToOwnerID == "user-2" &&
			e.FromOwnerID == nil
	})).Return(nil)

	result, err := svc.CreateTask(ctx, r, req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-2", *result.CurrentOwnerID)
	repo.AssertExpectations(t)
}

func TestUpdateTask_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPatch, "/", nil)

	existingTask := &Task{
		ID:        "task-1",
		Title:     "Old title",
		Status:    TaskStatusOpen,
		Priority:  TaskPriorityMedium,
		CreatedBy: "user-1",
	}

	newTitle := "New title"
	newStatus := TaskStatusInProgress
	req := UpdateTaskRequest{
		Title:  &newTitle,
		Status: &newStatus,
	}

	repo.On("GetByID", ctx, "task-1").Return(existingTask, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(t *Task) bool {
		return t.ID == "task-1" &&
			t.Title == "New title" &&
			t.Status == TaskStatusInProgress
	})).Return(nil)

	result, err := svc.UpdateTask(ctx, r, "task-1", req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New title", result.Title)
	assert.Equal(t, TaskStatusInProgress, result.Status)
	repo.AssertExpectations(t)
}

func TestAssignTask_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	existingTask := &Task{
		ID:             "task-1",
		Title:          "Task",
		Status:         TaskStatusOpen,
		CurrentOwnerID: nil,
		CreatedBy:      "user-1",
	}

	toOwner := "user-2"
	req := AssignTaskRequest{
		ToOwnerID: &toOwner,
	}

	repo.On("GetByID", ctx, "task-1").Return(existingTask, nil)
	repo.On("CreateAssignmentEvent", ctx, mock.MatchedBy(func(e *TaskAssignmentEvent) bool {
		return e.TaskID == "task-1" &&
			e.ToOwnerID != nil && *e.ToOwnerID == "user-2" &&
			e.FromOwnerID == nil
	})).Return(nil)
	repo.On("Update", ctx, mock.MatchedBy(func(t *Task) bool {
		return t.CurrentOwnerID != nil && *t.CurrentOwnerID == "user-2"
	})).Return(nil)

	result, err := svc.AssignTask(ctx, r, "task-1", req, "user-1", models.RoleNurse)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-2", *result.CurrentOwnerID)
	repo.AssertExpectations(t)
}

func TestAssignTask_InsufficientPermissions(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	currentOwner := "user-2"
	existingTask := &Task{
		ID:             "task-1",
		Title:          "Task",
		Status:         TaskStatusOpen,
		CurrentOwnerID: &currentOwner,
		CreatedBy:      "user-2",
	}

	toOwner := "user-3"
	req := AssignTaskRequest{
		ToOwnerID: &toOwner,
	}

	repo.On("GetByID", ctx, "task-1").Return(existingTask, nil)

	result, err := svc.AssignTask(ctx, r, "task-1", req, "user-1", models.RoleNurse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insufficient permissions")
	repo.AssertExpectations(t)
}

func TestAssignTask_AdminCanReassign(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	currentOwner := "user-2"
	existingTask := &Task{
		ID:             "task-1",
		Title:          "Task",
		Status:         TaskStatusOpen,
		CurrentOwnerID: &currentOwner,
		CreatedBy:      "user-2",
	}

	toOwner := "user-3"
	req := AssignTaskRequest{
		ToOwnerID: &toOwner,
	}

	repo.On("GetByID", ctx, "task-1").Return(existingTask, nil)
	repo.On("CreateAssignmentEvent", ctx, mock.Anything).Return(nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)

	result, err := svc.AssignTask(ctx, r, "task-1", req, "admin-1", models.RoleAdmin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	repo.AssertExpectations(t)
}

func TestCompleteTask_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	currentOwner := "user-1"
	existingTask := &Task{
		ID:             "task-1",
		Title:          "Task",
		Status:         TaskStatusInProgress,
		CurrentOwnerID: &currentOwner,
		CreatedBy:      "user-1",
	}

	notes := "Task completed successfully"
	req := CompleteTaskRequest{
		CompletionNotes: &notes,
	}

	repo.On("GetByID", ctx, "task-1").Return(existingTask, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(t *Task) bool {
		return t.Status == TaskStatusCompleted &&
			t.CompletedBy != nil && *t.CompletedBy == "user-1" &&
			t.CompletedAt != nil
	})).Return(nil)

	result, err := svc.CompleteTask(ctx, r, "task-1", req, "user-1", models.RoleNurse)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusCompleted, result.Status)
	assert.NotNil(t, result.CompletedBy)
	assert.Equal(t, "user-1", *result.CompletedBy)
	repo.AssertExpectations(t)
}

func TestCompleteTask_NotOwner(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	currentOwner := "user-2"
	existingTask := &Task{
		ID:             "task-1",
		Title:          "Task",
		Status:         TaskStatusInProgress,
		CurrentOwnerID: &currentOwner,
		CreatedBy:      "user-2",
	}

	req := CompleteTaskRequest{}

	repo.On("GetByID", ctx, "task-1").Return(existingTask, nil)

	result, err := svc.CompleteTask(ctx, r, "task-1", req, "user-1", models.RoleNurse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot complete task assigned to another user")
	repo.AssertExpectations(t)
}

func TestCompleteTask_AdminCanComplete(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	currentOwner := "user-2"
	existingTask := &Task{
		ID:             "task-1",
		Title:          "Task",
		Status:         TaskStatusInProgress,
		CurrentOwnerID: &currentOwner,
		CreatedBy:      "user-2",
	}

	req := CompleteTaskRequest{}

	repo.On("GetByID", ctx, "task-1").Return(existingTask, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)

	result, err := svc.CompleteTask(ctx, r, "task-1", req, "admin-1", models.RoleAdmin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TaskStatusCompleted, result.Status)
	repo.AssertExpectations(t)
}

func TestListTasks_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := ListTasksFilter{
		Limit:  10,
		Offset: 0,
	}

	tasks := []Task{
		{ID: "task-1", Title: "Task 1"},
		{ID: "task-2", Title: "Task 2"},
	}

	repo.On("List", ctx, filter).Return(tasks, int64(2), nil)

	result, total, err := svc.ListTasks(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	repo.AssertExpectations(t)
}

func TestGetTaskHistory_Paginated(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	events := []TaskAssignmentEvent{
		{ID: "e1", TaskID: "task-1"},
		{ID: "e2", TaskID: "task-1"},
	}

	repo.On("GetAssignmentHistoryPaginated", ctx, "task-1", 10, 0).Return(events, int64(5), nil)

	result, total, err := svc.GetTaskHistory(ctx, "task-1", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(5), total)
	repo.AssertExpectations(t)
}

func TestGetTaskByID_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	task := &Task{
		ID:    "task-1",
		Title: "Task 1",
	}

	repo.On("GetByID", ctx, "task-1").Return(task, nil)

	result, err := svc.GetTaskByID(ctx, "task-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "task-1", result.ID)
	repo.AssertExpectations(t)
}

func TestGetTaskByID_NotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	repo.On("GetByID", ctx, "task-999").Return(nil, fmt.Errorf("task not found"))

	result, err := svc.GetTaskByID(ctx, "task-999")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertExpectations(t)
}

func TestGetTasksWithOwnerDetails_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := ListTasksFilter{Limit: 10}

	tasks := []Task{
		{ID: "task-1", Title: "Task 1", CurrentOwnerID: nil}, // No owner
	}

	repo.On("List", ctx, filter).Return(tasks, int64(1), nil)

	result, total, err := svc.GetTasksWithOwnerDetails(ctx, filter)

	// Should succeed since task has no owner (no DB lookup needed)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
	assert.Nil(t, result[0].OwnerName) // No owner
	repo.AssertExpectations(t)
}

// Additional tests for GetTasksWithOwnerDetails

func TestGetTasksWithOwnerDetails_WithEncounterFilter(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	encounterID := "encounter-1"
	scopeType := ScopeTypeEncounter
	filter := ListTasksFilter{
		ScopeType: &scopeType,
		ScopeID:   &encounterID,
		Limit:     10,
	}

	tasks := []Task{
		{ID: "task-1", ScopeType: ScopeTypeEncounter, ScopeID: encounterID, CurrentOwnerID: nil},
		{ID: "task-2", ScopeType: ScopeTypeEncounter, ScopeID: encounterID, CurrentOwnerID: nil},
	}

	repo.On("List", ctx, filter).Return(tasks, int64(2), nil)

	result, total, err := svc.GetTasksWithOwnerDetails(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, encounterID, result[0].Task.ScopeID)
	repo.AssertExpectations(t)
}

func TestGetTasksWithOwnerDetails_WithUnitFilter(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	unitID := "unit-1"
	scopeType := ScopeTypeUnit
	filter := ListTasksFilter{
		ScopeType: &scopeType,
		ScopeID:   &unitID,
		Limit:     10,
	}

	tasks := []Task{
		{ID: "task-1", ScopeType: ScopeTypeUnit, ScopeID: unitID, CurrentOwnerID: nil},
	}

	repo.On("List", ctx, filter).Return(tasks, int64(1), nil)

	result, total, err := svc.GetTasksWithOwnerDetails(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, unitID, result[0].Task.ScopeID)
	repo.AssertExpectations(t)
}

func TestGetTasksWithOwnerDetails_WithStatusFilter(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	status := TaskStatusOpen
	filter := ListTasksFilter{
		Status: &status,
		Limit:  10,
	}

	tasks := []Task{
		{ID: "task-1", Status: TaskStatusOpen, CurrentOwnerID: nil},
	}

	repo.On("List", ctx, filter).Return(tasks, int64(1), nil)

	result, _, err := svc.GetTasksWithOwnerDetails(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, TaskStatusOpen, result[0].Task.Status)
	repo.AssertExpectations(t)
}

func TestGetTasksWithOwnerDetails_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := ListTasksFilter{Limit: 10}

	repo.On("List", ctx, filter).Return(nil, int64(0), fmt.Errorf("database error"))

	result, total, err := svc.GetTasksWithOwnerDetails(ctx, filter)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}
