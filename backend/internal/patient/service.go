package patient

import (
	"context"
	"errors"
	"time"
)

// Service handles patient business logic
type Service struct {
	repo *Repository
}

// NewService creates a new patient service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new patient with validation
func (s *Service) Create(ctx context.Context, req *CreatePatientRequest, byUserID string) (*Patient, error) {
	// Validate required fields
	if req.FirstName == "" {
		return nil, errors.New("firstName is required")
	}
	if req.LastName == "" {
		return nil, errors.New("lastName is required")
	}
	if req.MRN == "" {
		return nil, errors.New("mrn is required")
	}

	patient := &Patient{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		MRN:       req.MRN,
		CreatedBy: byUserID,
	}

	// Parse date of birth if provided
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, errors.New("dateOfBirth must be in ISO format (YYYY-MM-DD)")
		}
		patient.DateOfBirth = &dob
	}

	if err := s.repo.Create(ctx, patient); err != nil {
		return nil, err
	}

	return patient, nil
}

// GetByID retrieves a patient by ID
func (s *Service) GetByID(ctx context.Context, id string) (*Patient, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves patients based on filters
func (s *Service) List(ctx context.Context, f ListPatientsFilter) ([]*Patient, int64, error) {
	return s.repo.List(ctx, f)
}
