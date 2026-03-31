package careteam

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// RegisterRoutes registers care team routes
func RegisterRoutes(mux *http.ServeMux, db *database.DB, jwtService auth.TokenService) {
	// Wire dependencies: repo -> service -> handler
	repo := NewRepository(db)
	service := NewService(repo, db)
	handler := NewHandler(service, db)

	// Apply auth middleware to all routes
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/encounters/{encounterId}/care-team/assignments - Get care team assignments
	mux.Handle("GET /api/v1/encounters/{encounterId}/care-team/assignments",
		authMiddleware(http.HandlerFunc(handler.GetCareTeam)))

	// POST /api/v1/encounters/{encounterId}/care-team/assignments - Assign a role
	mux.Handle("POST /api/v1/encounters/{encounterId}/care-team/assignments",
		authMiddleware(http.HandlerFunc(handler.AssignRole)))

	// POST /api/v1/care-team/assignments/{assignmentId}/transfer - Transfer a role
	mux.Handle("POST /api/v1/care-team/assignments/{assignmentId}/transfer",
		authMiddleware(http.HandlerFunc(handler.TransferRole)))

	// GET /api/v1/encounters/{encounterId}/handoffs - Get handoff notes
	mux.Handle("GET /api/v1/encounters/{encounterId}/handoffs",
		authMiddleware(http.HandlerFunc(handler.GetHandoffs)))
}
