package consult

import (
	"context"
	"errors"
	"time"
)

// Service handles consult request business logic
type Service struct {
	repo *Repository
}

// NewService creates a new consult request service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new consult request with validation
func (s *Service) Create(ctx context.Context, req *CreateConsultRequest, byUserID string) (*ConsultRequest, error) {
	// Validate required fields
	if req.EncounterID == "" {
		return nil, errors.New("encounterId is required")
	}
	if req.TargetService == "" {
		return nil, errors.New("targetService is required")
	}
	if req.Reason == "" {
		return nil, errors.New("reason is required")
	}

	consult := &ConsultRequest{
		EncounterID:   req.EncounterID,
		TargetService: req.TargetService,
		Reason:        req.Reason,
		Urgency:       req.Urgency,
		Status:        ConsultStatusPending,
		CreatedBy:     byUserID,
	}

	if err := s.repo.Create(ctx, consult); err != nil {
		return nil, err
	}

	return consult, nil
}

// Accept accepts a pending consult request
func (s *Service) Accept(ctx context.Context, id string, byUserID string) (*ConsultRequest, error) {
	consult, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only pending consults can be accepted
	if consult.Status != ConsultStatusPending {
		return nil, errors.New("only pending consults can be accepted")
	}

	now := time.Now().UTC()
	consult.Status = ConsultStatusAccepted
	consult.AcceptedBy = &byUserID
	consult.AcceptedAt = &now

	if err := s.repo.Update(ctx, consult); err != nil {
		return nil, err
	}

	return consult, nil
}

// Decline declines a pending consult request
func (s *Service) Decline(ctx context.Context, id string, req *DeclineConsultRequest, byUserID string) (*ConsultRequest, error) {
	consult, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only pending consults can be declined
	if consult.Status != ConsultStatusPending {
		return nil, errors.New("only pending consults can be declined")
	}

	// Reason is required
	if req.Reason == "" {
		return nil, errors.New("reason is required")
	}

	now := time.Now().UTC()
	consult.Status = ConsultStatusDeclined
	consult.CloseReason = &req.Reason
	consult.ClosedAt = &now

	if err := s.repo.Update(ctx, consult); err != nil {
		return nil, err
	}

	return consult, nil
}

// Redirect redirects a pending consult request to another service
func (s *Service) Redirect(ctx context.Context, id string, req *RedirectConsultRequest, byUserID string) (*ConsultRequest, error) {
	consult, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only pending consults can be redirected
	if consult.Status != ConsultStatusPending {
		return nil, errors.New("only pending consults can be redirected")
	}

	// Reason and new target service are required
	if req.Reason == "" {
		return nil, errors.New("reason is required")
	}
	if req.TargetService == "" {
		return nil, errors.New("targetService is required")
	}

	now := time.Now().UTC()
	consult.Status = ConsultStatusRedirected
	consult.RedirectedTo = &req.TargetService
	consult.CloseReason = &req.Reason
	consult.ClosedAt = &now

	if err := s.repo.Update(ctx, consult); err != nil {
		return nil, err
	}

	return consult, nil
}

// Complete marks an accepted consult request as completed
func (s *Service) Complete(ctx context.Context, id string, byUserID string) (*ConsultRequest, error) {
	consult, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only accepted consults can be completed
	if consult.Status != ConsultStatusAccepted {
		return nil, errors.New("only accepted consults can be completed")
	}

	now := time.Now().UTC()
	consult.Status = ConsultStatusCompleted
	consult.ClosedAt = &now

	if err := s.repo.Update(ctx, consult); err != nil {
		return nil, err
	}

	return consult, nil
}

// List retrieves consult requests based on filters
func (s *Service) List(ctx context.Context, f ListConsultsFilter) ([]*ConsultRequest, int64, error) {
	return s.repo.List(ctx, f)
}

// GetByID retrieves a consult request by ID
func (s *Service) GetByID(ctx context.Context, id string) (*ConsultRequest, error) {
	return s.repo.GetByID(ctx, id)
}
