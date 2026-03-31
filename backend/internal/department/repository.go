package department

import (
	"context"

	"github.com/wardflow/backend/pkg/database"
)

// Repository defines department data access operations
type Repository interface {
	List(ctx context.Context, q string) ([]Department, error)
	Create(ctx context.Context, dept *Department) error
	GetByID(ctx context.Context, id string) (*Department, error)
}

type repository struct {
	db *database.DB
}

// NewRepository creates a new department repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context, q string) ([]Department, error) {
	var departments []Department
	tx := r.db.WithContext(ctx).Order("name ASC")

	if q != "" {
		searchPattern := "%" + q + "%"
		tx = tx.Where("name ILIKE ? OR code ILIKE ?", searchPattern, searchPattern)
	}

	if err := tx.Find(&departments).Error; err != nil {
		return nil, err
	}

	return departments, nil
}

func (r *repository) Create(ctx context.Context, dept *Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*Department, error) {
	var dept Department
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&dept).Error; err != nil {
		return nil, err
	}
	return &dept, nil
}
