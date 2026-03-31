package transport

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

// Handler handles transport HTTP requests
type Handler struct {
	service Service
	db      *database.DB
}

// NewHandler creates a new transport handler
func NewHandler(service Service, db *database.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// ListRequests handles GET /api/v1/transport/requests?status=&unitId=&limit=&offset=
func (h *Handler) ListRequests(w http.ResponseWriter, r *http.Request) {
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

	filter := ListTransportFilter{
		Status:  status,
		UnitID:  filteredUnitID,
		UnitIDs: filteredUnitIDs,
		Limit:   limit,
		Offset:  offset,
	}

	requests, total, err := h.service.ListRequests(r.Context(), filter)
	if err != nil {
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

	tr, err := h.service.CreateRequest(r.Context(), req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to create transport request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "transport_request",
		EntityID:   tr.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      *tr,
	})

	httputil.RespondJSON(w, http.StatusCreated, *tr)
}

// AcceptRequest handles POST /api/v1/transport/requests/{requestId}/accept
func (h *Handler) AcceptRequest(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	requestID := r.PathValue("requestId")

	var req AcceptTransportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	tr, err := h.service.AcceptRequest(r.Context(), requestID, req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to accept transport request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	assignedTo := req.AssignedTo
	if assignedTo == "" {
		assignedTo = userCtx.UserID
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "transport_request",
		EntityID:   requestID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Before:     map[string]any{"status": TransportStatusPending},
		After:      map[string]any{"status": TransportStatusAssigned, "assignedTo": assignedTo},
	})

	httputil.RespondJSON(w, http.StatusOK, *tr)
}

// UpdateRequest handles PATCH /api/v1/transport/requests/{requestId}
func (h *Handler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	requestID := r.PathValue("requestId")

	var req UpdateTransportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	tr, err := h.service.UpdateRequest(r.Context(), requestID, req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to update transport request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Build changedFields for audit
	changedFields := map[string]any{}
	if req.Origin != nil {
		changedFields["origin"] = *req.Origin
	}
	if req.Destination != nil {
		changedFields["destination"] = *req.Destination
	}
	if req.Priority != nil {
		changedFields["priority"] = *req.Priority
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "transport_request",
		EntityID:   requestID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Reason:     req.Reason,
		After:      changedFields,
	})

	httputil.RespondJSON(w, http.StatusOK, *tr)
}

// CompleteRequest handles POST /api/v1/transport/requests/{requestId}/complete
func (h *Handler) CompleteRequest(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	requestID := r.PathValue("requestId")

	tr, err := h.service.CompleteRequest(r.Context(), requestID, userCtx.UserID)
	if err != nil {
		logger.Error("failed to complete transport request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "transport_request",
		EntityID:   requestID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Before:     map[string]any{"status": TransportStatusAssigned},
		After:      map[string]any{"status": TransportStatusCompleted},
	})

	httputil.RespondJSON(w, http.StatusOK, *tr)
}

