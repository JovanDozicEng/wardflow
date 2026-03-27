package exception

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
)

// RegisterRoutes registers exception event routes
func RegisterRoutes(mux *http.ServeMux, h *Handler, jwtService *auth.JWTService) {
	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/exceptions - list exception events
	mux.Handle("GET /api/v1/exceptions", authMiddleware(http.HandlerFunc(h.List)))

	// POST /api/v1/exceptions - create exception event
	mux.Handle("POST /api/v1/exceptions", authMiddleware(http.HandlerFunc(h.Create)))

	// PATCH /api/v1/exceptions/{exceptionId} - update exception event
	mux.Handle("PATCH /api/v1/exceptions/{exceptionId}", authMiddleware(http.HandlerFunc(h.Update)))

	// POST /api/v1/exceptions/{exceptionId}/finalize - finalize exception event
	mux.Handle("POST /api/v1/exceptions/{exceptionId}/finalize", authMiddleware(http.HandlerFunc(h.Finalize)))

	// POST /api/v1/exceptions/{exceptionId}/correct - correct exception event
	mux.Handle("POST /api/v1/exceptions/{exceptionId}/correct", authMiddleware(http.HandlerFunc(h.Correct)))
}
