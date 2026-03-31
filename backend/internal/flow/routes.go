package flow

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// RegisterRoutes registers flow tracking routes
func RegisterRoutes(mux *http.ServeMux, db *database.DB, jwtService auth.TokenService) {
	// Wire dependencies: repo -> service -> handler
	repo := NewRepository(db)
	service := NewService(repo, db)
	handler := NewHandler(service, db)

	// Apply auth middleware to all routes
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/encounters/{encounterId}/flow - Get flow timeline
	mux.Handle("GET /api/v1/encounters/{encounterId}/flow",
		authMiddleware(http.HandlerFunc(handler.GetFlowTimeline)))

	// GET /api/v1/encounters/{encounterId}/flow/current - Get current state
	mux.Handle("GET /api/v1/encounters/{encounterId}/flow/current",
		authMiddleware(http.HandlerFunc(handler.GetCurrentState)))

	// POST /api/v1/encounters/{encounterId}/flow/transitions - Record a transition
	mux.Handle("POST /api/v1/encounters/{encounterId}/flow/transitions",
		authMiddleware(http.HandlerFunc(handler.RecordTransition)))

	// POST /api/v1/encounters/{encounterId}/flow/override - Override transition (privileged)
	mux.Handle("POST /api/v1/encounters/{encounterId}/flow/override",
		authMiddleware(http.HandlerFunc(handler.OverrideTransition)))
}
