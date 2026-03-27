package encounter

import (
	"context"
	"errors"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// ErrNotFound is returned when an encounter is not found
var ErrNotFound = errors.New("encounter not found")

// Repository handles encounter data access
type Repository struct {
	db *database.DB
}

// NewRepository creates a new encounter repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new encounter
func (r *Repository) Create(ctx context.Context, e *Encounter) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// GetByID retrieves an encounter by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Encounter, error) {
	var encounter Encounter
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&encounter).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &encounter, nil
}

// List retrieves encounters based on filters
func (r *Repository) List(ctx context.Context, f ListEncountersFilter) ([]*Encounter, int64, error) {
	var encounters []*Encounter
	var total int64

	query := r.db.WithContext(ctx).Model(&Encounter{})

	// Apply filters
	if f.UnitID != "" {
		query = query.Where("unit_id = ?", f.UnitID)
	}
	if f.DepartmentID != "" {
		query = query.Where("department_id = ?", f.DepartmentID)
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
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

	if err := query.Find(&encounters).Error; err != nil {
		return nil, 0, err
	}

	return encounters, total, nil
}

// Update updates an encounter
func (r *Repository) Update(ctx context.Context, e *Encounter) error {
	result := r.db.WithContext(ctx).Save(e)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
