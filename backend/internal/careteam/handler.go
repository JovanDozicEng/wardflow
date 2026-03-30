package careteam

import (
	"net/http"
	"strconv"

	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
)

// Handler handles HTTP requests for care team management
type Handler struct {
	service *Service
}

// NewHandler creates a new care team handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{
		service: NewService(db),
	}
}

// GetCareTeam returns the current care team for an encounter
// GET /api/v1/encounters/{encounterId}/care-team
func (h *Handler) GetCareTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encounterID := r.PathValue("encounterId")

	if encounterID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "encounterId is required")
		return
	}

	// Parse query params
	activeOnly := r.URL.Query().Get("activeOnly") == "true"
	withDetails := r.URL.Query().Get("withDetails") == "true"

	if withDetails {
		// Return with user details populated
		response, err := h.service.GetCareTeamWithDetails(ctx, encounterID)
		if err != nil {
			httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
			return
		}
		httputil.RespondJSON(w, http.StatusOK, response)
		return
	}

	// Return raw assignments
	assignments, err := h.service.ListCareTeam(ctx, encounterID, activeOnly)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response := ListAssignmentsResponse{
		Assignments: assignments,
		Total:       int64(len(assignments)),
	}
	httputil.RespondJSON(w, http.StatusOK, response)
}

// AssignRole assigns a user to a role in the care team
// POST /api/v1/encounters/{encounterId}/care-team/assignments
func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encounterID := r.PathValue("encounterId")

	if encounterID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "encounterId is required")
		return
	}

	// Get current user from context
	userCtx := auth.MustGetUserContext(ctx)

	// Parse request body
	var req AssignRoleRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Add RBAC check - only authorized users can assign roles
	// For now, any authenticated user can assign

	assignment, err := h.service.AssignRole(ctx, r, encounterID, req, userCtx.UserID)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "ASSIGNMENT_FAILED", err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, assignment)
}

// TransferRole transfers a role assignment to another user
// POST /api/v1/care-team/assignments/{assignmentId}/transfer
func (h *Handler) TransferRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assignmentID := r.PathValue("assignmentId")

	if assignmentID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "assignmentId is required")
		return
	}

	// Get current user from context
	userCtx := auth.MustGetUserContext(ctx)

	// Parse request body
	var req TransferRoleRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// TODO: Add RBAC check - only authorized users can transfer roles
	// For now, any authenticated user can transfer

	newAssignment, err := h.service.TransferRole(ctx, r, assignmentID, req, userCtx.UserID)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "TRANSFER_FAILED", err.Error())
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, newAssignment)
}

// GetHandoffs returns handoff notes for an encounter
// GET /api/v1/encounters/{encounterId}/handoffs
func (h *Handler) GetHandoffs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encounterID := r.PathValue("encounterId")

	if encounterID == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "encounterId is required")
		return
	}

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

	handoffs, total, err := h.service.GetHandoffs(ctx, encounterID, limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	response := ListHandoffsResponse{
		Handoffs: handoffs,
		Total:    total,
	}
	httputil.RespondJSON(w, http.StatusOK, response)
}
