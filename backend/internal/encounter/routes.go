package encounter

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
)

// RegisterRoutes registers encounter routes
func RegisterRoutes(mux *http.ServeMux, h *Handler, jwtService *auth.JWTService) {
	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/encounters - list encounters
	mux.Handle("GET /api/v1/encounters", authMiddleware(http.HandlerFunc(h.List)))

	// POST /api/v1/encounters - create encounter
	mux.Handle("POST /api/v1/encounters", authMiddleware(http.HandlerFunc(h.Create)))

	// GET /api/v1/encounters/{encounterId} - get encounter by ID
	mux.Handle("GET /api/v1/encounters/{encounterId}", authMiddleware(http.HandlerFunc(h.GetByID)))

	// PATCH /api/v1/encounters/{encounterId} - update encounter
	mux.Handle("PATCH /api/v1/encounters/{encounterId}", authMiddleware(http.HandlerFunc(h.Update)))
}
