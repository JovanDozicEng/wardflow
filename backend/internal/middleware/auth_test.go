package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
)

// Mock TokenService
type mockTokenService struct {
	mock.Mock
}

func (m *mockTokenService) GenerateToken(user *models.User) (string, int64, error) {
	args := m.Called(user)
	return args.String(0), args.Get(1).(int64), args.Error(2)
}

func (m *mockTokenService) ValidateToken(tokenString string) (*auth.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *mockTokenService) RefreshToken(tokenString string) (string, int64, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Get(1).(int64), args.Error(2)
}

// Test handler that checks user context
func createTestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userCtx, ok := auth.GetUserContext(r.Context())
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(userCtx.UserID))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("no-user"))
		}
	}
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("allows request with valid token", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := AuthMiddleware(tokenSvc)

		claims := &auth.Claims{
			UserID: "user-123",
			Email:  "test@example.com",
			Role:   models.RoleNurse,
		}

		tokenSvc.On("ValidateToken", "valid-token").Return(claims, nil)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "user-123", rr.Body.String())

		tokenSvc.AssertExpectations(t)
	})

	t.Run("returns 401 when Authorization header is missing", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := AuthMiddleware(tokenSvc)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		// No Authorization header
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns 401 for invalid Authorization header format", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := AuthMiddleware(tokenSvc)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "InvalidFormat token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns 401 for missing Bearer prefix", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := AuthMiddleware(tokenSvc)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "some-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns 401 for invalid token", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := AuthMiddleware(tokenSvc)

		tokenSvc.On("ValidateToken", "invalid-token").Return(nil, auth.ErrInvalidToken)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		tokenSvc.AssertExpectations(t)
	})

	t.Run("returns 401 for expired token", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := AuthMiddleware(tokenSvc)

		tokenSvc.On("ValidateToken", "expired-token").Return(nil, auth.ErrExpiredToken)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "Bearer expired-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		tokenSvc.AssertExpectations(t)
	})

	t.Run("sets user context correctly", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := AuthMiddleware(tokenSvc)

		claims := &auth.Claims{
			UserID:  "user-456",
			Email:   "admin@example.com",
			Role:    models.RoleAdmin,
			UnitIDs: models.StringArray{"unit-1"},
			DeptIDs: models.StringArray{"dept-1"},
		}

		tokenSvc.On("ValidateToken", "valid-token").Return(claims, nil)

		// Custom handler to verify context
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := auth.GetUserContext(r.Context())
			assert.True(t, ok)
			assert.Equal(t, "user-456", userCtx.UserID)
			assert.Equal(t, "admin@example.com", userCtx.Email)
			assert.Equal(t, models.RoleAdmin, userCtx.Role)
			assert.Len(t, userCtx.UnitIDs, 1)
			assert.Len(t, userCtx.DeptIDs, 1)
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		tokenSvc.AssertExpectations(t)
	})
}

func TestOptionalAuth(t *testing.T) {
	t.Run("sets user context when valid token provided", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := OptionalAuth(tokenSvc)

		claims := &auth.Claims{
			UserID: "user-123",
			Email:  "test@example.com",
			Role:   models.RoleNurse,
		}

		tokenSvc.On("ValidateToken", "valid-token").Return(claims, nil)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "user-123", rr.Body.String())

		tokenSvc.AssertExpectations(t)
	})

	t.Run("allows request without Authorization header", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := OptionalAuth(tokenSvc)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		// No Authorization header
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "no-user", rr.Body.String())
	})

	t.Run("allows request with invalid token format", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := OptionalAuth(tokenSvc)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "InvalidFormat")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "no-user", rr.Body.String())
	})

	t.Run("allows request when token validation fails", func(t *testing.T) {
		tokenSvc := new(mockTokenService)
		middleware := OptionalAuth(tokenSvc)

		tokenSvc.On("ValidateToken", "invalid-token").Return(nil, auth.ErrInvalidToken)

		handler := middleware(createTestHandler())

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		r.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "no-user", rr.Body.String())

		tokenSvc.AssertExpectations(t)
	})
}

func TestRequireRole(t *testing.T) {
	t.Run("allows admin access to any route", func(t *testing.T) {
		middleware := RequireRole(models.RoleNurse, models.RoleProvider)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := auth.SetUserContext(r.Context(), &auth.Claims{
			UserID: "admin-1",
			Role:   models.RoleAdmin,
		})
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("allows user with required role", func(t *testing.T) {
		middleware := RequireRole(models.RoleNurse)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := auth.SetUserContext(r.Context(), &auth.Claims{
			UserID: "user-1",
			Role:   models.RoleNurse,
		})
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("allows user with one of multiple required roles", func(t *testing.T) {
		middleware := RequireRole(models.RoleNurse, models.RoleProvider, models.RoleChargeNurse)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := auth.SetUserContext(r.Context(), &auth.Claims{
			UserID: "user-1",
			Role:   models.RoleProvider,
		})
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("returns forbidden when user lacks required role", func(t *testing.T) {
		middleware := RequireRole(models.RoleProvider, models.RoleAdmin)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		ctx := auth.SetUserContext(r.Context(), &auth.Claims{
			UserID: "user-1",
			Role:   models.RoleNurse,
		})
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("returns unauthorized when user context not found", func(t *testing.T) {
		middleware := RequireRole(models.RoleNurse)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		// No user context
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestRequireUnitAccess(t *testing.T) {
	getUnitID := func(r *http.Request) string {
		return r.PathValue("unitId")
	}

	t.Run("allows admin access to any unit", func(t *testing.T) {
		middleware := RequireUnitAccess(getUnitID)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/units/unit-1", nil)
		r.SetPathValue("unitId", "unit-1")
		ctx := auth.SetUserContext(r.Context(), &auth.Claims{
			UserID: "admin-1",
			Role:   models.RoleAdmin,
		})
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("allows user with unit access", func(t *testing.T) {
		middleware := RequireUnitAccess(getUnitID)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/units/unit-1", nil)
		r.SetPathValue("unitId", "unit-1")
		ctx := auth.SetUserContext(r.Context(), &auth.Claims{
			UserID:  "user-1",
			Role:    models.RoleNurse,
			UnitIDs: models.StringArray{"unit-1", "unit-2"},
		})
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("returns forbidden when user lacks unit access", func(t *testing.T) {
		middleware := RequireUnitAccess(getUnitID)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/units/unit-3", nil)
		r.SetPathValue("unitId", "unit-3")
		ctx := auth.SetUserContext(r.Context(), &auth.Claims{
			UserID:  "user-1",
			Role:    models.RoleNurse,
			UnitIDs: models.StringArray{"unit-1", "unit-2"},
		})
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("allows request when unitId is empty", func(t *testing.T) {
		middleware := RequireUnitAccess(getUnitID)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		// No unitId path value
		ctx := auth.SetUserContext(r.Context(), &auth.Claims{
			UserID:  "user-1",
			Role:    models.RoleNurse,
			UnitIDs: models.StringArray{},
		})
		r = r.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("returns unauthorized when user context not found", func(t *testing.T) {
		middleware := RequireUnitAccess(getUnitID)

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/api/units/unit-1", nil)
		r.SetPathValue("unitId", "unit-1")
		// No user context
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
