package bed

import (
	"context"
	"fmt"
	"time"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// Repository defines bed data access operations
type Repository interface {
	ListBeds(ctx context.Context, filter ListBedsFilter) ([]Bed, int64, error)
	CreateBed(ctx context.Context, bed *Bed) error
	GetBedByID(ctx context.Context, id string) (*Bed, error)
	CreateBedStatusEvent(ctx context.Context, event *BedStatusEvent) error
	UpdateBedFields(ctx context.Context, bedID string, updates map[string]any) error
	CreateBedRequest(ctx context.Context, req *BedRequest) error
	GetBedRequestByID(ctx context.Context, id string) (*BedRequest, error)
	UpdateBedRequestFields(ctx context.Context, requestID string, updates map[string]any) error
	AssignBed(ctx context.Context, requestID, bedID, encounterID, userID string, fromStatus BedStatus) error
}

type repository struct {
	db *database.DB
}

// NewRepository creates a new bed repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListBeds(ctx context.Context, filter ListBedsFilter) ([]Bed, int64, error) {
	tx := r.db.DB.Model(&Bed{})

	if filter.UnitID != "" {
		tx = tx.Where("unit_id = ?", filter.UnitID)
	} else if len(filter.UnitIDs) > 0 {
		tx = tx.Where("unit_id IN ?", filter.UnitIDs)
	}

	if filter.Status != "" {
		tx = tx.Where("current_status = ?", filter.Status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var beds []Bed
	if err := tx.Limit(filter.Limit).Offset(filter.Offset).Order("room ASC, label ASC").Find(&beds).Error; err != nil {
		return nil, 0, err
	}

	return beds, total, nil
}

func (r *repository) CreateBed(ctx context.Context, bed *Bed) error {
	return r.db.DB.Create(bed).Error
}

func (r *repository) GetBedByID(ctx context.Context, id string) (*Bed, error) {
	var bed Bed
	if err := r.db.DB.First(&bed, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &bed, nil
}

func (r *repository) CreateBedStatusEvent(ctx context.Context, event *BedStatusEvent) error {
	return r.db.DB.Create(event).Error
}

func (r *repository) UpdateBedFields(ctx context.Context, bedID string, updates map[string]any) error {
	return r.db.DB.Model(&Bed{}).Where("id = ?", bedID).Updates(updates).Error
}

func (r *repository) CreateBedRequest(ctx context.Context, req *BedRequest) error {
	return r.db.DB.Create(req).Error
}

func (r *repository) GetBedRequestByID(ctx context.Context, id string) (*BedRequest, error) {
	var req BedRequest
	if err := r.db.DB.First(&req, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *repository) UpdateBedRequestFields(ctx context.Context, requestID string, updates map[string]any) error {
	return r.db.DB.Model(&BedRequest{}).Where("id = ?", requestID).Updates(updates).Error
}

// AssignBed atomically assigns a bed to a request with transaction isolation
func (r *repository) AssignBed(ctx context.Context, requestID, bedID, encounterID, userID string, fromStatus BedStatus) error {
	return r.db.DB.Transaction(func(tx *gorm.DB) error {
		// Re-fetch bed with lock to prevent race condition
		var lockedBed Bed
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&lockedBed, "id = ?", bedID).Error; err != nil {
			return err
		}
		if lockedBed.CurrentStatus != BedStatusAvailable {
			return fmt.Errorf("bed is no longer available")
		}

		// Create status event
		now := time.Now().UTC()
		event := BedStatusEvent{
			BedID:      bedID,
			FromStatus: &fromStatus,
			ToStatus:   BedStatusOccupied,
			ChangedBy:  userID,
			ChangedAt:  now,
		}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}

		// Update bed status and encounter
		if err := tx.Model(&lockedBed).Updates(map[string]any{
			"current_status":       BedStatusOccupied,
			"current_encounter_id": encounterID,
		}).Error; err != nil {
			return err
		}

		// Update bed request status
		if err := tx.Model(&BedRequest{}).Where("id = ?", requestID).Updates(map[string]any{
			"status":          BedRequestStatusAssigned,
			"assigned_bed_id": bedID,
		}).Error; err != nil {
			return err
		}

		return nil
	})
}
