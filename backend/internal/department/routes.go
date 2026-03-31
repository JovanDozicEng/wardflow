package department

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// RegisterRoutes registers all department routes
func RegisterRoutes(mux *http.ServeMux, db *database.DB, jwtService auth.TokenService) {
	// Wire dependencies
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/departments - list departments
	mux.Handle("GET /api/v1/departments", authMiddleware(http.HandlerFunc(h.List)))

	// POST /api/v1/departments - create department (admin only, checked in handler)
	mux.Handle("POST /api/v1/departments", authMiddleware(http.HandlerFunc(h.Create)))

	// GET /api/v1/departments/{departmentId} - get department by ID
	mux.Handle("GET /api/v1/departments/{departmentId}", authMiddleware(http.HandlerFunc(h.Get)))
}
