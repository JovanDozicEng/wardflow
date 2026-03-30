package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
)

// Handler handles transport HTTP requests
type Handler struct {
	db *database.DB
}

// NewHandler creates a new transport handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// ListRequests handles GET /api/v1/transport/requests?status=&unitId=&limit=&offset=
func (h *Handler) ListRequests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := q.Get("status")
	limit := 50
	offset := 0
	if v, err := strconv.Atoi(q.Get("limit")); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(q.Get("offset")); err == nil && v >= 0 {
		offset = v
	}

	tx := h.db.DB.Model(&TransportRequest{})
	if status != "" {
		tx = tx.Where("status = ?", status)
	}

	var total int64
	tx.Count(&total)

	var requests []TransportRequest
	if err := tx.Limit(limit).Offset(offset).Order("created_at DESC").Find(&requests).Error; err != nil {
		logger.Error("failed to list transport requests: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list transport requests")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]any{
		"data":   requests,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateRequest handles POST /api/v1/transport/requests
func (h *Handler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	var req CreateTransportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	if req.EncounterID == "" || req.Origin == "" || req.Destination == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "encounterId, origin, and destination are required")
		return
	}
	if req.Priority == "" {
		req.Priority = "routine"
	}

	tr := TransportRequest{
		EncounterID: req.EncounterID,
		Origin:      req.Origin,
		Destination: req.Destination,
		Priority:    req.Priority,
		Status:      TransportStatusPending,
		CreatedBy:   userCtx.UserID,
	}
	if err := h.db.DB.Create(&tr).Error; err != nil {
		logger.Error("failed to create transport request: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create transport request")
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "transport_request",
		EntityID:   tr.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      tr,
	})

	httputil.RespondJSON(w, http.StatusCreated, tr)
}

// AcceptRequest handles POST /api/v1/transport/requests/{requestId}/accept
func (h *Handler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	requestID := r.PathValue("requestId")

	var tr TransportRequest
	if err := h.db.DB.First(&tr, "id = ?", requestID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "transport request not found")
		return
	}
	if tr.Status != TransportStatusPending {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "transport request is not pending")
		return
	}

	var req AcceptTransportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	assignedTo := req.AssignedTo
	if assignedTo == "" {
		assignedTo = userCtx.UserID
	}
	now := time.Now().UTC()

	if err := h.db.DB.Model(&tr).Updates(map[string]any{
		"status":      TransportStatusAssigned,
		"assigned_to": assignedTo,
		"assigned_at": now,
	}).Error; err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to accept transport request")
		return
	}

	// Record change event
	changedFields, _ := json.Marshal(map[string]any{
		"status":     TransportStatusAssigned,
		"assignedTo": assignedTo,
	})
	h.db.DB.Create(&TransportChangeEvent{
		RequestID:     requestID,
		ChangedFields: string(changedFields),
		ChangedBy:     userCtx.UserID,
		ChangedAt:     now,
	})

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "transport_request",
		EntityID:   requestID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Before:     map[string]any{"status": TransportStatusPending},
		After:      map[string]any{"status": TransportStatusAssigned, "assignedTo": assignedTo},
	})

	h.db.DB.First(&tr, "id = ?", requestID)
	httputil.RespondJSON(w, http.StatusOK, tr)
}

// UpdateRequest handles PATCH /api/v1/transport/requests/{requestId}
func (h *Handler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	requestID := r.PathValue("requestId")

	var tr TransportRequest
	if err := h.db.DB.First(&tr, "id = ?", requestID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "transport request not found")
		return
	}
	if tr.Status == TransportStatusCompleted || tr.Status == TransportStatusCancelled {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "cannot update a completed or cancelled request")
		return
	}

	var req UpdateTransportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	updates := map[string]any{}
	changedFields := map[string]any{}
	if req.Origin != nil {
		updates["origin"] = *req.Origin
		changedFields["origin"] = *req.Origin
	}
	if req.Destination != nil {
		updates["destination"] = *req.Destination
		changedFields["destination"] = *req.Destination
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
		changedFields["priority"] = *req.Priority
	}

	if len(updates) == 0 {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "no fields to update")
		return
	}

	h.db.DB.Model(&tr).Updates(updates)

	now := time.Now().UTC()
	cf, _ := json.Marshal(changedFields)
	h.db.DB.Create(&TransportChangeEvent{
		RequestID:     requestID,
		ChangedFields: string(cf),
		ChangedBy:     userCtx.UserID,
		Reason:        req.Reason,
		ChangedAt:     now,
	})

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "transport_request",
		EntityID:   requestID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Reason:     req.Reason,
		After:      changedFields,
	})

	h.db.DB.First(&tr, "id = ?", requestID)
	httputil.RespondJSON(w, http.StatusOK, tr)
}

// CompleteRequest handles POST /api/v1/transport/requests/{requestId}/complete
func (h *Handler) CompleteRequest(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	requestID := r.PathValue("requestId")

	var tr TransportRequest
	if err := h.db.DB.First(&tr, "id = ?", requestID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "transport request not found")
		return
	}
	if tr.Status != TransportStatusAssigned {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "transport request must be accepted before completing")
		return
	}

	now := time.Now().UTC()
	h.db.DB.Model(&tr).Update("status", TransportStatusCompleted)

	cf, _ := json.Marshal(map[string]any{"status": TransportStatusCompleted})
	h.db.DB.Create(&TransportChangeEvent{
		RequestID:     requestID,
		ChangedFields: string(cf),
		ChangedBy:     userCtx.UserID,
		ChangedAt:     now,
	})

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "transport_request",
		EntityID:   requestID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Before:     map[string]any{"status": TransportStatusAssigned},
		After:      map[string]any{"status": TransportStatusCompleted},
	})

	h.db.DB.First(&tr, "id = ?", requestID)
	httputil.RespondJSON(w, http.StatusOK, tr)
}
