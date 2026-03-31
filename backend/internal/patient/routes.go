package patient

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
)

// RegisterRoutes registers all patient routes
func RegisterRoutes(mux *http.ServeMux, h *Handler, jwtService auth.TokenService) {
	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/patients - list patients
	mux.Handle("GET /api/v1/patients", authMiddleware(http.HandlerFunc(h.List)))

	// POST /api/v1/patients - create patient
	mux.Handle("POST /api/v1/patients", authMiddleware(http.HandlerFunc(h.Create)))

	// GET /api/v1/patients/{patientId} - get patient by ID
	mux.Handle("GET /api/v1/patients/{patientId}", authMiddleware(http.HandlerFunc(h.Get)))
}
