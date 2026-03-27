package consult

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
)

// RegisterRoutes registers consult request routes
func RegisterRoutes(mux *http.ServeMux, h *Handler, jwtService *auth.JWTService) {
	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/consults - list consult requests
	mux.Handle("GET /api/v1/consults", authMiddleware(http.HandlerFunc(h.List)))

	// POST /api/v1/consults - create consult request
	mux.Handle("POST /api/v1/consults", authMiddleware(http.HandlerFunc(h.Create)))

	// POST /api/v1/consults/{consultId}/accept - accept consult request
	mux.Handle("POST /api/v1/consults/{consultId}/accept", authMiddleware(http.HandlerFunc(h.Accept)))

	// POST /api/v1/consults/{consultId}/decline - decline consult request
	mux.Handle("POST /api/v1/consults/{consultId}/decline", authMiddleware(http.HandlerFunc(h.Decline)))

	// POST /api/v1/consults/{consultId}/redirect - redirect consult request
	mux.Handle("POST /api/v1/consults/{consultId}/redirect", authMiddleware(http.HandlerFunc(h.Redirect)))

	// POST /api/v1/consults/{consultId}/complete - complete consult request
	mux.Handle("POST /api/v1/consults/{consultId}/complete", authMiddleware(http.HandlerFunc(h.Complete)))
}
