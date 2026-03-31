package incident

import (
	"context"
	"errors"

	"github.com/wardflow/backend/pkg/database"
	"gorm.io/gorm"
)

// ErrNotFound is returned when an incident is not found
var ErrNotFound = errors.New("incident not found")

// Repository defines the interface for incident data access
type Repository interface {
	Create(ctx context.Context, i *Incident) error
	GetByID(ctx context.Context, id string) (*Incident, error)
	List(ctx context.Context, f ListIncidentsFilter) ([]*Incident, int64, error)
	Update(ctx context.Context, i *Incident) error
	CreateStatusEvent(ctx context.Context, e *IncidentStatusEvent) error
	GetStatusHistory(ctx context.Context, incidentID string) ([]*IncidentStatusEvent, error)
}

// repository handles incident data access
type repository struct {
	db *database.DB
}

// NewRepository creates a new incident repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

// Create creates a new incident
func (r *repository) Create(ctx context.Context, i *Incident) error {
	return r.db.WithContext(ctx).Create(i).Error
}

// GetByID retrieves an incident by ID
func (r *repository) GetByID(ctx context.Context, id string) (*Incident, error) {
	var incident Incident
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&incident).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &incident, nil
}

// List retrieves incidents based on filters
func (r *repository) List(ctx context.Context, f ListIncidentsFilter) ([]*Incident, int64, error) {
	var incidents []*Incident
	var total int64

	query := r.db.WithContext(ctx).Model(&Incident{})

	// Apply filters
	if f.UnitID != "" {
		query = query.Where("unit_id = ?", f.UnitID)
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	if f.Type != "" {
		query = query.Where("type = ?", f.Type)
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

	// Order by event_time descending
	query = query.Order("event_time DESC")

	if err := query.Find(&incidents).Error; err != nil {
		return nil, 0, err
	}

	return incidents, total, nil
}

// Update updates an incident
func (r *repository) Update(ctx context.Context, i *Incident) error {
	result := r.db.WithContext(ctx).Save(i)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateStatusEvent creates an incident status event
func (r *repository) CreateStatusEvent(ctx context.Context, e *IncidentStatusEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

// GetStatusHistory retrieves all status events for an incident
func (r *repository) GetStatusHistory(ctx context.Context, incidentID string) ([]*IncidentStatusEvent, error) {
	var events []*IncidentStatusEvent
	err := r.db.WithContext(ctx).
		Where("incident_id = ?", incidentID).
		Order("changed_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}
