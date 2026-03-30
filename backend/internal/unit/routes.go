package unit

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
)

// RegisterRoutes registers all unit routes
func RegisterRoutes(mux *http.ServeMux, h *Handler, jwtService *auth.JWTService) {
	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/units - list units
	mux.Handle("GET /api/v1/units", authMiddleware(http.HandlerFunc(h.List)))

	// POST /api/v1/units - create unit (admin only, checked in handler)
	mux.Handle("POST /api/v1/units", authMiddleware(http.HandlerFunc(h.Create)))

	// GET /api/v1/units/{unitId} - get unit by ID
	mux.Handle("GET /api/v1/units/{unitId}", authMiddleware(http.HandlerFunc(h.Get)))
}
