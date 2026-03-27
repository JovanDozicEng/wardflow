package middleware

import (
	"net/http"
	"strings"

	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/logger"
)

// AuthMiddleware verifies JWT tokens and adds user context
func AuthMiddleware(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				logger.Warn("token validation failed: %v", err)
				respondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			// Add user context to request
			ctx := auth.SetUserContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := auth.GetUserContext(r.Context())
			if !ok {
				respondError(w, http.StatusUnauthorized, "user context not found")
				return
			}

			// Admin has access to everything
			if userCtx.Role == models.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				if userCtx.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireUnitAccess middleware checks if user can access a unit
func RequireUnitAccess(getUnitID func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := auth.GetUserContext(r.Context())
			if !ok {
				respondError(w, http.StatusUnauthorized, "user context not found")
				return
			}

			// Admin has access to all units
			if userCtx.Role == models.RoleAdmin {
				next.ServeHTTP(w, r)
				return
			}

			unitID := getUnitID(r)
			if unitID == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has access to this unit
			hasAccess := false
			for _, id := range userCtx.UnitIDs {
				if id == unitID {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				respondError(w, http.StatusForbidden, "no access to this unit")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth adds user context if token is present, but doesn't require it
func OptionalAuth(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := jwtService.ValidateToken(parts[1])
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := auth.SetUserContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuditLogger middleware logs authenticated actions with user context
func AuditLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCtx, ok := auth.GetUserContext(r.Context())
		if ok {
			logger.Info("[AUDIT] user=%s role=%s action=%s %s", 
				userCtx.UserID, userCtx.Role, r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
