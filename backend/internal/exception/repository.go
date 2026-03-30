package exception

import (
	"context"
	"errors"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// ErrNotFound is returned when an exception event is not found
var ErrNotFound = errors.New("exception event not found")

// Repository handles exception event data access
type Repository struct {
	db *database.DB
}

// NewRepository creates a new exception event repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new exception event
func (r *Repository) Create(ctx context.Context, e *ExceptionEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// GetByID retrieves an exception event by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*ExceptionEvent, error) {
	var exception ExceptionEvent
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&exception).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &exception, nil
}

// List retrieves exception events based on filters
func (r *Repository) List(ctx context.Context, f ListExceptionsFilter) ([]*ExceptionEvent, int64, error) {
	var exceptions []*ExceptionEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&ExceptionEvent{})

	// Apply filters
	if f.EncounterID != "" {
		query = query.Where("encounter_id = ?", f.EncounterID)
	}
	if f.Type != "" {
		query = query.Where("type = ?", f.Type)
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

	if err := query.Find(&exceptions).Error; err != nil {
		return nil, 0, err
	}

	return exceptions, total, nil
}

// Update updates an exception event
func (r *Repository) Update(ctx context.Context, e *ExceptionEvent) error {
	result := r.db.WithContext(ctx).Save(e)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
