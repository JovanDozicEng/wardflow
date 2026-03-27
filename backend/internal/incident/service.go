package incident

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// Service handles incident business logic
type Service struct {
	repo *Repository
}

// NewService creates a new incident service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new incident with validation
func (s *Service) Create(ctx context.Context, req *CreateIncidentRequest, byUserID string) (*Incident, error) {
	// Validate required fields
	if req.Type == "" {
		return nil, errors.New("type is required")
	}
	if req.EventTime.IsZero() {
		return nil, errors.New("eventTime is required")
	}

	// Marshal harm indicators to JSON if provided
	var harmIndicatorsJSON *string
	if req.HarmIndicators != nil && len(req.HarmIndicators) > 0 {
		data, err := json.Marshal(req.HarmIndicators)
		if err != nil {
			return nil, errors.New("invalid harmIndicators format")
		}
		jsonStr := string(data)
		harmIndicatorsJSON = &jsonStr
	}

	now := time.Now().UTC()
	incident := &Incident{
		EncounterID:    req.EncounterID,
		Type:           req.Type,
		Severity:       req.Severity,
		HarmIndicators: harmIndicatorsJSON,
		EventTime:      req.EventTime.UTC(),
		ReportedBy:     byUserID,
		ReportedAt:     now,
		Status:         IncidentStatusSubmitted,
	}

	if err := s.repo.Create(ctx, incident); err != nil {
		return nil, err
	}

	return incident, nil
}

// GetByID retrieves an incident by ID
func (s *Service) GetByID(ctx context.Context, id string) (*Incident, error) {
	return s.repo.GetByID(ctx, id)
}

// List retrieves incidents based on filters
func (s *Service) List(ctx context.Context, f ListIncidentsFilter) ([]*Incident, int64, error) {
	return s.repo.List(ctx, f)
}

// UpdateStatus updates the status of an incident and creates a status event
func (s *Service) UpdateStatus(ctx context.Context, id string, req *UpdateIncidentStatusRequest, byUserID string) (*Incident, error) {
	// Get existing incident
	incident, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate status is provided
	if req.Status == "" {
		return nil, errors.New("status is required")
	}

	// Create status event
	now := time.Now().UTC()
	fromStatus := incident.Status
	statusEvent := &IncidentStatusEvent{
		IncidentID: incident.ID,
		FromStatus: &fromStatus,
		ToStatus:   req.Status,
		ChangedBy:  byUserID,
		ChangedAt:  now,
		Note:       req.Note,
	}

	if err := s.repo.CreateStatusEvent(ctx, statusEvent); err != nil {
		return nil, err
	}

	// Update incident status
	incident.Status = req.Status

	if err := s.repo.Update(ctx, incident); err != nil {
		return nil, err
	}

	return incident, nil
}

// GetStatusHistory retrieves the status history for an incident
func (s *Service) GetStatusHistory(ctx context.Context, incidentID string) ([]*IncidentStatusEvent, error) {
	// Verify incident exists
	_, err := s.repo.GetByID(ctx, incidentID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetStatusHistory(ctx, incidentID)
}
