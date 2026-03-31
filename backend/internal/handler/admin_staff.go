package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// AdminStaffHandler handles admin-only staff management endpoints.
type AdminStaffHandler struct {
	db *database.DB
}

func NewAdminStaffHandler(db *database.DB) *AdminStaffHandler {
	return &AdminStaffHandler{db: db}
}

// StaffProfile is the full user profile returned to admins.
type StaffProfile struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Email         string             `json:"email"`
	Role          models.Role        `json:"role"`
	IsActive      bool               `json:"isActive"`
	UnitIDs       models.StringArray `json:"unitIds"`
	DepartmentIDs models.StringArray `json:"departmentIds"`
	CreatedAt     string             `json:"createdAt"`
	UpdatedAt     string             `json:"updatedAt"`
}

func toStaffProfile(u models.User) StaffProfile {
	unitIDs := u.UnitIDs
	if unitIDs == nil {
		unitIDs = models.StringArray{}
	}
	deptIDs := u.DepartmentIDs
	if deptIDs == nil {
		deptIDs = models.StringArray{}
	}
	return StaffProfile{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Role:          u.Role,
		IsActive:      u.IsActive,
		UnitIDs:       unitIDs,
		DepartmentIDs: deptIDs,
		CreatedAt:     u.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     u.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

// ListStaff returns paginated list of all users (admin only).
// GET /api/v1/admin/staff?q=<search>&role=<role>&limit=20&offset=0
func (h *AdminStaffHandler) ListStaff(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserContext(r.Context())
	if !ok || userCtx.Role != models.RoleAdmin {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "admin access required")
		return
	}

	q := r.URL.Query().Get("q")
	role := r.URL.Query().Get("role")
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	tx := h.db.DB.Model(&models.User{}).Order("name asc")
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}
	if role != "" {
		tx = tx.Where("role = ?", role)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	var users []models.User
	if err := tx.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	profiles := make([]StaffProfile, len(users))
	for i, u := range users {
		profiles[i] = toStaffProfile(u)
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]any{
		"data":   profiles,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateStaff updates a user's role, unit/department assignments, and active status (admin only).
// PATCH /api/v1/admin/staff/:userId
type UpdateStaffRequest struct {
	Role          *models.Role        `json:"role,omitempty"`
	IsActive      *bool               `json:"isActive,omitempty"`
	UnitIDs       *models.StringArray `json:"unitIds,omitempty"`
	DepartmentIDs *models.StringArray `json:"departmentIds,omitempty"`
}

func (h *AdminStaffHandler) UpdateStaff(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := auth.GetUserContext(r.Context())
	if !ok || userCtx.Role != models.RoleAdmin {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "admin access required")
		return
	}

	userID := r.PathValue("userId")
	if userID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "MISSING_PARAM", "userId is required")
		return
	}

	var req UpdateStaffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	var user models.User
	if err := h.db.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	updates := map[string]any{}
	if req.Role != nil {
		if !isValidRole(*req.Role) {
			httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_ROLE", "unknown role value")
			return
		}
		updates["role"] = *req.Role
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
		user.IsActive = *req.IsActive
	}
	if req.UnitIDs != nil {
		updates["unit_ids"] = *req.UnitIDs
		user.UnitIDs = *req.UnitIDs
	}
	if req.DepartmentIDs != nil {
		updates["department_ids"] = *req.DepartmentIDs
		user.DepartmentIDs = *req.DepartmentIDs
	}

	if len(updates) == 0 {
		httputil.RespondJSON(w, http.StatusOK, toStaffProfile(user))
		return
	}

	if err := h.db.DB.Model(&user).Updates(updates).Error; err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	// Re-fetch to get updated timestamps
	if err := h.db.DB.First(&user, "id = ?", userID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusOK, toStaffProfile(user))
}

func isValidRole(r models.Role) bool {
	switch r {
	case models.RoleNurse, models.RoleProvider, models.RoleChargeNurse,
		models.RoleOperations, models.RoleConsult, models.RoleTransport,
		models.RoleQualitySafety, models.RoleAdmin:
		return true
	}
	return false
}
