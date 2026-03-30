package incident

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
)

// Handler handles incident HTTP requests
type Handler struct {
	service *Service
	db      *database.DB
}

// NewHandler creates a new incident handler
func NewHandler(service *Service, db *database.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// hasStatusUpdateAccess checks if user has quality_safety or admin role
func hasStatusUpdateAccess(role models.Role) bool {
	return role == models.RoleQualitySafety || role == models.RoleAdmin
}

// List handles GET requests to list incidents
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Parse query parameters
	query := r.URL.Query()
	filter := ListIncidentsFilter{
		UnitID: query.Get("unitId"),
		Type:   query.Get("type"),
		Limit:  20, // default
		Offset: 0,
	}

	if status := query.Get("status"); status != "" {
		filter.Status = IncidentStatus(status)
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Check unit access if filtering by unit
	if filter.UnitID != "" && userCtx.Role != models.RoleAdmin {
		hasAccess := false
		for _, unitID := range userCtx.UnitIDs {
			if unitID == filter.UnitID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "no access to this unit")
			return
		}
	}

	incidents, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		logger.Error("failed to list incidents: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list incidents")
		return
	}

	response := models.PaginatedResponse{
		Data:   incidents,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	httputil.RespondJSON(w, http.StatusOK, response)
}

// Create handles POST requests to create an incident
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	var req CreateIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	incident, err := h.service.Create(r.Context(), &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to create incident: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "incident",
		EntityID:   incident.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      incident,
	})

	httputil.RespondJSON(w, http.StatusCreated, incident)
}

// GetByID handles GET requests to retrieve a single incident
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	id := r.PathValue("incidentId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "incident ID is required")
		return
	}

	incident, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "incident not found")
			return
		}
		logger.Error("failed to get incident: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get incident")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, incident)
}

// UpdateStatus handles POST requests to update incident status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Check RBAC: only quality_safety and admin
	if !hasStatusUpdateAccess(userCtx.Role) {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	// Extract ID from path
	id := r.PathValue("incidentId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "incident ID is required")
		return
	}

	var req UpdateIncidentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Get original for audit
	before, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "incident not found")
			return
		}
		logger.Error("failed to get incident: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get incident")
		return
	}

	incident, err := h.service.UpdateStatus(r.Context(), id, &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to update incident status: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "incident",
		EntityID:   incident.ID,
		Action:     "UPDATE_STATUS",
		ByUserID:   userCtx.UserID,
		Before:     before,
		After:      incident,
	})

	httputil.RespondJSON(w, http.StatusOK, incident)
}

// GetStatusHistory handles GET requests to retrieve incident status history
func (h *Handler) GetStatusHistory(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	id := r.PathValue("incidentId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "incident ID is required")
		return
	}

	events, err := h.service.GetStatusHistory(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "incident not found")
			return
		}
		logger.Error("failed to get incident status history: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get incident status history")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, events)
}
