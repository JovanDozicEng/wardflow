package task

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// RegisterRoutes registers task routes
func RegisterRoutes(mux *http.ServeMux, db *database.DB, jwtService auth.TokenService) {
	// Wire dependencies: repo -> service -> handler
	repo := NewRepository(db)
	service := NewService(repo, db)
	handler := NewHandler(service, db)

	// Apply auth middleware to all routes
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/tasks - List tasks with filters
	mux.Handle("GET /api/v1/tasks",
		authMiddleware(http.HandlerFunc(handler.ListTasks)))

	// GET /api/v1/tasks/{id} - Get a single task
	mux.Handle("GET /api/v1/tasks/{id}",
		authMiddleware(http.HandlerFunc(handler.GetTask)))

	// POST /api/v1/tasks - Create a task
	mux.Handle("POST /api/v1/tasks",
		authMiddleware(http.HandlerFunc(handler.CreateTask)))

	// PATCH /api/v1/tasks/{id} - Update a task
	mux.Handle("PATCH /api/v1/tasks/{id}",
		authMiddleware(http.HandlerFunc(handler.UpdateTask)))

	// POST /api/v1/tasks/{id}/assign - Assign a task
	mux.Handle("POST /api/v1/tasks/{id}/assign",
		authMiddleware(http.HandlerFunc(handler.AssignTask)))

	// POST /api/v1/tasks/{id}/complete - Complete a task
	mux.Handle("POST /api/v1/tasks/{id}/complete",
		authMiddleware(http.HandlerFunc(handler.CompleteTask)))

	// GET /api/v1/tasks/{id}/history - Get assignment history
	mux.Handle("GET /api/v1/tasks/{id}/history",
		authMiddleware(http.HandlerFunc(handler.GetTaskHistory)))
}
