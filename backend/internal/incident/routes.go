package incident

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
)

// RegisterRoutes registers incident routes
func RegisterRoutes(mux *http.ServeMux, h *Handler, jwtService auth.TokenService) {
	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/incidents - list incidents
	mux.Handle("GET /api/v1/incidents", authMiddleware(http.HandlerFunc(h.List)))

	// POST /api/v1/incidents - create incident
	mux.Handle("POST /api/v1/incidents", authMiddleware(http.HandlerFunc(h.Create)))

	// GET /api/v1/incidents/{incidentId} - get incident by ID
	mux.Handle("GET /api/v1/incidents/{incidentId}", authMiddleware(http.HandlerFunc(h.GetByID)))

	// POST /api/v1/incidents/{incidentId}/status - update incident status
	mux.Handle("POST /api/v1/incidents/{incidentId}/status", authMiddleware(http.HandlerFunc(h.UpdateStatus)))

	// GET /api/v1/incidents/{incidentId}/status-history - get incident status history
	mux.Handle("GET /api/v1/incidents/{incidentId}/status-history", authMiddleware(http.HandlerFunc(h.GetStatusHistory)))
}
