package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/logger"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService auth.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.RespondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.Name == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "email, password, and name are required")
		return
	}
	if len(req.Password) < 8 {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "password must be at least 8 characters")
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrEmailExists) {
			httputil.RespondError(w, r, http.StatusConflict, "CONFLICT", "email already exists")
			return
		}
		logger.Error("registration failed: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "registration failed")
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "registration successful",
		"user":    user.ToUserInfo(),
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.RespondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "email and password are required")
		return
	}

	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) || errors.Is(err, auth.ErrInvalidPassword) {
			httputil.RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "invalid email or password")
			return
		}
		if errors.Is(err, auth.ErrUserInactive) {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "user account is inactive")
			return
		}
		logger.Error("login failed: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "login failed")
		return
	}

	logger.Info("user logged in: %s (%s)", response.User.Email, response.User.Role)
	httputil.RespondJSON(w, http.StatusOK, response)
}

// Logout handles user logout (client-side token invalidation)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.RespondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	// JWT is stateless, so logout is handled client-side by removing the token
	// For audit purposes, log the logout
	userCtx, ok := auth.GetUserContext(r.Context())
	if ok {
		logger.Info("user logged out: %s", userCtx.Email)
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}

// Me returns the current authenticated user
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.RespondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	userCtx := auth.MustGetUserContext(r.Context())

	user, err := h.authService.GetUserByID(r.Context(), userCtx.UserID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		logger.Error("failed to get user: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get user")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, user.ToUserInfo())
}

// ChangePassword handles password change requests
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.RespondError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	userCtx := auth.MustGetUserContext(r.Context())

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "old password and new password are required")
		return
	}
	if len(req.NewPassword) < 8 {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "new password must be at least 8 characters")
		return
	}

	err := h.authService.ChangePassword(r.Context(), userCtx.UserID, &req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidPassword) {
			httputil.RespondError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "invalid old password")
			return
		}
		logger.Error("password change failed: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "password change failed")
		return
	}

	logger.Info("password changed for user: %s", userCtx.Email)
	httputil.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "password changed successfully",
	})
}
