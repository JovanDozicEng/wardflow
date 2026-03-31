package discharge

import (
	"context"
	"errors"
	"time"

	"github.com/wardflow/backend/internal/models"
)

// Service errors
var (
	ErrChecklistAlreadyExists = errors.New("discharge checklist already exists for this encounter")
	ErrItemAlreadyCompleted   = errors.New("checklist item is already completed")
	ErrDischargeAlreadyDone   = errors.New("discharge already completed")
	ErrIncompleteChecklist    = errors.New("required checklist items are not complete")
	ErrOverrideForbidden      = errors.New("only admin or charge nurse can override incomplete discharge checklist")
	ErrOverrideReasonRequired = errors.New("override reason is required")
)

// Service defines the business logic interface for discharge operations
type Service interface {
	InitChecklist(ctx context.Context, encounterID, userID string, req InitChecklistRequest) (*DischargeChecklist, error)
	GetChecklist(ctx context.Context, encounterID string) (*DischargeChecklist, error)
	CompleteItem(ctx context.Context, itemID, userID string) (*DischargeChecklistItem, error)
	CompleteDischarge(ctx context.Context, encounterID string, req CompleteDischargeRequest, userID string, userRole models.Role) (*DischargeChecklist, error)
}

type service struct {
	repo Repository
}

// NewService creates a new discharge service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// InitChecklist initializes a discharge checklist with default items
func (s *service) InitChecklist(ctx context.Context, encounterID, userID string, req InitChecklistRequest) (*DischargeChecklist, error) {
	// Check if checklist already exists
	_, err := s.repo.GetChecklistByEncounterID(ctx, encounterID)
	if err == nil {
		// Checklist exists, return conflict error
		return nil, ErrChecklistAlreadyExists
	}
	if !errors.Is(err, ErrNotFound) {
		// Database error
		return nil, err
	}

	// Set default discharge type if not provided
	if req.DischargeType == "" {
		req.DischargeType = "standard"
	}

	// Get default items for the discharge type
	defaults := DefaultItems(req.DischargeType)

	// Create checklist
	checklist := &DischargeChecklist{
		EncounterID:   encounterID,
		DischargeType: req.DischargeType,
		Status:        ChecklistStatusInProgress,
		CreatedBy:     userID,
	}

	// Create items
	items := make([]DischargeChecklistItem, len(defaults))
	for i, d := range defaults {
		items[i] = DischargeChecklistItem{
			Code:     d.Code,
			Label:    d.Label,
			Required: d.Required,
			Status:   ItemStatusOpen,
		}
	}

	// Create checklist with items in a transaction
	if err := s.repo.CreateChecklistWithItems(ctx, checklist, items); err != nil {
		return nil, err
	}

	// Fetch items to populate checklist
	fetchedItems, err := s.repo.GetItemsByChecklistID(ctx, checklist.ID)
	if err != nil {
		return nil, err
	}
	checklist.Items = fetchedItems

	return checklist, nil
}

// GetChecklist retrieves a discharge checklist with its items
func (s *service) GetChecklist(ctx context.Context, encounterID string) (*DischargeChecklist, error) {
	checklist, err := s.repo.GetChecklistByEncounterID(ctx, encounterID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetItemsByChecklistID(ctx, checklist.ID)
	if err != nil {
		return nil, err
	}
	checklist.Items = items

	return checklist, nil
}

// CompleteItem marks a checklist item as complete
func (s *service) CompleteItem(ctx context.Context, itemID, userID string) (*DischargeChecklistItem, error) {
	item, err := s.repo.GetItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if item.Status == ItemStatusDone {
		return nil, ErrItemAlreadyCompleted
	}

	now := time.Now().UTC()
	updates := map[string]any{
		"status":       ItemStatusDone,
		"completed_by": userID,
		"completed_at": now,
	}

	if err := s.repo.UpdateItemFields(ctx, itemID, updates); err != nil {
		return nil, err
	}

	// Refetch the updated item
	return s.repo.GetItemByID(ctx, itemID)
}

// CompleteDischarge completes the discharge process
func (s *service) CompleteDischarge(ctx context.Context, encounterID string, req CompleteDischargeRequest, userID string, userRole models.Role) (*DischargeChecklist, error) {
	// Get checklist
	checklist, err := s.repo.GetChecklistByEncounterID(ctx, encounterID)
	if err != nil {
		return nil, err
	}

	// Check if already completed
	if checklist.Status == ChecklistStatusComplete || checklist.Status == ChecklistStatusOverrideComplete {
		return nil, ErrDischargeAlreadyDone
	}

	// Check for incomplete required items
	incompleteRequired, err := s.repo.GetIncompleteRequiredItems(ctx, checklist.ID)
	if err != nil {
		return nil, err
	}

	// If there are incomplete required items and no override, return error
	if len(incompleteRequired) > 0 && !req.Override {
		return nil, ErrIncompleteChecklist
	}

	// Handle override logic
	if req.Override {
		// Override requires a reason
		if req.Reason == nil || *req.Reason == "" {
			return nil, ErrOverrideReasonRequired
		}
		// Only admin or charge nurse can override
		if userRole != models.RoleAdmin && userRole != models.RoleChargeNurse {
			return nil, ErrOverrideForbidden
		}
	}

	// Determine final status
	finalStatus := ChecklistStatusComplete
	if req.Override {
		finalStatus = ChecklistStatusOverrideComplete
	}

	// Update checklist
	now := time.Now().UTC()
	updates := map[string]any{
		"status":       finalStatus,
		"completed_by": userID,
		"completed_at": now,
	}
	if req.Override && req.Reason != nil {
		updates["override_reason"] = *req.Reason
	}

	if err := s.repo.UpdateChecklistFields(ctx, checklist.ID, updates); err != nil {
		return nil, err
	}

	// Return updated checklist with items
	return s.GetChecklist(ctx, encounterID)
}
