package task

import (
	"net/http"
	"strconv"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// Handler handles HTTP requests for task management
type Handler struct {
	service Service
	db      *database.DB
}

// NewHandler creates a new task handler
func NewHandler(service Service, db *database.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// ListTasks returns tasks with filters
// GET /api/v1/tasks
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query params
	filter := ListTasksFilter{
		Limit:  30,
		Offset: 0,
	}

	if scopeType := r.URL.Query().Get("scopeType"); scopeType != "" {
		st := ScopeType(scopeType)
		filter.ScopeType = &st
	}
	if scopeID := r.URL.Query().Get("scopeId"); scopeID != "" {
		filter.ScopeID = &scopeID
	}
	if status := r.URL.Query().Get("status"); status != "" {
		s := TaskStatus(status)
		filter.Status = &s
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		p := TaskPriority(priority)
		filter.Priority = &p
	}
	if ownerID := r.URL.Query().Get("ownerId"); ownerID != "" {
		filter.OwnerID = &ownerID
	}
	if overdue := r.URL.Query().Get("overdue"); overdue == "true" {
		t := true
		filter.Overdue = &t
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	// Check if detailed response with owner info is requested
	withOwner := r.URL.Query().Get("withOwner") == "true"

	if withOwner {
		tasks, total, err := h.service.GetTasksWithOwnerDetails(ctx, filter)
		if err != nil {
			httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}

		response := ListTasksDetailResponse{
			Tasks:  tasks,
			Total:  total,
			Limit:  filter.Limit,
			Offset: filter.Offset,
		}
		httputil.RespondJSON(w, http.StatusOK, response)
		return
	}

	tasks, total, err := h.service.ListTasks(ctx, filter)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response := ListTasksResponse{
		Tasks:  tasks,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}
	httputil.RespondJSON(w, http.StatusOK, response)
}

// CreateTask creates a new task
// POST /api/v1/tasks
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := auth.MustGetUserContext(ctx)

	var req CreateTaskRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	task, err := h.service.CreateTask(ctx, r, req, userCtx.UserID)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "CREATE_FAILED", err.Error())
		return
	}

	audit.Log(ctx, h.db, r, audit.Entry{
		EntityType: "task",
		EntityID:   task.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      task,
	})

	httputil.RespondJSON(w, http.StatusCreated, task)
}

// UpdateTask updates a task
// PATCH /api/v1/tasks/{id}
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := r.PathValue("id")
	userCtx := auth.MustGetUserContext(ctx)

	if taskID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "task id is required")
		return
	}

	var req UpdateTaskRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	task, err := h.service.UpdateTask(ctx, r, taskID, req, userCtx.UserID)
	if err != nil {
		if err.Error() == "task not found" {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httputil.RespondError(w, r, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	audit.Log(ctx, h.db, r, audit.Entry{
		EntityType: "task",
		EntityID:   taskID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		After:      task,
	})

	httputil.RespondJSON(w, http.StatusOK, task)
}

// AssignTask assigns a task to a user
// POST /api/v1/tasks/{id}/assign
func (h *Handler) AssignTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := r.PathValue("id")
	userCtx := auth.MustGetUserContext(ctx)

	if taskID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "task id is required")
		return
	}

	var req AssignTaskRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	task, err := h.service.AssignTask(ctx, r, taskID, req, userCtx.UserID, userCtx.Role)
	if err != nil {
		if err.Error() == "task not found" {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if err.Error() == "insufficient permissions to reassign task owned by another user" {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		httputil.RespondError(w, r, http.StatusBadRequest, "ASSIGN_FAILED", err.Error())
		return
	}

	audit.Log(ctx, h.db, r, audit.Entry{
		EntityType: "task",
		EntityID:   taskID,
		Action:     "ASSIGN",
		ByUserID:   userCtx.UserID,
		After:      task,
	})

	httputil.RespondJSON(w, http.StatusOK, task)
}

// CompleteTask marks a task as completed
// POST /api/v1/tasks/{id}/complete
func (h *Handler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := r.PathValue("id")
	userCtx := auth.MustGetUserContext(ctx)

	if taskID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "task id is required")
		return
	}

	var req CompleteTaskRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	task, err := h.service.CompleteTask(ctx, r, taskID, req, userCtx.UserID, userCtx.Role)
	if err != nil {
		if err.Error() == "task not found" {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if err.Error() == "cannot complete task assigned to another user" {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		httputil.RespondError(w, r, http.StatusBadRequest, "COMPLETE_FAILED", err.Error())
		return
	}

	audit.Log(ctx, h.db, r, audit.Entry{
		EntityType: "task",
		EntityID:   taskID,
		Action:     "COMPLETE",
		ByUserID:   userCtx.UserID,
		After:      task,
	})

	httputil.RespondJSON(w, http.StatusOK, task)
}

// GetTaskHistory returns assignment history for a task
// GET /api/v1/tasks/{id}/history
func (h *Handler) GetTaskHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := r.PathValue("id")

	if taskID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "task id is required")
		return
	}

	// Parse pagination
	limit := 30
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	events, total, err := h.service.GetTaskHistory(ctx, taskID, limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response := TaskHistoryResponse{
		TaskID: taskID,
		Events: events,
		Total:  total,
	}
	httputil.RespondJSON(w, http.StatusOK, response)
}

// GetTask returns a single task by ID
// GET /api/v1/tasks/{id}
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	taskID := r.PathValue("id")

	if taskID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "task id is required")
		return
	}

	task, err := h.service.GetTaskByID(ctx, taskID)
	if err != nil {
		if err.Error() == "task not found" {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusOK, task)
}
