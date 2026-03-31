package task

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
)

// mockService is a mock implementation of Service
type mockService struct {
	mock.Mock
}

func (m *mockService) CreateTask(ctx context.Context, r *http.Request, req CreateTaskRequest, currentUserID string) (*Task, error) {
	args := m.Called(ctx, r, req, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Task), args.Error(1)
}

func (m *mockService) UpdateTask(ctx context.Context, r *http.Request, taskID string, req UpdateTaskRequest, currentUserID string) (*Task, error) {
	args := m.Called(ctx, r, taskID, req, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Task), args.Error(1)
}

func (m *mockService) AssignTask(ctx context.Context, r *http.Request, taskID string, req AssignTaskRequest, currentUserID string, currentUserRole models.Role) (*Task, error) {
	args := m.Called(ctx, r, taskID, req, currentUserID, currentUserRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Task), args.Error(1)
}

func (m *mockService) CompleteTask(ctx context.Context, r *http.Request, taskID string, req CompleteTaskRequest, currentUserID string, currentUserRole models.Role) (*Task, error) {
	args := m.Called(ctx, r, taskID, req, currentUserID, currentUserRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Task), args.Error(1)
}

func (m *mockService) ListTasks(ctx context.Context, filter ListTasksFilter) ([]Task, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]Task), args.Get(1).(int64), args.Error(2)
}

func (m *mockService) GetTaskHistory(ctx context.Context, taskID string, limit, offset int) ([]TaskAssignmentEvent, int64, error) {
	args := m.Called(ctx, taskID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]TaskAssignmentEvent), args.Get(1).(int64), args.Error(2)
}

func (m *mockService) GetTaskByID(ctx context.Context, taskID string) (*Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Task), args.Error(1)
}

func (m *mockService) GetTasksWithOwnerDetails(ctx context.Context, filter ListTasksFilter) ([]TaskWithOwner, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]TaskWithOwner), args.Get(1).(int64), args.Error(2)
}

func TestHandler_ListTasks_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	tasks := []Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusOpen},
		{ID: "task-2", Title: "Task 2", Status: TaskStatusInProgress},
	}

	svc.On("ListTasks", mock.Anything, mock.MatchedBy(func(f ListTasksFilter) bool {
		return f.Limit == 30 && f.Offset == 0
	})).Return(tasks, int64(2), nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks", nil, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.ListTasks(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListTasksResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Len(t, response.Tasks, 2)
	assert.Equal(t, int64(2), response.Total)
	svc.AssertExpectations(t)
}

