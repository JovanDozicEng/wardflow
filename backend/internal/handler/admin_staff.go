package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// StaffService defines the service interface for staff management
type StaffService interface {
	ListStaff(ctx context.Context, q, role string, limit, offset int) ([]StaffProfile, int64, error)
	UpdateStaff(ctx context.Context, userID string, req UpdateStaffRequest) (*StaffProfile, error)
}

type staffService struct {
	db *database.DB
}

// NewStaffService creates a new staff service
func NewStaffService(db *database.DB) StaffService {
	return &staffService{db: db}
}

func (s *staffService) ListStaff(ctx context.Context, q, role string, limit, offset int) ([]StaffProfile, int64, error) {
	tx := s.db.DB.WithContext(ctx).Model(&models.User{}).Order("name asc")
	if q != "" {
		like := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}
	if role != "" {
		tx = tx.Where("role = ?", role)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.User
	if err := tx.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	profiles := make([]StaffProfile, len(users))
	for i, u := range users {
		profiles[i] = toStaffProfile(u)
	}

	return profiles, total, nil
}

func (s *staffService) UpdateStaff(ctx context.Context, userID string, req UpdateStaffRequest) (*StaffProfile, error) {
	var user models.User
	if err := s.db.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	updates := map[string]any{}
	if req.Role != nil {
		if !isValidRole(*req.Role) {
			return nil, errors.New("invalid role")
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
		profile := toStaffProfile(user)
		return &profile, nil
	}

	if err := s.db.DB.WithContext(ctx).Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Re-fetch to get updated timestamps
	if err := s.db.DB.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	profile := toStaffProfile(user)
	return &profile, nil
}

// AdminStaffHandler handles admin-only staff management endpoints.
type AdminStaffHandler struct {
	service StaffService
}

func NewAdminStaffHandler(service StaffService) *AdminStaffHandler {
	return &AdminStaffHandler{service: service}
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

	profiles, total, err := h.service.ListStaff(r.Context(), q, role, limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
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

	profile, err := h.service.UpdateStaff(r.Context(), userID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "user not found")
			return
		}
		if err.Error() == "invalid role" {
			httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_ROLE", "unknown role value")
			return
		}
		httputil.RespondError(w, r, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusOK, profile)
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
