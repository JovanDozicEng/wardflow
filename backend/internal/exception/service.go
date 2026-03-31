package exception

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Service defines the interface for exception event business logic
type Service interface {
	Create(ctx context.Context, req *CreateExceptionRequest, byUserID string) (*ExceptionEvent, error)
	Update(ctx context.Context, id string, req *UpdateExceptionRequest, byUserID string) (*ExceptionEvent, error)
	Finalize(ctx context.Context, id string, byUserID string) (*ExceptionEvent, error)
	Correct(ctx context.Context, id string, req *CorrectExceptionRequest, byUserID string) (*ExceptionEvent, error)
	List(ctx context.Context, f ListExceptionsFilter) ([]*ExceptionEvent, int64, error)
	GetByID(ctx context.Context, id string) (*ExceptionEvent, error)
}

// service handles exception event business logic
type service struct {
	repo Repository
}

// NewService creates a new exception event service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Create creates a new exception event with validation
func (s *service) Create(ctx context.Context, req *CreateExceptionRequest, byUserID string) (*ExceptionEvent, error) {
	// Validate required fields
	if req.EncounterID == "" {
		return nil, errors.New("encounterId is required")
	}
	if req.Type == "" {
		return nil, errors.New("type is required")
	}

	// Marshal data to JSON
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		return nil, errors.New("invalid data format")
	}

	// Empty requiredFields as JSON object
	requiredFieldsJSON := "{}"

	exception := &ExceptionEvent{
		EncounterID:    req.EncounterID,
		Type:           req.Type,
		Status:         ExceptionStatusDraft,
		RequiredFields: requiredFieldsJSON,
		Data:           string(dataJSON),
		InitiatedBy:    byUserID,
		InitiatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, exception); err != nil {
		return nil, err
	}

	return exception, nil
}

// Update updates a draft exception event
func (s *service) Update(ctx context.Context, id string, req *UpdateExceptionRequest, byUserID string) (*ExceptionEvent, error) {
	exception, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only draft exceptions can be updated
	if exception.Status != ExceptionStatusDraft {
		return nil, errors.New("only draft exceptions can be updated")
	}

	// Marshal new data to JSON
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		return nil, errors.New("invalid data format")
	}

	exception.Data = string(dataJSON)

	if err := s.repo.Update(ctx, exception); err != nil {
		return nil, err
	}

	return exception, nil
}

// Finalize finalizes a draft exception event
func (s *service) Finalize(ctx context.Context, id string, byUserID string) (*ExceptionEvent, error) {
	exception, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only draft exceptions can be finalized
	if exception.Status != ExceptionStatusDraft {
		return nil, errors.New("only draft exceptions can be finalized")
	}

	now := time.Now().UTC()
	exception.Status = ExceptionStatusFinalized
	exception.FinalizedBy = &byUserID
	exception.FinalizedAt = &now

	if err := s.repo.Update(ctx, exception); err != nil {
		return nil, err
	}

	return exception, nil
}

// Correct creates a corrected version of a finalized exception event (IMMUTABILITY PATTERN)
func (s *service) Correct(ctx context.Context, id string, req *CorrectExceptionRequest, byUserID string) (*ExceptionEvent, error) {
	originalException, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only finalized exceptions can be corrected
	if originalException.Status != ExceptionStatusFinalized {
		return nil, errors.New("only finalized exceptions can be corrected")
	}

	// Validate required fields
	if req.Reason == "" {
		return nil, errors.New("reason is required")
	}

	// Marshal new data to JSON
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		return nil, errors.New("invalid data format")
	}

	// Step 1: Create NEW exception event with status=finalized
	now := time.Now().UTC()
	newException := &ExceptionEvent{
		EncounterID:    originalException.EncounterID,
		Type:           originalException.Type,
		Status:         ExceptionStatusFinalized,
		RequiredFields: originalException.RequiredFields,
		Data:           string(dataJSON),
		InitiatedBy:    originalException.InitiatedBy,
		InitiatedAt:    originalException.InitiatedAt,
		FinalizedBy:    &byUserID,
		FinalizedAt:    &now,
	}

	if err := s.repo.Create(ctx, newException); err != nil {
		return nil, err
	}

	// Step 2: Update original exception to mark as corrected
	originalException.Status = ExceptionStatusCorrected
	originalException.CorrectedByEventID = &newException.ID
	originalException.CorrectionReason = &req.Reason

	if err := s.repo.Update(ctx, originalException); err != nil {
		return nil, err
	}

	// Step 3: Return the NEW event
	return newException, nil
}

// List retrieves exception events based on filters
func (s *service) List(ctx context.Context, f ListExceptionsFilter) ([]*ExceptionEvent, int64, error) {
	return s.repo.List(ctx, f)
}

// GetByID retrieves an exception event by ID
func (s *service) GetByID(ctx context.Context, id string) (*ExceptionEvent, error) {
	return s.repo.GetByID(ctx, id)
}
