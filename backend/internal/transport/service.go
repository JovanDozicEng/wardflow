package transport

import (
	"context"
	"encoding/json"
	"time"

	"github.com/wardflow/backend/pkg/database"
)

// Service defines transport business logic operations
type Service interface {
	ListRequests(ctx context.Context, filter ListTransportFilter) ([]TransportRequest, int64, error)
	CreateRequest(ctx context.Context, req CreateTransportRequest, userID string) (*TransportRequest, error)
	GetRequest(ctx context.Context, id string) (*TransportRequest, error)
	AcceptRequest(ctx context.Context, requestID string, req AcceptTransportRequest, userID string) (*TransportRequest, error)
	UpdateRequest(ctx context.Context, requestID string, req UpdateTransportRequest, userID string) (*TransportRequest, error)
	CompleteRequest(ctx context.Context, requestID, userID string) (*TransportRequest, error)
}

type service struct {
	repo Repository
	db   *database.DB
}

// NewService creates a new transport service
func NewService(repo Repository, db *database.DB) Service {
	return &service{
		repo: repo,
		db:   db,
	}
}

func (s *service) ListRequests(ctx context.Context, filter ListTransportFilter) ([]TransportRequest, int64, error) {
	return s.repo.ListRequests(ctx, filter)
}

func (s *service) CreateRequest(ctx context.Context, req CreateTransportRequest, userID string) (*TransportRequest, error) {
	if req.EncounterID == "" || req.Origin == "" || req.Destination == "" {
		return nil, ErrValidation("encounterId, origin, and destination are required")
	}

	priority := req.Priority
	if priority == "" {
		priority = "routine"
	}

	tr := TransportRequest{
		EncounterID: req.EncounterID,
		Origin:      req.Origin,
		Destination: req.Destination,
		Priority:    priority,
		Status:      TransportStatusPending,
		CreatedBy:   userID,
	}

	if err := s.repo.CreateRequest(ctx, &tr); err != nil {
		return nil, err
	}

	return &tr, nil
}

func (s *service) GetRequest(ctx context.Context, id string) (*TransportRequest, error) {
	return s.repo.GetRequestByID(ctx, id)
}

func (s *service) AcceptRequest(ctx context.Context, requestID string, req AcceptTransportRequest, userID string) (*TransportRequest, error) {
	tr, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if tr.Status != TransportStatusPending {
		return nil, ErrInvalidState("transport request is not pending")
	}

	assignedTo := req.AssignedTo
	if assignedTo == "" {
		assignedTo = userID
	}
	now := time.Now().UTC()

	if err := s.repo.UpdateRequestFields(ctx, requestID, map[string]any{
		"status":      TransportStatusAssigned,
		"assigned_to": assignedTo,
		"assigned_at": now,
	}); err != nil {
		return nil, err
	}

	// Create change event
	changedFields, _ := json.Marshal(map[string]any{
		"status":     TransportStatusAssigned,
		"assignedTo": assignedTo,
	})
	_ = s.repo.CreateChangeEvent(ctx, &TransportChangeEvent{
		RequestID:     requestID,
		ChangedFields: string(changedFields),
		ChangedBy:     userID,
		ChangedAt:     now,
	})

	// Fetch updated request
	return s.repo.GetRequestByID(ctx, requestID)
}

func (s *service) UpdateRequest(ctx context.Context, requestID string, req UpdateTransportRequest, userID string) (*TransportRequest, error) {
	tr, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if tr.Status == TransportStatusCompleted || tr.Status == TransportStatusCancelled {
		return nil, ErrInvalidState("cannot update a completed or cancelled request")
	}

	updates := map[string]any{}
	changedFields := map[string]any{}

	if req.Origin != nil {
		updates["origin"] = *req.Origin
		changedFields["origin"] = *req.Origin
	}
	if req.Destination != nil {
		updates["destination"] = *req.Destination
		changedFields["destination"] = *req.Destination
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
		changedFields["priority"] = *req.Priority
	}

	if len(updates) == 0 {
		return nil, ErrValidation("no fields to update")
	}

	if err := s.repo.UpdateRequestFields(ctx, requestID, updates); err != nil {
		return nil, err
	}

	// Create change event
	now := time.Now().UTC()
	cf, _ := json.Marshal(changedFields)
	_ = s.repo.CreateChangeEvent(ctx, &TransportChangeEvent{
		RequestID:     requestID,
		ChangedFields: string(cf),
		ChangedBy:     userID,
		Reason:        req.Reason,
		ChangedAt:     now,
	})

	// Fetch updated request
	return s.repo.GetRequestByID(ctx, requestID)
}

func (s *service) CompleteRequest(ctx context.Context, requestID, userID string) (*TransportRequest, error) {
	tr, err := s.repo.GetRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if tr.Status != TransportStatusAssigned {
		return nil, ErrInvalidState("transport request must be accepted before completing")
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateRequestFields(ctx, requestID, map[string]any{"status": TransportStatusCompleted}); err != nil {
		return nil, err
	}

	// Create change event
	cf, _ := json.Marshal(map[string]any{"status": TransportStatusCompleted})
	_ = s.repo.CreateChangeEvent(ctx, &TransportChangeEvent{
		RequestID:     requestID,
		ChangedFields: string(cf),
		ChangedBy:     userID,
		ChangedAt:     now,
	})

	// Fetch updated request
	return s.repo.GetRequestByID(ctx, requestID)
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
