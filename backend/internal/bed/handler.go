package bed

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
)

// Handler handles bed HTTP requests
type Handler struct {
	db *database.DB
}

// NewHandler creates a new bed handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// ListBeds handles GET /api/v1/beds?unitId=&status=&limit=&offset=
func (h *Handler) ListBeds(w http.ResponseWriter, r *http.Request) {
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

	tx := h.db.DB.Model(&Bed{})
	if unitID != "" {
		tx = tx.Where("unit_id = ?", unitID)
	}
	if status != "" {
		tx = tx.Where("current_status = ?", status)
	}

	var total int64
	tx.Count(&total)

	var beds []Bed
	if err := tx.Limit(limit).Offset(offset).Order("room ASC, label ASC").Find(&beds).Error; err != nil {
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
	if req.UnitID == "" || req.Room == "" || req.Label == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "unitId, room, and label are required")
		return
	}

	bed := Bed{
		UnitID:        req.UnitID,
		Room:          req.Room,
		Label:         req.Label,
		Capabilities:  StringSlice(req.Capabilities),
		CurrentStatus: BedStatusAvailable,
	}
	if err := h.db.DB.Create(&bed).Error; err != nil {
		logger.Error("failed to create bed: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create bed")
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "bed",
		EntityID:   bed.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      bed,
	})

	httputil.RespondJSON(w, http.StatusCreated, bed)
}

// GetBed handles GET /api/v1/beds/{bedId}
func (h *Handler) GetBed(w http.ResponseWriter, r *http.Request) {
	bedID := r.PathValue("bedId")
	var bed Bed
	if err := h.db.DB.First(&bed, "id = ?", bedID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "bed not found")
		return
	}
	httputil.RespondJSON(w, http.StatusOK, bed)
}

// UpdateBedStatus handles POST /api/v1/beds/{bedId}/status
func (h *Handler) UpdateBedStatus(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	bedID := r.PathValue("bedId")

	var bed Bed
	if err := h.db.DB.First(&bed, "id = ?", bedID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "bed not found")
		return
	}

	var req UpdateBedStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	if req.Status == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "status is required")
		return
	}

	fromStatus := bed.CurrentStatus
	event := BedStatusEvent{
		BedID:      bedID,
		FromStatus: &fromStatus,
		ToStatus:   req.Status,
		Reason:     req.Reason,
		ChangedBy:  userCtx.UserID,
		ChangedAt:  time.Now().UTC(),
	}
	if err := h.db.DB.Create(&event).Error; err != nil {
		logger.Error("failed to create bed status event: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update bed status")
		return
	}

	updates := map[string]any{"current_status": req.Status}
	if req.Status != BedStatusOccupied {
		updates["current_encounter_id"] = nil
	}
	if err := h.db.DB.Model(&bed).Updates(updates).Error; err != nil {
		logger.Error("failed to update bed: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update bed")
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "bed",
		EntityID:   bedID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Reason:     req.Reason,
		Before:     map[string]any{"status": fromStatus},
		After:      map[string]any{"status": req.Status},
	})

	httputil.RespondJSON(w, http.StatusOK, event)
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

	priority := req.Priority
	if priority == "" {
		priority = "routine"
	}

	bedReq := BedRequest{
		EncounterID:          encounterID,
		RequiredCapabilities: StringSlice(req.RequiredCapabilities),
		Priority:             priority,
		Status:               BedRequestStatusPending,
		CreatedBy:            userCtx.UserID,
	}
	if err := h.db.DB.Create(&bedReq).Error; err != nil {
		logger.Error("failed to create bed request: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create bed request")
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "bed_request",
		EntityID:   bedReq.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      bedReq,
	})

	httputil.RespondJSON(w, http.StatusCreated, bedReq)
}

// AssignBed handles POST /api/v1/bed-requests/{requestId}/assign
func (h *Handler) AssignBed(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	requestID := r.PathValue("requestId")

	var bedReq BedRequest
	if err := h.db.DB.First(&bedReq, "id = ?", requestID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "bed request not found")
		return
	}
	if bedReq.Status != BedRequestStatusPending {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "bed request is not pending")
		return
	}

	var req AssignBedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	if req.BedID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "bedId is required")
		return
	}

	// Check bed is available
	var bed Bed
	if err := h.db.DB.First(&bed, "id = ?", req.BedID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "bed not found")
		return
	}
	if bed.CurrentStatus != BedStatusAvailable {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "bed is not available for assignment")
		return
	}

	// Update bed to occupied
	fromStatus := bed.CurrentStatus
	event := BedStatusEvent{
		BedID:      req.BedID,
		FromStatus: &fromStatus,
		ToStatus:   BedStatusOccupied,
		ChangedBy:  userCtx.UserID,
		ChangedAt:  time.Now().UTC(),
	}
	h.db.DB.Create(&event)
	h.db.DB.Model(&bed).Updates(map[string]any{
		"current_status":      BedStatusOccupied,
		"current_encounter_id": bedReq.EncounterID,
	})

	// Update bed request
	h.db.DB.Model(&bedReq).Updates(map[string]any{
		"status":          BedRequestStatusAssigned,
		"assigned_bed_id": req.BedID,
	})

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "bed_request",
		EntityID:   requestID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		After:      map[string]any{"status": "assigned", "bedId": req.BedID},
	})

	h.db.DB.First(&bedReq, "id = ?", requestID)
	httputil.RespondJSON(w, http.StatusOK, bedReq)
}
