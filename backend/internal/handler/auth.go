package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/logger"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "email, password, and name are required")
		return
	}
	if len(req.Password) < 8 {
		respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrEmailExists) {
			respondError(w, http.StatusConflict, "email already exists")
			return
		}
		logger.Error("registration failed: %v", err)
		respondError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "registration successful",
		"user":    user.ToUserInfo(),
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) || errors.Is(err, auth.ErrInvalidPassword) {
			respondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		if errors.Is(err, auth.ErrUserInactive) {
			respondError(w, http.StatusForbidden, "user account is inactive")
			return
		}
		logger.Error("login failed: %v", err)
		respondError(w, http.StatusInternalServerError, "login failed")
		return
	}

	logger.Info("user logged in: %s (%s)", response.User.Email, response.User.Role)
	respondJSON(w, http.StatusOK, response)
}

// Logout handles user logout (client-side token invalidation)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// JWT is stateless, so logout is handled client-side by removing the token
	// For audit purposes, log the logout
	userCtx, ok := auth.GetUserContext(r.Context())
	if ok {
		logger.Info("user logged out: %s", userCtx.Email)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "logout successful",
	})
}

// Me returns the current authenticated user
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userCtx := auth.MustGetUserContext(r.Context())

	user, err := h.authService.GetUserByID(r.Context(), userCtx.UserID)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		logger.Error("failed to get user: %v", err)
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, user.ToUserInfo())
}

// ChangePassword handles password change requests
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userCtx := auth.MustGetUserContext(r.Context())

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "old password and new password are required")
		return
	}
	if len(req.NewPassword) < 8 {
		respondError(w, http.StatusBadRequest, "new password must be at least 8 characters")
		return
	}

	err := h.authService.ChangePassword(r.Context(), userCtx.UserID, &req)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidPassword) {
			respondError(w, http.StatusUnauthorized, "invalid old password")
			return
		}
		logger.Error("password change failed: %v", err)
		respondError(w, http.StatusInternalServerError, "password change failed")
		return
	}

	logger.Info("password changed for user: %s", userCtx.Email)
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "password changed successfully",
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to encode response: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}
