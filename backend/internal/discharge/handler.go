package discharge

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/wardflow/backend/internal/audit"
	"github.com/wardflow/backend/internal/httputil"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/auth"
	"github.com/wardflow/backend/pkg/database"
	"github.com/wardflow/backend/pkg/logger"
	"gorm.io/gorm"
)

// Handler handles discharge HTTP requests
type Handler struct {
	db *database.DB
}

// NewHandler creates a new discharge handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

// InitChecklist handles POST /api/v1/encounters/{encounterId}/discharge-checklist/init
func (h *Handler) InitChecklist(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	encounterID := r.PathValue("encounterId")

	// Check if checklist already exists — distinguish not-found from DB error
	var existing DischargeChecklist
	err := h.db.DB.Where("encounter_id = ?", encounterID).First(&existing).Error
	if err == nil {
		httputil.RespondError(w, r, http.StatusConflict, "CONFLICT", "discharge checklist already exists for this encounter")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("failed to check existing checklist: %v", err)
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to initialize checklist")
		return
	}

	var req InitChecklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body")
		return
	}
	if req.DischargeType == "" {
		req.DischargeType = "standard"
	}

	// Seed checklist and default items in a single transaction
	defaults := DefaultItems(req.DischargeType)
	var checklist DischargeChecklist
	txErr := h.db.DB.Transaction(func(tx *gorm.DB) error {
		checklist = DischargeChecklist{
			EncounterID:   encounterID,
			DischargeType: req.DischargeType,
			Status:        ChecklistStatusInProgress,
			CreatedBy:     userCtx.UserID,
		}
		if err := tx.Create(&checklist).Error; err != nil {
			return fmt.Errorf("failed to create checklist: %w", err)
		}
		for _, d := range defaults {
			item := DischargeChecklistItem{
				ChecklistID: checklist.ID,
				Code:        d.Code,
				Label:       d.Label,
				Required:    d.Required,
				Status:      ItemStatusOpen,
			}
			if err := tx.Create(&item).Error; err != nil {
				return fmt.Errorf("failed to seed item %s: %w", d.Code, err)
			}
		}
		return nil
	})
	if txErr != nil {
		logger.Error("failed to initialize discharge checklist: %v", txErr)
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

	// Return checklist with items
	var items []DischargeChecklistItem
	h.db.DB.Where("checklist_id = ?", checklist.ID).Find(&items)
	checklist.Items = items

	httputil.RespondJSON(w, http.StatusCreated, checklist)
}

// GetChecklist handles GET /api/v1/encounters/{encounterId}/discharge-checklist
func (h *Handler) GetChecklist(w http.ResponseWriter, r *http.Request) {
	encounterID := r.PathValue("encounterId")

	var checklist DischargeChecklist
	if err := h.db.DB.Where("encounter_id = ?", encounterID).First(&checklist).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "discharge checklist not found")
		return
	}

	var items []DischargeChecklistItem
	h.db.DB.Where("checklist_id = ?", checklist.ID).Order("required DESC, code ASC").Find(&items)
	checklist.Items = items

	httputil.RespondJSON(w, http.StatusOK, checklist)
}

// CompleteItem handles POST /api/v1/discharge-checklist/items/{itemId}/complete
func (h *Handler) CompleteItem(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	itemID := r.PathValue("itemId")

	var item DischargeChecklistItem
	if err := h.db.DB.First(&item, "id = ?", itemID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "checklist item not found")
		return
	}
	if item.Status == ItemStatusDone {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "item is already completed")
		return
	}

	now := time.Now().UTC()
	userID := userCtx.UserID
	if err := h.db.DB.Model(&item).Updates(map[string]any{
		"status":       ItemStatusDone,
		"completed_by": userID,
		"completed_at": now,
	}).Error; err != nil {
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

	if err := h.db.DB.First(&item, "id = ?", itemID).Error; err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch updated item")
		return
	}
	httputil.RespondJSON(w, http.StatusOK, item)
}

// CompleteDischarge handles POST /api/v1/encounters/{encounterId}/discharge/complete
func (h *Handler) CompleteDischarge(w http.ResponseWriter, r *http.Request) {
	userCtx := auth.MustGetUserContext(r.Context())
	encounterID := r.PathValue("encounterId")

	var checklist DischargeChecklist
	if err := h.db.DB.Where("encounter_id = ?", encounterID).First(&checklist).Error; err != nil {
		httputil.RespondError(w, r, http.StatusNotFound, "NOT_FOUND", "discharge checklist not found")
		return
	}
	if checklist.Status == ChecklistStatusComplete || checklist.Status == ChecklistStatusOverrideComplete {
		httputil.RespondError(w, r, http.StatusBadRequest, "INVALID_STATE", "discharge already completed")
		return
	}

	var req CompleteDischargeRequest
	// Ignore decode errors — body may be empty
	json.NewDecoder(r.Body).Decode(&req)

	// Check required items
	var incompleteRequired []DischargeChecklistItem
	h.db.DB.Where("checklist_id = ? AND required = ? AND status = ?", checklist.ID, true, ItemStatusOpen).Find(&incompleteRequired)

	if len(incompleteRequired) > 0 && !req.Override {
		httputil.RespondError(w, r, http.StatusBadRequest, "INCOMPLETE_CHECKLIST",
			"required checklist items are not complete; use override=true with a reason to proceed")
		return
	}

	// Override requires reason and privileged role
	if req.Override {
		if req.Reason == nil || *req.Reason == "" {
			httputil.RespondError(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "override reason is required")
			return
		}
		if userCtx.Role != models.RoleAdmin && userCtx.Role != models.RoleChargeNurse {
			httputil.RespondError(w, r, http.StatusForbidden, "FORBIDDEN", "only admin or charge nurse can override incomplete discharge checklist")
			return
		}
	}

	now := time.Now().UTC()
	userID := userCtx.UserID
	finalStatus := ChecklistStatusComplete
	if req.Override {
		finalStatus = ChecklistStatusOverrideComplete
	}

	updates := map[string]any{
		"status":       finalStatus,
		"completed_by": userID,
		"completed_at": now,
	}
	if req.Override && req.Reason != nil {
		updates["override_reason"] = *req.Reason
	}
	if err := h.db.DB.Model(&checklist).Updates(updates).Error; err != nil {
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
		After:      map[string]any{"status": finalStatus},
	})

	if err := h.db.DB.Where("encounter_id = ?", encounterID).First(&checklist).Error; err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch updated checklist")
		return
	}
	var items []DischargeChecklistItem
	h.db.DB.Where("checklist_id = ?", checklist.ID).Order("required DESC, code ASC").Find(&items)
	checklist.Items = items

	httputil.RespondJSON(w, http.StatusOK, checklist)
}
