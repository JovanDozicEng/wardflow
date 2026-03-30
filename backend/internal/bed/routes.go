package bed

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
)

// RegisterRoutes registers all bed routes
func RegisterRoutes(mux *http.ServeMux, h *Handler, jwtService *auth.JWTService) {
	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/beds - list beds
	mux.Handle("GET /api/v1/beds", authMiddleware(http.HandlerFunc(h.ListBeds)))

	// POST /api/v1/beds - create bed
	mux.Handle("POST /api/v1/beds", authMiddleware(http.HandlerFunc(h.CreateBed)))

	// GET /api/v1/beds/{bedId} - get bed by ID
	mux.Handle("GET /api/v1/beds/{bedId}", authMiddleware(http.HandlerFunc(h.GetBed)))

	// POST /api/v1/beds/{bedId}/status - update bed status
	mux.Handle("POST /api/v1/beds/{bedId}/status", authMiddleware(http.HandlerFunc(h.UpdateBedStatus)))

	// POST /api/v1/encounters/{encounterId}/bed-requests - create bed request
	mux.Handle("POST /api/v1/encounters/{encounterId}/bed-requests", authMiddleware(http.HandlerFunc(h.CreateBedRequest)))

	// POST /api/v1/bed-requests/{requestId}/assign - assign bed to request
	mux.Handle("POST /api/v1/bed-requests/{requestId}/assign", authMiddleware(http.HandlerFunc(h.AssignBed)))
}
