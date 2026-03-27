package dashboard

import (
	"net/http"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// Handler handles HTTP requests for dashboard
type Handler struct {
	service *Service
}

// NewHandler creates a new dashboard handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{
		service: NewService(db),
	}
}

// GetHuddleDashboard returns aggregated metrics for the daily huddle
// GET /api/v1/dashboard/huddle
func (h *Handler) GetHuddleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := auth.MustGetUserContext(ctx)

	// Parse filter params
	filter := FilterParams{}

	if unitID := r.URL.Query().Get("unitId"); unitID != "" {
		filter.UnitID = &unitID
	}
	if deptID := r.URL.Query().Get("departmentId"); deptID != "" {
		filter.DepartmentID = &deptID
	}

	// Get metrics with RBAC enforcement
	metrics, err := h.service.GetHuddleMetrics(ctx, filter, userCtx.Role, userCtx.UnitIDs, userCtx.DeptIDs)
	if err != nil {
		if err.Error() == "unauthorized access to unit" || err.Error() == "unauthorized access to department" {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusOK, metrics)
}
