package encounter

import (
	"context"
	"errors"
	"time"
)

// Service defines the interface for encounter business logic
type Service interface {
	Create(ctx context.Context, req *CreateEncounterRequest, byUserID string) (*Encounter, error)
	GetByID(ctx context.Context, id string) (*Encounter, error)
	List(ctx context.Context, f ListEncountersFilter) ([]*Encounter, int64, error)
	Update(ctx context.Context, id string, req *UpdateEncounterRequest, byUserID string) (*Encounter, error)
}

// service handles encounter business logic
type service struct {
	repo Repository
}

// NewService creates a new encounter service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Create creates a new encounter with validation
func (s *service) Create(ctx context.Context, req *CreateEncounterRequest, byUserID string) (*Encounter, error) {
	// Validate required fields
	if req.PatientID == "" {
		return nil, errors.New("patientId is required")
	}
	if req.UnitID == "" {
		return nil, errors.New("unitId is required")
	}
	if req.DepartmentID == "" {
		return nil, errors.New("departmentId is required")
	}

	// Set default startedAt if not provided
	startedAt := time.Now().UTC()
	if req.StartedAt != nil {
		startedAt = req.StartedAt.UTC()
	}

	encounter := &Encounter{
		PatientID:    req.PatientID,
		UnitID:       req.UnitID,
		DepartmentID: req.DepartmentID,
		Status:       EncounterStatusActive,
		StartedAt:    startedAt,
		CreatedBy:    byUserID,
		UpdatedBy:    byUserID,
	}

	if err := s.repo.Create(ctx, encounter); err != nil {
		return nil, err
	}

	return encounter, nil
}

// GetByID retrieves an encounter by ID
func (s *service) GetByID(ctx context.Context, id string) (*Encounter, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves encounters based on filters
func (s *service) List(ctx context.Context, f ListEncountersFilter) ([]*Encounter, int64, error) {
	return s.repo.List(ctx, f)
}

// Update updates an encounter with validation
func (s *service) Update(ctx context.Context, id string, req *UpdateEncounterRequest, byUserID string) (*Encounter, error) {
	// Fetch existing encounter
	encounter, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate status transitions
	if req.Status != nil {
		// Cannot re-activate discharged or cancelled encounters
		if encounter.Status == EncounterStatusDischarged || encounter.Status == EncounterStatusCancelled {
			if *req.Status == EncounterStatusActive {
				return nil, errors.New("cannot re-activate discharged or cancelled encounter")
			}
		}
		encounter.Status = *req.Status
	}

	if req.EndedAt != nil {
		endedAt := req.EndedAt.UTC()
		encounter.EndedAt = &endedAt
	}

	if req.UnitID != nil {
		encounter.UnitID = *req.UnitID
	}

	encounter.UpdatedBy = byUserID

	if err := s.repo.Update(ctx, encounter); err != nil {
		return nil, err
	}

	return encounter, nil
}
