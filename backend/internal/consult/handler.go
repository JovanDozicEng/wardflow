package consult

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

// Handler handles consult request HTTP requests
type Handler struct {
	service Service
	db      *database.DB
}

// NewHandler creates a new consult request handler
func NewHandler(service Service, db *database.DB) *Handler {
	return &Handler{
		service: service,
		db:      db,
	}
}

// hasProviderAccess checks if user has provider, consult, or admin role
func hasProviderAccess(role models.Role) bool {
	return role == models.RoleProvider || role == models.RoleConsult || role == models.RoleAdmin
}

// List handles GET requests to list consult requests
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Parse query parameters
	query := r.URL.Query()
	filter := ListConsultsFilter{
		UnitID:        query.Get("unitId"),
		TargetService: query.Get("targetService"),
		Limit:         20, // default
		Offset:        0,
	}

	if status := query.Get("status"); status != "" {
		filter.Status = ConsultStatus(status)
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

	consults, total, err := h.service.List(r.Context(), filter)
	if err != nil {
		logger.Error("failed to list consult requests: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list consult requests")
		return
	}

	response := models.PaginatedResponse{
		Data:   consults,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	httputil.RespondJSON(w, http.StatusOK, response)
}

// Create handles POST requests to create a consult request
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	var req CreateConsultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	consult, err := h.service.Create(r.Context(), &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to create consult request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "consult",
		EntityID:   consult.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      consult,
	})

	httputil.RespondJSON(w, http.StatusCreated, consult)
}

// Accept handles POST requests to accept a consult request
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Check RBAC: only provider, consult, or admin
	if !hasProviderAccess(userCtx.Role) {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	// Extract ID from path
	id := r.PathValue("consultId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "consult ID is required")
		return
	}

	// Get original for audit
	before, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "consult request not found")
			return
		}
		logger.Error("failed to get consult request: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get consult request")
		return
	}

	consult, err := h.service.Accept(r.Context(), id, userCtx.UserID)
	if err != nil {
		logger.Error("failed to accept consult request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "consult",
		EntityID:   consult.ID,
		Action:     "ACCEPT",
		ByUserID:   userCtx.UserID,
		Before:     before,
		After:      consult,
	})

	httputil.RespondJSON(w, http.StatusOK, consult)
}

// Decline handles POST requests to decline a consult request
func (h *Handler) Decline(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Check RBAC: only provider, consult, or admin
	if !hasProviderAccess(userCtx.Role) {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	// Extract ID from path
	id := r.PathValue("consultId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "consult ID is required")
		return
	}

	var req DeclineConsultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Get original for audit
	before, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "consult request not found")
			return
		}
		logger.Error("failed to get consult request: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get consult request")
		return
	}

	consult, err := h.service.Decline(r.Context(), id, &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to decline consult request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "consult",
		EntityID:   consult.ID,
		Action:     "DECLINE",
		ByUserID:   userCtx.UserID,
		Before:     before,
		After:      consult,
	})

	httputil.RespondJSON(w, http.StatusOK, consult)
}

// Redirect handles POST requests to redirect a consult request
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Check RBAC: only provider, consult, or admin
	if !hasProviderAccess(userCtx.Role) {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	// Extract ID from path
	id := r.PathValue("consultId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "consult ID is required")
		return
	}

	var req RedirectConsultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	// Get original for audit
	before, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "consult request not found")
			return
		}
		logger.Error("failed to get consult request: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get consult request")
		return
	}

	result, err := h.service.Redirect(r.Context(), id, &req, userCtx.UserID)
	if err != nil {
		logger.Error("failed to redirect consult request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Audit the closure of the original consult
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "consult",
		EntityID:   result.Original.ID,
		Action:     "REDIRECT",
		ByUserID:   userCtx.UserID,
		Before:     before,
		After:      result.Original,
	})

	// Audit the creation of the new consult
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "consult",
		EntityID:   result.NewConsult.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      result.NewConsult,
	})

	httputil.RespondJSON(w, http.StatusOK, result)
}

// Complete handles POST requests to complete a consult request
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())

	// Check RBAC: only provider, consult, or admin
	if !hasProviderAccess(userCtx.Role) {
		httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "insufficient permissions")
		return
	}

	// Extract ID from path
	id := r.PathValue("consultId")
	if id == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "consult ID is required")
		return
	}

	// Get original for audit
	before, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "consult request not found")
			return
		}
		logger.Error("failed to get consult request: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get consult request")
		return
	}

	consult, err := h.service.Complete(r.Context(), id, userCtx.UserID)
	if err != nil {
		logger.Error("failed to complete consult request: %v", err)
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Write audit log
	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "consult",
		EntityID:   consult.ID,
		Action:     "COMPLETE",
		ByUserID:   userCtx.UserID,
		Before:     before,
		After:      consult,
	})

	httputil.RespondJSON(w, http.StatusOK, consult)
}
