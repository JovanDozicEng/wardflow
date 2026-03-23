package router

import (
	"net/http"

	"github.com/wardflow/backend/internal/handler"
	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// Router sets up HTTP routes
type Router struct {
	mux         *http.ServeMux
	authHandler *handler.AuthHandler
	jwtService  *auth.JWTService
}

// New creates a new router
func New(db *database.DB, jwtService *auth.JWTService, authService *auth.Service) *Router {
	r := &Router{
		mux:         http.NewServeMux(),
		authHandler: handler.NewAuthHandler(authService),
		jwtService:  jwtService,
	}

	r.setupRoutes(db)
	return r
}

// ServeHTTP implements http.Handler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func (r *Router) setupRoutes(db *database.DB) {
	// Public routes (no auth required)
	r.mux.HandleFunc("/health", healthHandler(db))
	r.mux.HandleFunc("/auth/register", r.authHandler.Register)
	r.mux.HandleFunc("/auth/login", r.authHandler.Login)

	// Protected routes (auth required)
	r.mux.Handle("/auth/logout", 
		middleware.AuthMiddleware(r.jwtService)(
			http.HandlerFunc(r.authHandler.Logout)))
	
	r.mux.Handle("/auth/me", 
		middleware.AuthMiddleware(r.jwtService)(
			http.HandlerFunc(r.authHandler.Me)))
	
	r.mux.Handle("/auth/change-password", 
		middleware.AuthMiddleware(r.jwtService)(
			http.HandlerFunc(r.authHandler.ChangePassword)))

	// Future protected endpoints will be added here
	// Example with role-based access:
	// r.mux.Handle("/api/admin/users",
	//     middleware.AuthMiddleware(r.jwtService)(
	//         middleware.RequireRole(models.RoleAdmin)(
	//             http.HandlerFunc(adminHandler.ListUsers))))
}

func healthHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Check database health
		dbHealth, err := db.HealthCheck(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","database":"error"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		response := `{"status":"ok","database":"` + dbHealth["status"].(string) + `"}`
		w.Write([]byte(response))
	}
}
