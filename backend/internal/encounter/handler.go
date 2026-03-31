package encounter

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

// Handler handles encounter HTTP requests
type Handler struct {
	service Service
	db      *database.DB
}

// NewHandler creates a new encounter handler
func NewHandler(service Service, db *database.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// List handles GET requests to list encounters
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Parse query parameters
	query := r.URL.Query()
	filter := ListEncountersFilter{
		UnitID:       query.Get("unitId"),
		DepartmentID: query.Get("departmentId"),
		Status:       query.Get("status"),
		Limit:        20, // default
		Offset:       0,
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

	encounters, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		logger.Error("failed to list encounters: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list encounters")
		return
	}

	response := models.PaginatedResponse{
		Data:   encounters,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	httputil.RespondJSON(w, http.StatusOK, response)
}

// Create handles POST requests to create an encounter
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	var req CreateEncounterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Check unit access (unless admin)
	if userCtx.Role != models.RoleAdmin {
		hasAccess := false
		for _, unitID := range userCtx.UnitIDs {
			if unitID == req.UnitID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "no access to this unit")
			return
		}
	}

	encounter, err := h.service.Create(r.Context(), &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to create encounter: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "encounter",
		EntityID:   encounter.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      encounter,
	})

	httputil.RespondJSON(w, http.StatusCreated, encounter)
}

// GetByID handles GET requests to retrieve a single encounter
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Extract ID from path (Go 1.22+ pattern)
	id := r.PathValue("encounterId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "encounter ID is required")
		return
	}

	encounter, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "encounter not found")
			return
		}
		logger.Error("failed to get encounter: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get encounter")
		return
	}

	// Check unit access (unless admin)
	if userCtx.Role != models.RoleAdmin {
		hasAccess := false
		for _, unitID := range userCtx.UnitIDs {
			if unitID == encounter.UnitID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "no access to this unit")
			return
		}
	}

	httputil.RespondJSON(w, http.StatusOK, encounter)
}

// Update handles PATCH requests to update an encounter
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Extract ID from path (Go 1.22+ pattern)
	id := r.PathValue("encounterId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "encounter ID is required")
		return
	}

	// Get the existing encounter first to check unit access
	existingEncounter, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "encounter not found")
			return
		}
		logger.Error("failed to get encounter: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get encounter")
		return
	}

	// Check unit access (unless admin)
	if userCtx.Role != models.RoleAdmin {
		hasAccess := false
		for _, unitID := range userCtx.UnitIDs {
			if unitID == existingEncounter.UnitID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "no access to this unit")
			return
		}
	}

	var req UpdateEncounterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// If changing unit, check access to new unit as well
	if req.UnitID != nil && userCtx.Role != models.RoleAdmin {
		hasAccess := false
		for _, unitID := range userCtx.UnitIDs {
			if unitID == *req.UnitID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "no access to the new unit")
			return
		}
	}

	encounter, err := h.service.Update(r.Context(), id, &req, userCtx.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "encounter not found")
			return
		}
		logger.Error("failed to update encounter: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "encounter",
		EntityID:   encounter.ID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Before:     existingEncounter,
		After:      encounter,
	})

	httputil.RespondJSON(w, http.StatusOK, encounter)
}
