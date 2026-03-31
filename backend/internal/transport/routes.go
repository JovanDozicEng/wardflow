package transport

import (
	"net/http"

	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// RegisterRoutes registers all transport routes
func RegisterRoutes(mux *http.ServeMux, db *database.DB, jwtService auth.TokenService) {
	// Wire dependencies
	repo := NewRepository(db)
	svc := NewService(repo, db)
	h := NewHandler(svc, db)

	// All routes require authentication
	authMiddleware := middleware.AuthMiddleware(jwtService)

	// GET /api/v1/transport/requests - list transport requests
	mux.Handle("GET /api/v1/transport/requests", authMiddleware(http.HandlerFunc(h.ListRequests)))

	// POST /api/v1/transport/requests - create transport request
	mux.Handle("POST /api/v1/transport/requests", authMiddleware(http.HandlerFunc(h.CreateRequest)))

	// POST /api/v1/transport/requests/{requestId}/accept - accept transport request
	mux.Handle("POST /api/v1/transport/requests/{requestId}/accept", authMiddleware(http.HandlerFunc(h.AcceptRequest)))

	// PATCH /api/v1/transport/requests/{requestId} - update transport request
	mux.Handle("PATCH /api/v1/transport/requests/{requestId}", authMiddleware(http.HandlerFunc(h.UpdateRequest)))

	// POST /api/v1/transport/requests/{requestId}/complete - complete transport request
	mux.Handle("POST /api/v1/transport/requests/{requestId}/complete", authMiddleware(http.HandlerFunc(h.CompleteRequest)))
}
