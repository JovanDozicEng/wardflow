package router

import (
"net/http"

"github.com/wardflow/backend/internal/careteam"
"github.com/wardflow/backend/internal/consult"
"github.com/wardflow/backend/internal/dashboard"
"github.com/wardflow/backend/internal/department"
"github.com/wardflow/backend/internal/encounter"
"github.com/wardflow/backend/internal/exception"
"github.com/wardflow/backend/internal/flow"
"github.com/wardflow/backend/internal/handler"
"github.com/wardflow/backend/internal/incident"
"github.com/wardflow/backend/internal/middleware"
"github.com/wardflow/backend/internal/patient"
"github.com/wardflow/backend/internal/task"
"github.com/wardflow/backend/internal/unit"
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
allowedOrigins: []string{
"http://localhost:5173",
"http://localhost:5174",
"http://localhost:5175",
"http://localhost:5176",
"http://localhost:3000",
},
}

r.setupRoutes(db)

// Wrap mux with CorrelationID middleware, then CORS middleware (outermost layer)
h := middleware.CorrelationID(r.mux)
r.handler = middleware.CORSMiddleware(r.allowedOrigins)(h)

return r
}

// ServeHTTP implements http.Handler with CORS middleware
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
r.handler.ServeHTTP(w, req)
}

func (r *Router) setupRoutes(db *database.DB) {
// System routes (no auth required)
r.mux.HandleFunc("GET /health", healthHandler(db))
r.mux.HandleFunc("GET /readyz", readyHandler(db))

// Auth routes — all under /api/v1
r.mux.HandleFunc("POST /api/v1/auth/register", r.authHandler.Register)
r.mux.HandleFunc("POST /api/v1/auth/login", r.authHandler.Login)

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

// Patient routes
patientRepo := patient.NewRepository(db)
patientService := patient.NewService(patientRepo)
patientHandler := patient.NewHandler(patientService, db)
patient.RegisterRoutes(r.mux, patientHandler, r.jwtService)

// Reference data routes - Departments and Units
departmentHandler := department.NewHandler(db)
department.RegisterRoutes(r.mux, departmentHandler, r.jwtService)

unitHandler := unit.NewHandler(db)
unit.RegisterRoutes(r.mux, unitHandler, r.jwtService)

// Clinical Core routes (Developer A)
careteam.RegisterRoutes(r.mux, db, r.jwtService)
flow.RegisterRoutes(r.mux, db, r.jwtService)
task.RegisterRoutes(r.mux, db, r.jwtService)
dashboard.RegisterRoutes(r.mux, db, r.jwtService)

// Governance & Safety routes (Developer C)
consultRepo := consult.NewRepository(db)
consultService := consult.NewService(consultRepo)
consultHandler := consult.NewHandler(consultService, db)
consult.RegisterRoutes(r.mux, consultHandler, r.jwtService)

exceptionRepo := exception.NewRepository(db)
exceptionService := exception.NewService(exceptionRepo)
exceptionHandler := exception.NewHandler(exceptionService, db)
exception.RegisterRoutes(r.mux, exceptionHandler, r.jwtService)

incidentRepo := incident.NewRepository(db)
incidentService := incident.NewService(incidentRepo)
incidentHandler := incident.NewHandler(incidentService, db)
incident.RegisterRoutes(r.mux, incidentHandler, r.jwtService)

// Users — searchable list for care team assignment
usersHandler := handler.NewUsersHandler(db)
r.mux.Handle("GET /api/v1/users",
	middleware.AuthMiddleware(r.jwtService)(http.HandlerFunc(usersHandler.ListUsers)),
)
}

func healthHandler(db *database.DB) http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

w.Header().Set("Content-Type", "application/json")

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
