package patient

import (
	"context"
	"errors"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a patient is not found
var ErrNotFound = errors.New("patient not found")

// Repository handles patient data access
type Repository struct {
	db *database.DB
}

// NewRepository creates a new patient repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new patient
func (r *Repository) Create(ctx context.Context, p *Patient) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// GetByID retrieves a patient by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*Patient, error) {
	var patient Patient
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&patient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &patient, nil
}

// List retrieves patients based on filters
func (r *Repository) List(ctx context.Context, f ListPatientsFilter) ([]*Patient, int64, error) {
	var patients []*Patient
	var total int64

	query := r.db.WithContext(ctx).Model(&Patient{})

	// Apply search filter with OR conditions on multiple fields
	if f.Q != "" {
		searchPattern := "%" + f.Q + "%"
		query = query.Where(
			"first_name ILIKE ? OR last_name ILIKE ? OR mrn ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
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

	// Order by last name, then first name
	query = query.Order("last_name ASC, first_name ASC")

	if err := query.Find(&patients).Error; err != nil {
		return nil, 0, err
	}

	return patients, total, nil
}
