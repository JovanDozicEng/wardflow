package bed

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
)

// Handler handles bed HTTP requests
type Handler struct {
	service Service
	db      *database.DB
}

// NewHandler creates a new bed handler
func NewHandler(service Service, db *database.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// ListBeds handles GET /api/v1/beds?unitId=&status=&limit=&offset=
func (h *Handler) ListBeds(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	q := r.URL.Query()
	unitID := q.Get("unitId")
	status := q.Get("status")
	limit := 50
	offset := 0
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v >= 0 {
		offset = v
	}

	// Unit-scoped RBAC: apply unit filter before calling service
	isAdmin := userCtx.Role == models.RoleAdmin
	filteredUnitID := unitID
	var filteredUnitIDs []string
	if !isAdmin {
		if unitID != "" {
			// Validate the requested unit is in user's authorized units
			allowed := false
			for _, u := range userCtx.UnitIDs {
				if u == unitID {
					allowed = true
					break
				}
			}
			if !allowed {
				httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "not authorized to access the requested unit")
				return
			}
		} else {
			// Non-admin without specific unit filter: restrict to their units
			filteredUnitIDs = userCtx.UnitIDs
			filteredUnitID = "" // Clear single unit filter
		}
	}

	filter := ListBedsFilter{
		UnitID:  filteredUnitID,
		UnitIDs: filteredUnitIDs,
		Status:  status,
		Limit:   limit,
		Offset:  offset,
	}

	beds, total, err := h.service.ListBeds(r.Context(), filter)
	if err != nil {
		logger.Error("failed to list beds: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list beds")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]any{
		"data":   beds,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateBed handles POST /api/v1/beds (admin only)
func (h *Handler) CreateBed(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	if userCtx.Role != models.RoleAdmin {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "admin access required")
		return
	}

	var req CreateBedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	bed, err := h.service.CreateBed(r.Context(), req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to create bed: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "bed",
		EntityID:   bed.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      *bed,
	})

	httputil.RespondJSON(w, http.StatusCreated, *bed)
}

// GetBed handles GET /api/v1/beds/{bedId}
func (h *Handler) GetBed(w http.ResponseWriter, r *http.Request) {
	bedID := r.PathValue("bedId")
	bed, err := h.service.GetBed(r.Context(), bedID)
	if err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "bed not found")
		return
	}
	httputil.RespondJSON(w, http.StatusOK, *bed)
}

// UpdateBedStatus handles POST /api/v1/beds/{bedId}/status
func (h *Handler) UpdateBedStatus(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	bedID := r.PathValue("bedId")

	// Only operations, charge_nurse, and admin may change bed status
	allowedRoles := map[models.Role]bool{
		models.RoleAdmin:       true,
		models.RoleOperations:  true,
		models.RoleChargeNurse: true,
	}
	if !allowedRoles[userCtx.Role] {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "only operations, charge nurse, or admin can update bed status")
		return
	}

	var req UpdateBedStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	event, err := h.service.UpdateBedStatus(r.Context(), bedID, req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to update bed status: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	fromStatus := event.FromStatus
	var fromStatusVal BedStatus
	if fromStatus != nil {
		fromStatusVal = *fromStatus
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "bed",
		EntityID:   bedID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Reason:     req.Reason,
		Before:     map[string]any{"status": fromStatusVal},
		After:      map[string]any{"status": event.ToStatus},
	})

	httputil.RespondJSON(w, http.StatusOK, *event)
}

// CreateBedRequest handles POST /api/v1/encounters/{encounterId}/bed-requests
func (h *Handler) CreateBedRequest(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	encounterID := r.PathValue("encounterId")

	var req CreateBedRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	bedReq, err := h.service.CreateBedRequest(r.Context(), encounterID, userCtx.UserID, req)
	if err != nil {
		logger.Error("failed to create bed request: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "bed_request",
		EntityID:   bedReq.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      *bedReq,
	})

	httputil.RespondJSON(w, http.StatusCreated, *bedReq)
}

// AssignBed handles POST /api/v1/bed-requests/{requestId}/assign
func (h *Handler) AssignBed(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	requestID := r.PathValue("requestId")

	var req AssignBedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	bedReq, err := h.service.AssignBed(r.Context(), requestID, req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to assign bed: %v", err)
		httputil.RespondError(w, r, http.StatusConflict, "ASSIGN_FAILED", err.Error())
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "bed_request",
		EntityID:   requestID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		After:      map[string]any{"status": "assigned", "bedId": bedReq.AssignedBedID},
	})

	httputil.RespondJSON(w, http.StatusOK, *bedReq)
}
