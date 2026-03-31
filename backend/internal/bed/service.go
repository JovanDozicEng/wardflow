package bed

import (
	"context"
	"time"

	"github.com/wardflow/backend/pkg/database"
)

// Service defines bed business logic operations
type Service interface {
	ListBeds(ctx context.Context, filter ListBedsFilter) ([]Bed, int64, error)
	CreateBed(ctx context.Context, req CreateBedRequest, userID string) (*Bed, error)
	GetBed(ctx context.Context, id string) (*Bed, error)
	UpdateBedStatus(ctx context.Context, bedID string, req UpdateBedStatusRequest, userID string) (*BedStatusEvent, error)
	CreateBedRequest(ctx context.Context, encounterID, userID string, req CreateBedRequestRequest) (*BedRequest, error)
	AssignBed(ctx context.Context, requestID string, req AssignBedRequest, userID string) (*BedRequest, error)
}

type service struct {
	repo Repository
	db   *database.DB
}

// NewService creates a new bed service
func NewService(repo Repository, db *database.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

func (s *service) ListBeds(ctx context.Context, filter ListBedsFilter) ([]Bed, int64, error) {
	return s.repo.ListBeds(ctx, filter)
}

func (s *service) CreateBed(ctx context.Context, req CreateBedRequest, userID string) (*Bed, error) {
	if req.UnitID == "" || req.Room == "" || req.Label == "" {
		return nil, ErrValidation("unitId, room, and label are required")
	}

	bed := Bed{
		UnitID:        req.UnitID,
		Room:          req.Room,
		Label:         req.Label,
		Capabilities:  StringSlice(req.Capabilities),
		CurrentStatus: BedStatusAvailable,
	}

	if err := s.repo.CreateBed(ctx, &bed); err != nil {
		return nil, err
	}

	return &bed, nil
}

func (s *service) GetBed(ctx context.Context, id string) (*Bed, error) {
	return s.repo.GetBedByID(ctx, id)
}

func (s *service) UpdateBedStatus(ctx context.Context, bedID string, req UpdateBedStatusRequest, userID string) (*BedStatusEvent, error) {
	if req.Status == "" {
		return nil, ErrValidation("status is required")
	}

	// Fetch current bed to get fromStatus
	bed, err := s.repo.GetBedByID(ctx, bedID)
	if err != nil {
		return nil, err
	}

	fromStatus := bed.CurrentStatus
	event := BedStatusEvent{
		BedID:      bedID,
		FromStatus: &fromStatus,
		ToStatus:   req.Status,
		Reason:     req.Reason,
		ChangedBy:  userID,
		ChangedAt:  time.Now().UTC(),
	}

	if err := s.repo.CreateBedStatusEvent(ctx, &event); err != nil {
		return nil, err
	}

	updates := map[string]any{"current_status": req.Status}
	if req.Status != BedStatusOccupied {
		updates["current_encounter_id"] = nil
	}

	if err := s.repo.UpdateBedFields(ctx, bedID, updates); err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *service) CreateBedRequest(ctx context.Context, encounterID, userID string, req CreateBedRequestRequest) (*BedRequest, error) {
	priority := req.Priority
	if priority == "" {
		priority = "routine"
	}

	bedReq := BedRequest{
		EncounterID:          encounterID,
		RequiredCapabilities: StringSlice(req.RequiredCapabilities),
		Priority:             priority,
		Status:               BedRequestStatusPending,
		CreatedBy:            userID,
	}

	if err := s.repo.CreateBedRequest(ctx, &bedReq); err != nil {
		return nil, err
	}

	return &bedReq, nil
}

func (s *service) AssignBed(ctx context.Context, requestID string, req AssignBedRequest, userID string) (*BedRequest, error) {
	if req.BedID == "" {
		return nil, ErrValidation("bedId is required")
	}

	// Fetch bed request
	bedReq, err := s.repo.GetBedRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if bedReq.Status != BedRequestStatusPending {
		return nil, ErrInvalidState("bed request is not pending")
	}

	// Check bed is available
	bed, err := s.repo.GetBedByID(ctx, req.BedID)
	if err != nil {
		return nil, err
	}
	if bed.CurrentStatus != BedStatusAvailable {
		return nil, ErrInvalidState("bed is not available for assignment")
	}

	// Assign bed atomically
	if err := s.repo.AssignBed(ctx, requestID, req.BedID, bedReq.EncounterID, userID, bed.CurrentStatus); err != nil {
		return nil, err
	}

	// Fetch updated request
	updatedReq, err := s.repo.GetBedRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	return updatedReq, nil
}

// Error helpers
func ErrValidation(msg string) error {
	return &ValidationError{Message: msg}
}

func ErrInvalidState(msg string) error {
	return &InvalidStateError{Message: msg}
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type InvalidStateError struct {
	Message string
}

func (e *InvalidStateError) Error() string {
	return e.Message
}
