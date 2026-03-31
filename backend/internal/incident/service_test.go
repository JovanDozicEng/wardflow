package incident

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepository is a mock implementation of Repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, i *Incident) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Incident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Incident), args.Error(1)
}

func (m *mockRepository) List(ctx context.Context, f ListIncidentsFilter) ([]*Incident, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Incident), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) Update(ctx context.Context, i *Incident) error {
	args := m.Called(ctx, i)
	return args.Error(0)
}

func (m *mockRepository) CreateStatusEvent(ctx context.Context, e *IncidentStatusEvent) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *mockRepository) GetStatusHistory(ctx context.Context, incidentID string) ([]*IncidentStatusEvent, error) {
	args := m.Called(ctx, incidentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*IncidentStatusEvent), args.Error(1)
}

func TestService_Create_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	eventTime := time.Now().UTC()
	req := &CreateIncidentRequest{
		Type:      "fall",
		EventTime: eventTime,
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(i *Incident) bool {
		return i.Type == "fall" && i.Status == IncidentStatusSubmitted
	})).Run(func(args mock.Arguments) {
		// Simulate database auto-generating ID
		inc := args.Get(1).(*Incident)
		inc.ID = "inc-123"
	}).Return(nil)

	incident, err := svc.Create(context.Background(), req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, incident)
	assert.Equal(t, "inc-123", incident.ID)
	assert.Equal(t, "fall", incident.Type)
	assert.Equal(t, IncidentStatusSubmitted, incident.Status)
	assert.Equal(t, "user-1", incident.ReportedBy)
	assert.True(t, incident.EventTime.Equal(eventTime))
	repo.AssertExpectations(t)
}

func TestService_Create_WithHarmIndicators(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	eventTime := time.Now().UTC()
	severity := "major"
	req := &CreateIncidentRequest{
		Type:      "medication_error",
		EventTime: eventTime,
		Severity:  &severity,
		HarmIndicators: map[string]interface{}{
			"patientImpact": "moderate",
			"intervention":  "required",
		},
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(i *Incident) bool {
		return i.Type == "medication_error" && i.HarmIndicators != nil
	})).Run(func(args mock.Arguments) {
		// Simulate database auto-generating ID
		inc := args.Get(1).(*Incident)
		inc.ID = "inc-124"
	}).Return(nil)

	incident, err := svc.Create(context.Background(), req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, incident)
	assert.Equal(t, "inc-124", incident.ID)
	assert.NotNil(t, incident.HarmIndicators)
	assert.NotNil(t, incident.Severity)
	assert.Equal(t, "major", *incident.Severity)
	repo.AssertExpectations(t)
}

func TestService_Create_MissingType(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &CreateIncidentRequest{
		EventTime: time.Now().UTC(),
	}

	incident, err := svc.Create(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, incident)
	assert.Equal(t, "type is required", err.Error())
}

func TestService_Create_MissingEventTime(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &CreateIncidentRequest{
		Type: "fall",
	}

	incident, err := svc.Create(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, incident)
	assert.Equal(t, "eventTime is required", err.Error())
}

func TestService_Create_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &CreateIncidentRequest{
		Type:      "fall",
		EventTime: time.Now().UTC(),
	}

	dbErr := errors.New("database error")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*incident.Incident")).Return(dbErr)

	incident, err := svc.Create(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, incident)
	assert.Equal(t, dbErr, err)
	repo.AssertExpectations(t)
}

func TestService_GetByID_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	expectedIncident := &Incident{
		ID:     "inc-1",
		Type:   "fall",
		Status: IncidentStatusSubmitted,
	}

	repo.On("GetByID", mock.Anything, "inc-1").Return(expectedIncident, nil)

	incident, err := svc.GetByID(context.Background(), "inc-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedIncident, incident)
	repo.AssertExpectations(t)
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	repo.On("GetByID", mock.Anything, "inc-999").Return(nil, ErrNotFound)

	incident, err := svc.GetByID(context.Background(), "inc-999")

	assert.Error(t, err)
	assert.Nil(t, incident)
	assert.Equal(t, ErrNotFound, err)
	repo.AssertExpectations(t)
}

func TestService_List_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	filter := ListIncidentsFilter{
		Status: IncidentStatusSubmitted,
		Limit:  10,
		Offset: 0,
	}

	expectedIncidents := []*Incident{
		{ID: "inc-1", Type: "fall", Status: IncidentStatusSubmitted},
		{ID: "inc-2", Type: "medication_error", Status: IncidentStatusSubmitted},
	}

	repo.On("List", mock.Anything, filter).Return(expectedIncidents, int64(2), nil)

	incidents, total, err := svc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedIncidents, incidents)
	assert.Equal(t, int64(2), total)
	repo.AssertExpectations(t)
}

func TestService_List_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	filter := ListIncidentsFilter{Limit: 10}
	dbErr := errors.New("database error")

	repo.On("List", mock.Anything, filter).Return(nil, int64(0), dbErr)

	incidents, total, err := svc.List(context.Background(), filter)

	assert.Error(t, err)
	assert.Nil(t, incidents)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, dbErr, err)
	repo.AssertExpectations(t)
}

