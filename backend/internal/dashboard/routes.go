package dashboard

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// RegisterRoutes registers dashboard routes
func RegisterRoutes(mux *http.ServeMux, db *database.DB, jwtService auth.TokenService) {
	// Wire dependencies: repo -> service -> handler
	repo := NewRepository(db)
	service := NewService(repo, db)
	handler := NewHandler(service)

	// Apply auth middleware to all routes
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/dashboard/huddle - Get huddle dashboard metrics
	mux.Handle("GET /api/v1/dashboard/huddle",
		authMiddleware(http.HandlerFunc(handler.GetHuddleDashboard)))
}
