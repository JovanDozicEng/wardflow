package flow

import (
	"net/http"
	"strconv"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// Handler handles HTTP requests for flow tracking
type Handler struct {
	service *Service
	db      *database.DB
}

// NewHandler creates a new flow handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{
		service: NewService(db),
		db:      db,
	}
}

// GetFlowTimeline returns the flow state timeline for an encounter
// GET /api/v1/encounters/{encounterId}/flow
func (h *Handler) GetFlowTimeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encounterID := r.PathValue("encounterId")

	if encounterID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "encounterId is required")
		return
	}

	// Parse query params
	withActors := r.URL.Query().Get("withActors") == "true"
	paginated := r.URL.Query().Get("paginated") == "true"

	if withActors {
		// Return timeline with actor details
		response, err := h.service.GetTimelineWithActors(ctx, encounterID)
		if err != nil {
			httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		httputil.RespondJSON(w, http.StatusOK, response)
		return
	}

	if paginated {
		// Parse pagination params
		limit := 30
		offset := 0

		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		response, err := h.service.GetTimelinePaginated(ctx, encounterID, limit, offset)
		if err != nil {
			httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		httputil.RespondJSON(w, http.StatusOK, response)
		return
	}

	// Return simple timeline
	response, err := h.service.GetTimeline(ctx, encounterID)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httputil.RespondJSON(w, http.StatusOK, response)
}

// RecordTransition records a new state transition
// POST /api/v1/encounters/{encounterId}/flow/transitions
func (h *Handler) RecordTransition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encounterID := r.PathValue("encounterId")

	if encounterID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "encounterId is required")
		return
	}

	// Get current user from context
	userCtx := auth.MustGetUserContext(ctx)

	// Parse request body
	var req CreateTransitionRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Add RBAC check - only authorized users can record transitions
	// For now, any authenticated user can record transitions

	transition, err := h.service.RecordTransition(ctx, r, encounterID, req, userCtx.UserID)
	if err != nil {
		// Check if it's a validation error (invalid transition)
		httputil.RespondError(w, r, http.StatusBadRequest, "TRANSITION_FAILED", err.Error())
		return
	}

	audit.Log(ctx, h.db, r, audit.Entry{
		EntityType: "flow_state_transition",
		EntityID:   transition.ID,
		Action:     "TRANSITION",
		ByUserID:   userCtx.UserID,
		After:      transition,
	})

	httputil.RespondJSON(w, http.StatusCreated, transition)
}

// OverrideTransition records a privileged state transition override
// POST /api/v1/encounters/{encounterId}/flow/override
func (h *Handler) OverrideTransition(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encounterID := r.PathValue("encounterId")

	if encounterID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "encounterId is required")
		return
	}

	// Get current user from context
	userCtx := auth.MustGetUserContext(ctx)

	// Parse request body
	var req OverrideTransitionRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Service will validate RBAC (admin/operations only)
	transition, err := h.service.OverrideTransition(ctx, r, encounterID, req, userCtx.UserID, userCtx.Role)
	if err != nil {
		// Check if it's a permission error
		if err.Error() == "insufficient permissions to override flow transitions; requires admin or operations role" {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		httputil.RespondError(w, r, http.StatusBadRequest, "OVERRIDE_FAILED", err.Error())
		return
	}

	reason := req.Reason
	audit.Log(ctx, h.db, r, audit.Entry{
		EntityType: "flow_state_transition",
		EntityID:   transition.ID,
		Action:     "OVERRIDE",
		ByUserID:   userCtx.UserID,
		Reason:     &reason,
		After:      transition,
	})

	httputil.RespondJSON(w, http.StatusCreated, transition)
}

// GetCurrentState returns the current flow state for an encounter
// GET /api/v1/encounters/{encounterId}/flow/current
func (h *Handler) GetCurrentState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encounterID := r.PathValue("encounterId")

	if encounterID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "encounterId is required")
		return
	}

	currentState, err := h.service.GetCurrentState(ctx, encounterID)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	if currentState == nil {
		httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"encounterId":  encounterID,
			"currentState": nil,
		})
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"encounterId":  encounterID,
		"currentState": *currentState,
	})
}
