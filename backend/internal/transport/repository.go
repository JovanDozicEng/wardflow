package transport

import (
	"context"

	"github.com/wardflow/backend/pkg/database"
)

// Repository defines transport data access operations
type Repository interface {
	ListRequests(ctx context.Context, filter ListTransportFilter) ([]TransportRequest, int64, error)
	CreateRequest(ctx context.Context, req *TransportRequest) error
	GetRequestByID(ctx context.Context, id string) (*TransportRequest, error)
	UpdateRequestFields(ctx context.Context, requestID string, updates map[string]any) error
	CreateChangeEvent(ctx context.Context, event *TransportChangeEvent) error
}

type repository struct {
	db *database.DB
}

// NewRepository creates a new transport repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListRequests(ctx context.Context, filter ListTransportFilter) ([]TransportRequest, int64, error) {
	tx := r.db.DB.Model(&TransportRequest{})

	if filter.UnitID != "" {
		tx = tx.Where("unit_id = ?", filter.UnitID)
	} else if len(filter.UnitIDs) > 0 {
		tx = tx.Where("unit_id IN ?", filter.UnitIDs)
	}

	if filter.Status != "" {
		tx = tx.Where("status = ?", filter.Status)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var requests []TransportRequest
	if err := tx.Limit(filter.Limit).Offset(filter.Offset).Order("created_at DESC").Find(&requests).Error; err != nil {
		return nil, 0, err
	}

	return requests, total, nil
}

func (r *repository) CreateRequest(ctx context.Context, req *TransportRequest) error {
	return r.db.DB.Create(req).Error
}

func (r *repository) GetRequestByID(ctx context.Context, id string) (*TransportRequest, error) {
	var req TransportRequest
	if err := r.db.DB.First(&req, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *repository) UpdateRequestFields(ctx context.Context, requestID string, updates map[string]any) error {
	return r.db.DB.Model(&TransportRequest{}).Where("id = ?", requestID).Updates(updates).Error
}

func (r *repository) CreateChangeEvent(ctx context.Context, event *TransportChangeEvent) error {
	return r.db.DB.Create(event).Error
}
