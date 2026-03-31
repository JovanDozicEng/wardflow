package discharge

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
)

// Handler handles discharge HTTP requests
type Handler struct {
	service Service
	db      *database.DB // kept for audit logging
}

// NewHandler creates a new discharge handler
func NewHandler(service Service, db *database.DB) *Handler {
	return &Handler{service: service, db: db}
}

// InitChecklist handles POST /api/v1/encounters/{encounterId}/discharge-checklist/init
func (h *Handler) InitChecklist(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	encounterID := r.PathValue("encounterId")

	var req InitChecklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}

	checklist, err := h.service.InitChecklist(r.Context(), encounterID, userCtx.UserID, req)
	if err != nil {
		if errors.Is(err, ErrChecklistAlreadyExists) {
			httputil.RespondError(w, r, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		logger.Error("failed to initialize discharge checklist: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create checklist")
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "discharge_checklist",
		EntityID:   checklist.ID,
		Action:     "CREATE",
		ByUserID:   userCtx.UserID,
		After:      checklist,
	})

	httputil.RespondJSON(w, http.StatusCreated, checklist)
}

// GetChecklist handles GET /api/v1/encounters/{encounterId}/discharge-checklist
func (h *Handler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	encounterID := r.PathValue("encounterId")

	checklist, err := h.service.GetChecklist(r.Context(), encounterID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "discharge checklist not found")
			return
		}
		logger.Error("failed to get discharge checklist: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve checklist")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, checklist)
}

// CompleteItem handles POST /api/v1/discharge-checklist/items/{itemId}/complete
func (h *Handler) CompleteItem(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	itemID := r.PathValue("itemId")

	item, err := h.service.CompleteItem(r.Context(), itemID, userCtx.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "checklist item not found")
			return
		}
		if errors.Is(err, ErrItemAlreadyCompleted) {
			httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "item is already completed")
			return
		}
		logger.Error("failed to complete checklist item: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to complete checklist item")
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "discharge_checklist_item",
		EntityID:   itemID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		After:      map[string]any{"status": ItemStatusDone},
	})

	httputil.RespondJSON(w, http.StatusOK, item)
}

// CompleteDischarge handles POST /api/v1/encounters/{encounterId}/discharge/complete
func (h *Handler) CompleteDischarge(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	encounterID := r.PathValue("encounterId")

	var req CompleteDischargeRequest
	// Ignore decode errors — body may be empty
	json.NewDecoder(r.Body).Decode(&req)

	checklist, err := h.service.CompleteDischarge(r.Context(), encounterID, req, userCtx.UserID, userCtx.Role)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "discharge checklist not found")
			return
		}
		if errors.Is(err, ErrDischargeAlreadyDone) {
			httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "discharge already completed")
			return
		}
		if errors.Is(err, ErrIncompleteChecklist) {
			httputil.RespondError(w, r, http.StatusBadRequest, "INCOMPLETE_CHECKLIST",
				"required checklist items are not complete; use override=true with a reason to proceed")
			return
		}
		if errors.Is(err, ErrOverrideReasonRequired) {
			httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "override reason is required")
			return
		}
		if errors.Is(err, ErrOverrideForbidden) {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "only admin or charge nurse can override incomplete discharge checklist")
			return
		}
		logger.Error("failed to complete discharge: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to complete discharge")
		return
	}

	audit.Log(r.Context(), h.db, r, audit.Entry{
		EntityType: "discharge_checklist",
		EntityID:   checklist.ID,
		Action:     "UPDATE",
		ByUserID:   userCtx.UserID,
		Reason:     req.Reason,
		After:      map[string]any{"status": checklist.Status},
	})

	httputil.RespondJSON(w, http.StatusOK, checklist)
}
