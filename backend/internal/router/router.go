package router

import (
	"net/http"

	"github.com/wardflow/backend/internal/consult"
	"github.com/wardflow/backend/internal/encounter"
	"github.com/wardflow/backend/internal/exception"
	"github.com/wardflow/backend/internal/handler"
	"github.com/wardflow/backend/internal/incident"
	"github.com/wardflow/backend/internal/middleware"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// Router sets up HTTP routes
type Router struct {
	handler        http.Handler // CORS-wrapped mux
	mux            *http.ServeMux
	authHandler    *handler.AuthHandler
	jwtService     *auth.JWTService
	allowedOrigins []string
}

// New creates a new router
func New(db *database.DB, jwtService *auth.JWTService, authService *auth.Service) *Router {
	mux := http.NewServeMux()
	
	r := &Router{
		mux:         mux,
		authHandler: handler.NewAuthHandler(authService),
		jwtService:  jwtService,
		// Allow frontend origins (add more as needed)
		allowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:5175",
			"http://localhost:5176",
			"http://localhost:3000", // Common React dev port
		},
	}

	r.setupRoutes(db)
	
	// Wrap mux with CorrelationID middleware, then CORS middleware (outermost layer)
	handler := middleware.CorrelationID(r.mux)
	r.handler = middleware.CORSMiddleware(r.allowedOrigins)(handler)
	
	return r
}

// ServeHTTP implements http.Handler with CORS middleware
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (r *Router) setupRoutes(db *database.DB) {
	// Public routes (no auth required) - health checks at root level
	r.mux.HandleFunc("GET /health", healthHandler(db))
	r.mux.HandleFunc("GET /readyz", readyHandler(db))
	
	// Auth routes - now with /api/v1 prefix for consistency
	r.mux.HandleFunc("POST /api/v1/auth/register", r.authHandler.Register)
	r.mux.HandleFunc("POST /api/v1/auth/login", r.authHandler.Login)

	// Protected auth routes
	r.mux.Handle("POST /api/v1/auth/logout", 
		middleware.AuthMiddleware(r.jwtService)(
			http.HandlerFunc(r.authHandler.Logout)))
	
	r.mux.Handle("GET /api/v1/auth/me", 
		middleware.AuthMiddleware(r.jwtService)(
			http.HandlerFunc(r.authHandler.Me)))
	
	r.mux.Handle("POST /api/v1/auth/change-password", 
		middleware.AuthMiddleware(r.jwtService)(
			http.HandlerFunc(r.authHandler.ChangePassword)))

	// Encounter routes
	encounterRepo := encounter.NewRepository(db)
	encounterService := encounter.NewService(encounterRepo)
	encounterHandler := encounter.NewHandler(encounterService, db)
	encounter.RegisterRoutes(r.mux, encounterHandler, r.jwtService)

	// Consult routes
	consultRepo := consult.NewRepository(db)
	consultService := consult.NewService(consultRepo)
	consultHandler := consult.NewHandler(consultService, db)
	consult.RegisterRoutes(r.mux, consultHandler, r.jwtService)

	// Exception routes
	exceptionRepo := exception.NewRepository(db)
	exceptionService := exception.NewService(exceptionRepo)
	exceptionHandler := exception.NewHandler(exceptionService, db)
	exception.RegisterRoutes(r.mux, exceptionHandler, r.jwtService)

	// Incident routes
	incidentRepo := incident.NewRepository(db)
	incidentService := incident.NewService(incidentRepo)
	incidentHandler := incident.NewHandler(incidentService, db)
	incident.RegisterRoutes(r.mux, incidentHandler, r.jwtService)

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

func readyHandler(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check database health
		_, err := db.HealthCheck(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	}
}
