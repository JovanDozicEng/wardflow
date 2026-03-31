package handler

import (
	"context"
	"net/http"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
)

// UserService defines the service interface for user listing
type UserService interface {
	ListUsers(ctx context.Context, q, role string) ([]UserSummary, error)
}

type userService struct {
	db *database.DB
}

// NewUserService creates a new user service
func NewUserService(db *database.DB) UserService {
	return &userService{db: db}
}

func (s *userService) ListUsers(ctx context.Context, q, role string) ([]UserSummary, error) {
	var users []models.User
	tx := s.db.DB.WithContext(ctx).Where("is_active = ?", true).Order("name asc")

	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}
	if role != "" {
		tx = tx.Where("role = ?", role)
	}

	if err := tx.Limit(20).Find(&users).Error; err != nil {
		return nil, err
	}

	summaries := make([]UserSummary, len(users))
	for i, u := range users {
		summaries[i] = UserSummary{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role}
	}
	return summaries, nil
}

type UsersHandler struct {
	service UserService
}

func NewUsersHandler(service UserService) *UsersHandler {
	return &UsersHandler{service: service}
}

type UserSummary struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Email string      `json:"email"`
	Role  models.Role `json:"role"`
}

// ListUsers returns a searchable list of active users.
// GET /api/v1/users?q=<search>&role=<role>
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	role := r.URL.Query().Get("role")

	summaries, err := h.service.ListUsers(r.Context(), q, role)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusOK, summaries)
}
