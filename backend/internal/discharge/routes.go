package discharge

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
)

// RegisterRoutes registers all discharge routes
func RegisterRoutes(mux *http.ServeMux, h *Handler, jwtService *auth.JWTService) {
	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// POST /api/v1/encounters/{encounterId}/discharge-checklist/init - initialize discharge checklist
	mux.Handle("POST /api/v1/encounters/{encounterId}/discharge-checklist/init", authMiddleware(http.HandlerFunc(h.InitChecklist)))

	// GET /api/v1/encounters/{encounterId}/discharge-checklist - get discharge checklist
	mux.Handle("GET /api/v1/encounters/{encounterId}/discharge-checklist", authMiddleware(http.HandlerFunc(h.GetChecklist)))

	// POST /api/v1/discharge-checklist/items/{itemId}/complete - complete checklist item
	mux.Handle("POST /api/v1/discharge-checklist/items/{itemId}/complete", authMiddleware(http.HandlerFunc(h.CompleteItem)))

	// POST /api/v1/encounters/{encounterId}/discharge/complete - complete discharge
	mux.Handle("POST /api/v1/encounters/{encounterId}/discharge/complete", authMiddleware(http.HandlerFunc(h.CompleteDischarge)))
}
