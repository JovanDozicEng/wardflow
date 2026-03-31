package discharge

import (
	"context"
	"errors"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a discharge checklist is not found
var ErrNotFound = errors.New("discharge checklist not found")

// Repository defines the data access interface for discharge operations
type Repository interface {
	GetChecklistByEncounterID(ctx context.Context, encounterID string) (*DischargeChecklist, error)
	CreateChecklistWithItems(ctx context.Context, checklist *DischargeChecklist, items []DischargeChecklistItem) error
	GetItemByID(ctx context.Context, id string) (*DischargeChecklistItem, error)
	UpdateItemFields(ctx context.Context, itemID string, updates map[string]any) error
	GetItemsByChecklistID(ctx context.Context, checklistID string) ([]DischargeChecklistItem, error)
	GetIncompleteRequiredItems(ctx context.Context, checklistID string) ([]DischargeChecklistItem, error)
	UpdateChecklistFields(ctx context.Context, checklistID string, updates map[string]any) error
}

type repository struct {
	db *database.DB
}

// NewRepository creates a new discharge repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

// GetChecklistByEncounterID retrieves a checklist by encounter ID
func (r *repository) GetChecklistByEncounterID(ctx context.Context, encounterID string) (*DischargeChecklist, error) {
	var checklist DischargeChecklist
	err := r.db.DB.WithContext(ctx).Where("encounter_id = ?", encounterID).First(&checklist).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &checklist, nil
}

// CreateChecklistWithItems creates a checklist with items in a transaction
func (r *repository) CreateChecklistWithItems(ctx context.Context, checklist *DischargeChecklist, items []DischargeChecklistItem) error {
	return r.db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(checklist).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].ChecklistID = checklist.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetItemByID retrieves a checklist item by ID
func (r *repository) GetItemByID(ctx context.Context, id string) (*DischargeChecklistItem, error) {
	var item DischargeChecklistItem
	err := r.db.DB.WithContext(ctx).First(&item, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

// UpdateItemFields updates specific fields of a checklist item
func (r *repository) UpdateItemFields(ctx context.Context, itemID string, updates map[string]any) error {
	return r.db.DB.WithContext(ctx).Model(&DischargeChecklistItem{}).Where("id = ?", itemID).Updates(updates).Error
}

// GetItemsByChecklistID retrieves all items for a checklist
func (r *repository) GetItemsByChecklistID(ctx context.Context, checklistID string) ([]DischargeChecklistItem, error) {
	var items []DischargeChecklistItem
	err := r.db.DB.WithContext(ctx).Where("checklist_id = ?", checklistID).Order("required DESC, code ASC").Find(&items).Error
	return items, err
}

// GetIncompleteRequiredItems retrieves incomplete required items for a checklist
func (r *repository) GetIncompleteRequiredItems(ctx context.Context, checklistID string) ([]DischargeChecklistItem, error) {
	var items []DischargeChecklistItem
	err := r.db.DB.WithContext(ctx).Where("checklist_id = ? AND required = ? AND status = ?", checklistID, true, ItemStatusOpen).Find(&items).Error
	return items, err
}

// UpdateChecklistFields updates specific fields of a checklist
func (r *repository) UpdateChecklistFields(ctx context.Context, checklistID string, updates map[string]any) error {
	return r.db.DB.WithContext(ctx).Model(&DischargeChecklist{}).Where("id = ?", checklistID).Updates(updates).Error
}
