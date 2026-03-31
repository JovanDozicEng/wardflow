package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
	"github.com/wardflow/backend/pkg/auth"
)

// Mock AuthService
type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.LoginResponse), args.Error(1)
}

func (m *mockAuthService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockAuthService) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *mockAuthService) DeactivateUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthHandler_Register(t *testing.T) {
	t.Run("successfully registers new user", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		req := &models.RegisterRequest{
			Email:    "newuser@example.com",
			Password: "password123",
			Name:     "New User",
			Role:     models.RoleNurse,
		}

		user := &models.User{
			ID:       "user-123",
			Email:    req.Email,
			Name:     req.Name,
			Role:     req.Role,
			IsActive: true,
		}

		svc.On("Register", mock.Anything, req).Return(user, nil)

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/register", req)
		rr := httptest.NewRecorder()

		handler.Register(rr, r)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var response map[string]interface{}
		testutil.DecodeJSON(t, rr, &response)
		assert.Equal(t, "registration successful", response["message"])

		svc.AssertExpectations(t)
	})

	t.Run("returns conflict for duplicate email", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		req := &models.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "Test User",
			Role:     models.RoleNurse,
		}

		svc.On("Register", mock.Anything, req).Return(nil, auth.ErrEmailExists)

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/register", req)
		rr := httptest.NewRecorder()

		handler.Register(rr, r)

		assert.Equal(t, http.StatusConflict, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/register", nil)
		r.Body = http.NoBody
		rr := httptest.NewRecorder()

		handler.Register(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns bad request for missing required fields", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		req := &models.RegisterRequest{
			Email: "test@example.com",
			// Missing password and name
		}

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/register", req)
		rr := httptest.NewRecorder()

		handler.Register(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns bad request for short password", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		req := &models.RegisterRequest{
			Email:    "test@example.com",
			Password: "short",
			Name:     "Test User",
			Role:     models.RoleNurse,
		}

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/register", req)
		rr := httptest.NewRecorder()

		handler.Register(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("successfully authenticates user", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		response := &models.LoginResponse{
			Token:     "jwt.token.here",
			ExpiresAt: 1234567890,
			User: &models.UserInfo{
				ID:    "user-123",
				Email: req.Email,
				Role:  models.RoleNurse,
			},
		}

		svc.On("Login", mock.Anything, req).Return(response, nil)

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/login", req)
		rr := httptest.NewRecorder()

		handler.Login(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.LoginResponse
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, "jwt.token.here", result.Token)
		assert.Equal(t, "test@example.com", result.User.Email)

		svc.AssertExpectations(t)
	})

	t.Run("returns unauthorized for wrong password", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		svc.On("Login", mock.Anything, req).Return(nil, auth.ErrInvalidPassword)

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/login", req)
		rr := httptest.NewRecorder()

		handler.Login(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns unauthorized for non-existent user", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		req := &models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		svc.On("Login", mock.Anything, req).Return(nil, auth.ErrUserNotFound)

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/login", req)
		rr := httptest.NewRecorder()

		handler.Login(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns forbidden for inactive user", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		req := &models.LoginRequest{
			Email:    "inactive@example.com",
			Password: "password123",
		}

		svc.On("Login", mock.Anything, req).Return(nil, auth.ErrUserInactive)

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/login", req)
		rr := httptest.NewRecorder()

		handler.Login(rr, r)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request for missing credentials", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		req := &models.LoginRequest{
			Email: "test@example.com",
			// Missing password
		}

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/login", req)
		rr := httptest.NewRecorder()

		handler.Login(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns internal error on service failure", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		svc.On("Login", mock.Anything, req).Return(nil, errors.New("database error"))

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/auth/login", req)
		rr := httptest.NewRecorder()

		handler.Login(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestAuthHandler_Me(t *testing.T) {
	t.Run("returns current user info", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		userID := "user-123"
		user := &models.User{
			ID:       userID,
			Email:    "test@example.com",
			Name:     "Test User",
			Role:     models.RoleNurse,
			IsActive: true,
		}

		svc.On("GetUserByID", mock.Anything, userID).Return(user, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/auth/me", nil, userID, models.RoleNurse)
		rr := httptest.NewRecorder()

		handler.Me(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result models.UserInfo
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, "test@example.com", result.Email)

		svc.AssertExpectations(t)
	})

	t.Run("returns not found when user doesn't exist", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		userID := "user-123"

		svc.On("GetUserByID", mock.Anything, userID).Return(nil, auth.ErrUserNotFound)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/auth/me", nil, userID, models.RoleNurse)
		rr := httptest.NewRecorder()

		handler.Me(rr, r)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestAuthHandler_ChangePassword(t *testing.T) {
	t.Run("successfully changes password", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		userID := "user-123"
		req := &models.ChangePasswordRequest{
			OldPassword: "oldpassword",
			NewPassword: "newpassword123",
		}

		svc.On("ChangePassword", mock.Anything, userID, req).Return(nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/auth/change-password", req, userID, models.RoleNurse)
		rr := httptest.NewRecorder()

		handler.ChangePassword(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]string
		testutil.DecodeJSON(t, rr, &response)
		assert.Equal(t, "password changed successfully", response["message"])

		svc.AssertExpectations(t)
	})

	t.Run("returns unauthorized for wrong old password", func(t *testing.T) {
		svc := new(mockAuthService)
		handler := NewAuthHandler(svc)

		userID := "user-123"
		req := &models.ChangePasswordRequest{
			OldPassword: "wrongpassword",
			NewPassword: "newpassword123",
		}

		svc.On("ChangePassword", mock.Anything, userID, req).Return(auth.ErrInvalidPassword)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/auth/change-password", req, userID, models.RoleNurse)
		rr := httptest.NewRecorder()

		handler.ChangePassword(rr, r)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request for short new password", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		userID := "user-123"
		req := &models.ChangePasswordRequest{
			OldPassword: "oldpassword",
			NewPassword: "short",
		}

		r := testutil.NewRequest(http.MethodPost, "/api/v1/auth/change-password", req, userID, models.RoleNurse)
		rr := httptest.NewRecorder()

		handler.ChangePassword(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns bad request for missing fields", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		userID := "user-123"
		req := &models.ChangePasswordRequest{
			OldPassword: "oldpassword",
			// Missing NewPassword
		}

		r := testutil.NewRequest(http.MethodPost, "/api/v1/auth/change-password", req, userID, models.RoleNurse)
		rr := httptest.NewRecorder()

		handler.ChangePassword(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("successfully logs out", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil, "user-123", models.RoleNurse)
		rr := httptest.NewRecorder()

		handler.Logout(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]string
		testutil.DecodeJSON(t, rr, &response)
		assert.Equal(t, "logout successful", response["message"])
	})
}

func TestAuthHandler_Logout_MethodNotAllowed(t *testing.T) {
	handler := NewAuthHandler(nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/auth/logout", nil, "user-123", models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.Logout(rr, r)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestAuthHandler_Me_MethodNotAllowed(t *testing.T) {
	handler := NewAuthHandler(nil)

	r := testutil.NewRequest(http.MethodPost, "/api/v1/auth/me", nil, "user-123", models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.Me(rr, r)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestAuthHandler_Me_InternalError(t *testing.T) {
	svc := new(mockAuthService)
	handler := NewAuthHandler(svc)

	userID := "user-123"

	svc.On("GetUserByID", mock.Anything, userID).Return(nil, errors.New("database error"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/auth/me", nil, userID, models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.Me(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestAuthHandler_ChangePassword_MethodNotAllowed(t *testing.T) {
	handler := NewAuthHandler(nil)

	userID := "user-123"
	req := &models.ChangePasswordRequest{
		OldPassword: "oldpassword",
		NewPassword: "newpassword123",
	}

	r := testutil.NewRequest(http.MethodGet, "/api/v1/auth/change-password", req, userID, models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, r)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestAuthHandler_ChangePassword_InvalidJSON(t *testing.T) {
	handler := NewAuthHandler(nil)

	userID := "user-123"

	r := testutil.NewRequest(http.MethodPost, "/api/v1/auth/change-password", nil, userID, models.RoleNurse)
	r.Body = http.NoBody
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_ChangePassword_InternalError(t *testing.T) {
	svc := new(mockAuthService)
	handler := NewAuthHandler(svc)

	userID := "user-123"
	req := &models.ChangePasswordRequest{
		OldPassword: "oldpassword",
		NewPassword: "newpassword123",
	}

	svc.On("ChangePassword", mock.Anything, userID, req).Return(errors.New("database error"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/auth/change-password", req, userID, models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}
