package handler

import (
	"net/http"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
)

type UsersHandler struct {
	db *database.DB
}

func NewUsersHandler(db *database.DB) *UsersHandler {
	return &UsersHandler{db: db}
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

	var users []models.User
	tx := h.db.DB.Where("is_active = ?", true).Order("name asc")

	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}
	if role != "" {
		tx = tx.Where("role = ?", role)
	}

	if err := tx.Limit(20).Find(&users).Error; err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	summaries := make([]UserSummary, len(users))
	for i, u := range users {
		summaries[i] = UserSummary{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role}
	}
	httputil.RespondJSON(w, http.StatusOK, summaries)
}
