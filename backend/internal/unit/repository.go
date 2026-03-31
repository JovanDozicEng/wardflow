package unit

import (
	"context"

	"github.com/wardflow/backend/pkg/database"
)

// Repository defines unit data access operations
type Repository interface {
	List(ctx context.Context, q, departmentID string) ([]Unit, error)
	Create(ctx context.Context, unit *Unit) error
	GetByID(ctx context.Context, id string) (*Unit, error)
}

type repository struct {
	db *database.DB
}

// NewRepository creates a new unit repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context, q, departmentID string) ([]Unit, error) {
	var units []Unit
	tx := r.db.WithContext(ctx).Order("name ASC")

	if q != "" {
		searchPattern := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR code ILIKE ?", searchPattern, searchPattern)
	}

	if departmentID != "" {
		tx = tx.Where("department_id = ?", departmentID)
	}

	if err := tx.Find(&units).Error; err != nil {
		return nil, err
	}

	return units, nil
}

func (r *repository) Create(ctx context.Context, unit *Unit) error {
	return r.db.WithContext(ctx).Create(unit).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*Unit, error) {
	var unit Unit
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&unit).Error; err != nil {
		return nil, err
	}
	return &unit, nil
}