func TestHandler_ListTasks_WithFilters(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	tasks := []Task{
		{ID: "task-1", Title: "Task 1", Status: TaskStatusOpen},
	}

	svc.On("ListTasks", mock.Anything, mock.MatchedBy(func(f ListTasksFilter) bool {
		return f.Status != nil && *f.Status == TaskStatusOpen &&
			f.Priority != nil && *f.Priority == TaskPriorityHigh &&
			f.Limit == 10
	})).Return(tasks, int64(1), nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks?status=open&priority=high&limit=10", nil, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.ListTasks(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListTasksResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Len(t, response.Tasks, 1)
	svc.AssertExpectations(t)
}

func TestHandler_ListTasks_WithOwnerDetails(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	ownerName := "John Doe"
	tasksWithOwner := []TaskWithOwner{
		{Task: Task{ID: "task-1", Title: "Task 1"}, OwnerName: &ownerName},
	}

	svc.On("GetTasksWithOwnerDetails", mock.Anything, mock.Anything).Return(tasksWithOwner, int64(1), nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks?withOwner=true", nil, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.ListTasks(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListTasksDetailResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Len(t, response.Tasks, 1)
	assert.Equal(t, "John Doe", *response.Tasks[0].OwnerName)
	svc.AssertExpectations(t)
}

func TestHandler_CreateTask_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateTaskRequest{
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Check vitals",
	}

	task := &Task{
		ID:        "task-1",
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Check vitals",
		Status:    TaskStatusOpen,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}

	svc.On("CreateTask", mock.Anything, mock.Anything, reqBody, "user-1").Return(task, nil)

	r := testutil.NewRequest(http.MethodPost, "/api/v1/tasks", reqBody, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.CreateTask(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response Task
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "task-1", response.ID)
	assert.Equal(t, "Check vitals", response.Title)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateTask_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	newTitle := "Updated title"
	reqBody := UpdateTaskRequest{
		Title: &newTitle,
	}

	task := &Task{
		ID:    "task-1",
		Title: "Updated title",
	}

	svc.On("UpdateTask", mock.Anything, mock.Anything, "task-1", reqBody, "user-1").Return(task, nil)

	r := testutil.NewRequest(http.MethodPatch, "/api/v1/tasks/task-1", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response Task
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "Updated title", response.Title)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateTask_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	newTitle := "Updated title"
	reqBody := UpdateTaskRequest{
		Title: &newTitle,
	}

	svc.On("UpdateTask", mock.Anything, mock.Anything, "task-999", reqBody, "user-1").
		Return(nil, fmt.Errorf("task not found"))

	r := testutil.NewRequest(http.MethodPatch, "/api/v1/tasks/task-999", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-999")

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, r)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_AssignTask_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	toOwner := "user-2"
	reqBody := AssignTaskRequest{
		ToOwnerID: &toOwner,
	}

	task := &Task{
		ID:             "task-1",
		Title:          "Task",
		CurrentOwnerID: &toOwner,
	}

	svc.On("AssignTask", mock.Anything, mock.Anything, "task-1", reqBody, "user-1", models.RoleChargeNurse).
		Return(task, nil)

	r := testutil.NewRequest(http.MethodPost, "/api/v1/tasks/task-1/assign", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.AssignTask(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response Task
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "user-2", *response.CurrentOwnerID)
	svc.AssertExpectations(t)
}

func TestHandler_AssignTask_Forbidden(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	toOwner := "user-2"
	reqBody := AssignTaskRequest{
		ToOwnerID: &toOwner,
	}

	svc.On("AssignTask", mock.Anything, mock.Anything, "task-1", reqBody, "user-1", models.RoleNurse).
		Return(nil, fmt.Errorf("insufficient permissions to reassign task owned by another user"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/tasks/task-1/assign", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.AssignTask(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_CompleteTask_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	notes := "Task completed"
	reqBody := CompleteTaskRequest{
		CompletionNotes: &notes,
	}

	completedBy := "user-1"
	completedAt := time.Now()
	task := &Task{
		ID:          "task-1",
		Title:       "Task",
		Status:      TaskStatusCompleted,
		CompletedBy: &completedBy,
		CompletedAt: &completedAt,
	}

	svc.On("CompleteTask", mock.Anything, mock.Anything, "task-1", reqBody, "user-1", models.RoleNurse).
		Return(task, nil)

	r := testutil.NewRequest(http.MethodPost, "/api/v1/tasks/task-1/complete", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.CompleteTask(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response Task
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, TaskStatusCompleted, response.Status)
	assert.NotNil(t, response.CompletedBy)
	svc.AssertExpectations(t)
}

func TestHandler_CompleteTask_Forbidden(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CompleteTaskRequest{}

	svc.On("CompleteTask", mock.Anything, mock.Anything, "task-1", reqBody, "user-1", models.RoleNurse).
		Return(nil, fmt.Errorf("cannot complete task assigned to another user"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/tasks/task-1/complete", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.CompleteTask(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetTaskHistory_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	events := []TaskAssignmentEvent{
		{ID: "e1", TaskID: "task-1"},
		{ID: "e2", TaskID: "task-1"},
	}

	svc.On("GetTaskHistory", mock.Anything, "task-1", 30, 0).Return(events, int64(2), nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks/task-1/history", nil, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.GetTaskHistory(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response TaskHistoryResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "task-1", response.TaskID)
	assert.Len(t, response.Events, 2)
	assert.Equal(t, int64(2), response.Total)
	svc.AssertExpectations(t)
}

func TestHandler_GetTask_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	task := &Task{
		ID:    "task-1",
		Title: "Task 1",
	}

	svc.On("GetTaskByID", mock.Anything, "task-1").Return(task, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks/task-1", nil, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.GetTask(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response Task
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "task-1", response.ID)
	svc.AssertExpectations(t)
}

func TestHandler_GetTask_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetTaskByID", mock.Anything, "task-999").Return(nil, fmt.Errorf("task not found"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks/task-999", nil, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-999")

	rr := httptest.NewRecorder()
	handler.GetTask(rr, r)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

// Additional test cases for improved coverage

func TestHandler_CreateTask_MissingRequiredFields(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	// Missing ScopeID
	reqBody := CreateTaskRequest{
		ScopeType: ScopeTypeEncounter,
		Title:     "Check vitals",
	}

	svc.On("CreateTask", mock.Anything, mock.Anything, reqBody, "user-1").
		Return(nil, fmt.Errorf("scopeId is required"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/tasks", reqBody, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.CreateTask(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateTask_MissingFields(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	// Empty update request (no fields to update)
	reqBody := UpdateTaskRequest{}

	svc.On("UpdateTask", mock.Anything, mock.Anything, "task-1", reqBody, "user-1").
		Return(nil, fmt.Errorf("no fields to update"))

	r := testutil.NewRequest(http.MethodPatch, "/api/v1/tasks/task-1", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.UpdateTask(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_AssignTask_RoleNotAllowed(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	toOwner := "user-2"
	reqBody := AssignTaskRequest{
		ToOwnerID: &toOwner,
	}

	svc.On("AssignTask", mock.Anything, mock.Anything, "task-1", reqBody, "user-1", models.RoleNurse).
		Return(nil, fmt.Errorf("insufficient permissions to reassign task owned by another user"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/tasks/task-1/assign", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.AssignTask(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_CompleteTask_NotInCompletableState(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CompleteTaskRequest{}

	svc.On("CompleteTask", mock.Anything, mock.Anything, "task-1", reqBody, "user-1", models.RoleNurse).
		Return(nil, fmt.Errorf("task is already completed"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/tasks/task-1/complete", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.CompleteTask(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetTaskHistory_WithPagination(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	events := []TaskAssignmentEvent{
		{ID: "e1", TaskID: "task-1"},
	}

	svc.On("GetTaskHistory", mock.Anything, "task-1", 10, 5).Return(events, int64(20), nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks/task-1/history?limit=10&offset=5", nil, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-1")

	rr := httptest.NewRecorder()
	handler.GetTaskHistory(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response TaskHistoryResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "task-1", response.TaskID)
	assert.Len(t, response.Events, 1)
	assert.Equal(t, int64(20), response.Total)
	svc.AssertExpectations(t)
}

func TestHandler_GetTaskHistory_TaskNotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetTaskHistory", mock.Anything, "task-999", 30, 0).
		Return(nil, int64(0), fmt.Errorf("task not found"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks/task-999/history", nil, "user-1", models.RoleNurse)
	r.SetPathValue("id", "task-999")

	rr := httptest.NewRecorder()
	handler.GetTaskHistory(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_ListTasks_WithOwnerDetails_Additional(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	tasks := []TaskWithOwner{
		{
			Task: Task{
				ID:        "task-1",
				ScopeType: ScopeTypeEncounter,
				ScopeID:   "enc-1",
				Title:     "Test Task",
				Status:    TaskStatusOpen,
				Priority:  TaskPriorityMedium,
				CreatedBy: "user-1",
			},
			OwnerName:  stringPtr("John Doe"),
			OwnerEmail: stringPtr("john@example.com"),
		},
	}

	svc.On("GetTasksWithOwnerDetails", mock.Anything, mock.MatchedBy(func(f ListTasksFilter) bool {
		return f.Limit == 30 && f.Offset == 0
	})).Return(tasks, int64(1), nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/tasks?withOwner=true", nil, "user-1", models.RoleNurse)
	rr := httptest.NewRecorder()
	handler.ListTasks(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListTasksDetailResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Len(t, response.Tasks, 1)
	assert.Equal(t, "John Doe", *response.Tasks[0].OwnerName)
	svc.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}
