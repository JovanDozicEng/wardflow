package consult

import (
	"context"
	"errors"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a consult request is not found
var ErrNotFound = errors.New("consult request not found")

// Repository handles consult request data access
type Repository struct {
	db *database.DB
}

// NewRepository creates a new consult request repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new consult request
func (r *Repository) Create(ctx context.Context, c *ConsultRequest) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// GetByID retrieves a consult request by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*ConsultRequest, error) {
	var consult ConsultRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&consult).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &consult, nil
}

// List retrieves consult requests based on filters
func (r *Repository) List(ctx context.Context, f ListConsultsFilter) ([]*ConsultRequest, int64, error) {
	var consults []*ConsultRequest
	var total int64

	query := r.db.WithContext(ctx).Model(&ConsultRequest{})

	// Apply filters
	if f.UnitID != "" {
		query = query.Where("unit_id = ?", f.UnitID)
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	if f.TargetService != "" {
		query = query.Where("target_service = ?", f.TargetService)
	}

	// Count total matching records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if f.Limit > 0 {
		query = query.Limit(f.Limit)
	}
	if f.Offset > 0 {
		query = query.Offset(f.Offset)
	}

	// Order by created_at descending
	query = query.Order("created_at DESC")

	if err := query.Find(&consults).Error; err != nil {
		return nil, 0, err
	}

	return consults, total, nil
}

// Update updates a consult request
func (r *Repository) Update(ctx context.Context, c *ConsultRequest) error {
	result := r.db.WithContext(ctx).Save(c)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
