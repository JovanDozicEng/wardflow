package exception

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

// Handler handles exception event HTTP requests
type Handler struct {
	service *Service
	db      *database.DB
}

// NewHandler creates a new exception event handler
func NewHandler(service *Service, db *database.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// hasUpdateAccess checks if user has provider, charge_nurse, quality_safety, or admin role
func hasUpdateAccess(role models.Role) bool {
	return role == models.RoleProvider ||
		role == models.RoleChargeNurse ||
		role == models.RoleQualitySafety ||
		role == models.RoleAdmin
}

// hasCorrectAccess checks if user has quality_safety or admin role
func hasCorrectAccess(role models.Role) bool {
	return role == models.RoleQualitySafety || role == models.RoleAdmin
}

// List handles GET requests to list exception events
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	filter := ListExceptionsFilter{
		EncounterID: query.Get("encounterId"),
		Type:        query.Get("type"),
		Limit:       20, // default
		Offset:      0,
	}

	if status := query.Get("status"); status != "" {
		filter.Status = ExceptionStatus(status)
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

	exceptions, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		logger.Error("failed to list exception events: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list exception events")
		return
	}

	response := models.PaginatedResponse{
		Data:   exceptions,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	httputil.RespondJSON(w, http.StatusOK, response)
}

// Create handles POST requests to create an exception event
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	var req CreateExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	exception, err := h.service.Create(r.Context(), &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to create exception event: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "exception",
		EntityID:   exception.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      exception,
	})

	httputil.RespondJSON(w, http.StatusCreated, exception)
}

// Update handles PATCH requests to update an exception event
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Check RBAC
	if !hasUpdateAccess(userCtx.Role) {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	// Extract ID from path
	id := r.PathValue("exceptionId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "exception ID is required")
		return
	}

	var req UpdateExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Get original for audit
	before, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "exception event not found")
			return
		}
		logger.Error("failed to get exception event: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get exception event")
		return
	}

	exception, err := h.service.Update(r.Context(), id, &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to update exception event: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "exception",
		EntityID:   exception.ID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Before:     before,
		After:      exception,
	})

	httputil.RespondJSON(w, http.StatusOK, exception)
}

// Finalize handles POST requests to finalize an exception event
func (h *Handler) Finalize(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Check RBAC
	if !hasUpdateAccess(userCtx.Role) {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	// Extract ID from path
	id := r.PathValue("exceptionId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "exception ID is required")
		return
	}

	// Get original for audit
	before, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "exception event not found")
			return
		}
		logger.Error("failed to get exception event: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get exception event")
		return
	}

	exception, err := h.service.Finalize(r.Context(), id, userCtx.UserID)
	if err != nil {
		logger.Error("failed to finalize exception event: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "exception",
		EntityID:   exception.ID,
		Action:     "FINALIZE",
		ByUserID:   userCtx.UserID,
		Before:     before,
		After:      exception,
	})

	httputil.RespondJSON(w, http.StatusOK, exception)
}

// Correct handles POST requests to correct a finalized exception event
func (h *Handler) Correct(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Check RBAC: only quality_safety and admin
	if !hasCorrectAccess(userCtx.Role) {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	// Extract ID from path
	id := r.PathValue("exceptionId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "exception ID is required")
		return
	}

	var req CorrectExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Get original for audit (will be marked as corrected)
	before, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "exception event not found")
			return
		}
		logger.Error("failed to get exception event: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get exception event")
		return
	}

	// Correct returns the NEW exception event
	newException, err := h.service.Correct(r.Context(), id, &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to correct exception event: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log for the correction (log the original being corrected)
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "exception",
		EntityID:   before.ID,
		Action:     "CORRECT",
		ByUserID:   userCtx.UserID,
		Before:     before,
		After:      map[string]interface{}{
			"corrected":          true,
			"correctedByEventId": newException.ID,
		},
	})

	// Return the NEW exception event
	httputil.RespondJSON(w, http.StatusCreated, newException)
}