func TestService_UpdateStatus_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingIncident := &Incident{
		ID:     "inc-1",
		Type:   "fall",
		Status: IncidentStatusSubmitted,
	}

	req := &UpdateIncidentStatusRequest{
		Status: IncidentStatusUnderReview,
	}

	repo.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	repo.On("CreateStatusEvent", mock.Anything, mock.MatchedBy(func(e *IncidentStatusEvent) bool {
		return e.IncidentID == "inc-1" &&
			e.ToStatus == IncidentStatusUnderReview &&
			e.ChangedBy == "user-1"
	})).Return(nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(i *Incident) bool {
		return i.ID == "inc-1" && i.Status == IncidentStatusUnderReview
	})).Return(nil)

	incident, err := svc.UpdateStatus(context.Background(), "inc-1", req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, incident)
	assert.Equal(t, IncidentStatusUnderReview, incident.Status)
	repo.AssertExpectations(t)
}

func TestService_UpdateStatus_WithNote(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusUnderReview,
	}

	note := "Reviewed by safety team"
	req := &UpdateIncidentStatusRequest{
		Status: IncidentStatusClosed,
		Note:   &note,
	}

	repo.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	repo.On("CreateStatusEvent", mock.Anything, mock.MatchedBy(func(e *IncidentStatusEvent) bool {
		return e.Note != nil && *e.Note == "Reviewed by safety team"
	})).Return(nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*incident.Incident")).Return(nil)

	incident, err := svc.UpdateStatus(context.Background(), "inc-1", req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, incident)
	assert.Equal(t, IncidentStatusClosed, incident.Status)
	repo.AssertExpectations(t)
}

func TestService_UpdateStatus_IncidentNotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &UpdateIncidentStatusRequest{
		Status: IncidentStatusClosed,
	}

	repo.On("GetByID", mock.Anything, "inc-999").Return(nil, ErrNotFound)

	incident, err := svc.UpdateStatus(context.Background(), "inc-999", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, incident)
	assert.Equal(t, ErrNotFound, err)
	repo.AssertExpectations(t)
}

func TestService_UpdateStatus_EmptyStatus(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusSubmitted,
	}

	req := &UpdateIncidentStatusRequest{
		Status: "",
	}

	repo.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)

	incident, err := svc.UpdateStatus(context.Background(), "inc-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, incident)
	assert.Equal(t, "status is required", err.Error())
	repo.AssertExpectations(t)
}

func TestService_UpdateStatus_CreateEventError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusSubmitted,
	}

	req := &UpdateIncidentStatusRequest{
		Status: IncidentStatusUnderReview,
	}

	eventErr := errors.New("event creation error")
	repo.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	repo.On("CreateStatusEvent", mock.Anything, mock.AnythingOfType("*incident.IncidentStatusEvent")).Return(eventErr)

	incident, err := svc.UpdateStatus(context.Background(), "inc-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, incident)
	assert.Equal(t, eventErr, err)
	repo.AssertExpectations(t)
}

func TestService_UpdateStatus_UpdateError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusSubmitted,
	}

	req := &UpdateIncidentStatusRequest{
		Status: IncidentStatusUnderReview,
	}

	updateErr := errors.New("update error")
	repo.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	repo.On("CreateStatusEvent", mock.Anything, mock.AnythingOfType("*incident.IncidentStatusEvent")).Return(nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*incident.Incident")).Return(updateErr)

	incident, err := svc.UpdateStatus(context.Background(), "inc-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, incident)
	assert.Equal(t, updateErr, err)
	repo.AssertExpectations(t)
}

func TestService_GetStatusHistory_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusClosed,
	}

	fromStatus := IncidentStatusSubmitted
	expectedEvents := []*IncidentStatusEvent{
		{
			ID:         "event-1",
			IncidentID: "inc-1",
			FromStatus: &fromStatus,
			ToStatus:   IncidentStatusUnderReview,
		},
		{
			ID:         "event-2",
			IncidentID: "inc-1",
			FromStatus: ptrTo(IncidentStatusUnderReview),
			ToStatus:   IncidentStatusClosed,
		},
	}

	repo.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	repo.On("GetStatusHistory", mock.Anything, "inc-1").Return(expectedEvents, nil)

	events, err := svc.GetStatusHistory(context.Background(), "inc-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedEvents, events)
	assert.Len(t, events, 2)
	repo.AssertExpectations(t)
}

func TestService_GetStatusHistory_IncidentNotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	repo.On("GetByID", mock.Anything, "inc-999").Return(nil, ErrNotFound)

	events, err := svc.GetStatusHistory(context.Background(), "inc-999")

	assert.Error(t, err)
	assert.Nil(t, events)
	assert.Equal(t, ErrNotFound, err)
	repo.AssertExpectations(t)
}

func TestService_GetStatusHistory_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingIncident := &Incident{
		ID: "inc-1",
	}

	dbErr := errors.New("database error")
	repo.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	repo.On("GetStatusHistory", mock.Anything, "inc-1").Return(nil, dbErr)

	events, err := svc.GetStatusHistory(context.Background(), "inc-1")

	assert.Error(t, err)
	assert.Nil(t, events)
	assert.Equal(t, dbErr, err)
	repo.AssertExpectations(t)
}

// Helper function to create pointer to IncidentStatus
func ptrTo(s IncidentStatus) *IncidentStatus {
	return &s
}
