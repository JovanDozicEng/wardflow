package transport

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// mockRepository is a mock implementation of Repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) ListRequests(ctx context.Context, filter ListTransportFilter) ([]TransportRequest, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]TransportRequest), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) CreateRequest(ctx context.Context, req *TransportRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockRepository) GetRequestByID(ctx context.Context, id string) (*TransportRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TransportRequest), args.Error(1)
}

func (m *mockRepository) UpdateRequestFields(ctx context.Context, requestID string, updates map[string]any) error {
	args := m.Called(ctx, requestID, updates)
	return args.Error(0)
}

func (m *mockRepository) CreateChangeEvent(ctx context.Context, event *TransportChangeEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestService_CreateRequest_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	req := CreateTransportRequest{
		EncounterID: "enc-1",
		Origin:      "ER",
		Destination: "ICU",
		Priority:    "urgent",
	}

	repo.On("CreateRequest", mock.Anything, mock.MatchedBy(func(tr *TransportRequest) bool {
		return tr.EncounterID == "enc-1" &&
			tr.Origin == "ER" &&
			tr.Destination == "ICU" &&
			tr.Priority == "urgent" &&
			tr.Status == TransportStatusPending
	})).Run(func(args mock.Arguments) {
		// Simulate database auto-generating ID
		tr := args.Get(1).(*TransportRequest)
		tr.ID = "tr-123"
	}).Return(nil)

	result, err := svc.CreateRequest(context.Background(), req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "tr-123", result.ID)
	assert.Equal(t, "enc-1", result.EncounterID)
	assert.Equal(t, "ER", result.Origin)
	assert.Equal(t, "ICU", result.Destination)
	assert.Equal(t, "urgent", result.Priority)
	assert.Equal(t, TransportStatusPending, result.Status)
	assert.Equal(t, "user-1", result.CreatedBy)
	repo.AssertExpectations(t)
}

func TestService_CreateRequest_DefaultPriority(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	req := CreateTransportRequest{
		EncounterID: "enc-1",
		Origin:      "ER",
		Destination: "ICU",
		// Priority not provided
	}

	repo.On("CreateRequest", mock.Anything, mock.MatchedBy(func(tr *TransportRequest) bool {
		return tr.Priority == "routine"
	})).Run(func(args mock.Arguments) {
		// Simulate database auto-generating ID
		tr := args.Get(1).(*TransportRequest)
		tr.ID = "tr-124"
	}).Return(nil)

	result, err := svc.CreateRequest(context.Background(), req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "tr-124", result.ID)
	assert.Equal(t, "routine", result.Priority)
	repo.AssertExpectations(t)
}

func TestService_CreateRequest_MissingEncounterID(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	req := CreateTransportRequest{
		Origin:      "ER",
		Destination: "ICU",
	}

	result, err := svc.CreateRequest(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "encounterId")
}

func TestService_CreateRequest_MissingOrigin(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	req := CreateTransportRequest{
		EncounterID: "enc-1",
		Destination: "ICU",
	}

	result, err := svc.CreateRequest(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "origin")
}

func TestService_CreateRequest_MissingDestination(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	req := CreateTransportRequest{
		EncounterID: "enc-1",
		Origin:      "ER",
	}

	result, err := svc.CreateRequest(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "destination")
}

func TestService_CreateRequest_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	req := CreateTransportRequest{
		EncounterID: "enc-1",
		Origin:      "ER",
		Destination: "ICU",
	}

	dbErr := errors.New("database error")
	repo.On("CreateRequest", mock.Anything, mock.AnythingOfType("*transport.TransportRequest")).Return(dbErr)

	result, err := svc.CreateRequest(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbErr, err)
	repo.AssertExpectations(t)
}

func TestService_GetRequest_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	expectedTR := &TransportRequest{
		ID:          "tr-1",
		EncounterID: "enc-1",
		Origin:      "ER",
		Destination: "ICU",
		Status:      TransportStatusPending,
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(expectedTR, nil)

	result, err := svc.GetRequest(context.Background(), "tr-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedTR, result)
	repo.AssertExpectations(t)
}

func TestService_GetRequest_NotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	repo.On("GetRequestByID", mock.Anything, "tr-999").Return(nil, gorm.ErrRecordNotFound)

	result, err := svc.GetRequest(context.Background(), "tr-999")

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestService_ListRequests_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	filter := ListTransportFilter{
		Status: "pending",
		Limit:  50,
		Offset: 0,
	}

	expectedTRs := []TransportRequest{
		{ID: "tr-1", Status: TransportStatusPending},
		{ID: "tr-2", Status: TransportStatusPending},
	}

	repo.On("ListRequests", mock.Anything, filter).Return(expectedTRs, int64(2), nil)

	result, total, err := svc.ListRequests(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedTRs, result)
	assert.Equal(t, int64(2), total)
	repo.AssertExpectations(t)
}

func TestService_AcceptRequest_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusPending,
	}

	updatedTR := &TransportRequest{
		ID:         "tr-1",
		Status:     TransportStatusAssigned,
		AssignedTo: strPtr("transport-user"),
	}

	req := AcceptTransportRequest{
		AssignedTo: "transport-user",
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil).Once()
	repo.On("UpdateRequestFields", mock.Anything, "tr-1", mock.MatchedBy(func(updates map[string]any) bool {
		return updates["status"] == TransportStatusAssigned &&
			updates["assigned_to"] == "transport-user"
	})).Return(nil)
	repo.On("CreateChangeEvent", mock.Anything, mock.AnythingOfType("*transport.TransportChangeEvent")).Return(nil)
	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(updatedTR, nil).Once()

	result, err := svc.AcceptRequest(context.Background(), "tr-1", req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TransportStatusAssigned, result.Status)
	assert.Equal(t, "transport-user", *result.AssignedTo)
	repo.AssertExpectations(t)
}

func TestService_AcceptRequest_DefaultAssignedToCurrentUser(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusPending,
	}

	updatedTR := &TransportRequest{
		ID:         "tr-1",
		Status:     TransportStatusAssigned,
		AssignedTo: strPtr("user-1"),
	}

	req := AcceptTransportRequest{
		AssignedTo: "", // Empty, should default to userID
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil).Once()
	repo.On("UpdateRequestFields", mock.Anything, "tr-1", mock.MatchedBy(func(updates map[string]any) bool {
		return updates["assigned_to"] == "user-1"
	})).Return(nil)
	repo.On("CreateChangeEvent", mock.Anything, mock.AnythingOfType("*transport.TransportChangeEvent")).Return(nil)
	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(updatedTR, nil).Once()

	result, err := svc.AcceptRequest(context.Background(), "tr-1", req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-1", *result.AssignedTo)
	repo.AssertExpectations(t)
}

func TestService_AcceptRequest_InvalidState(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusCompleted,
	}

	req := AcceptTransportRequest{
		AssignedTo: "transport-user",
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil)

	result, err := svc.AcceptRequest(context.Background(), "tr-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &InvalidStateError{}, err)
	repo.AssertExpectations(t)
}

func TestService_UpdateRequest_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:          "tr-1",
		Status:      TransportStatusPending,
		Origin:      "ER",
		Destination: "ICU",
	}

	newOrigin := "Room 101"
	newDest := "Radiology"
	newPriority := "emergent"

	req := UpdateTransportRequest{
		Origin:      &newOrigin,
		Destination: &newDest,
		Priority:    &newPriority,
	}

	updatedTR := &TransportRequest{
		ID:          "tr-1",
		Status:      TransportStatusPending,
		Origin:      "Room 101",
		Destination: "Radiology",
		Priority:    "emergent",
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil).Once()
	repo.On("UpdateRequestFields", mock.Anything, "tr-1", mock.MatchedBy(func(updates map[string]any) bool {
		return updates["origin"] == "Room 101" &&
			updates["destination"] == "Radiology" &&
			updates["priority"] == "emergent"
	})).Return(nil)
	repo.On("CreateChangeEvent", mock.Anything, mock.AnythingOfType("*transport.TransportChangeEvent")).Return(nil)
	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(updatedTR, nil).Once()

	result, err := svc.UpdateRequest(context.Background(), "tr-1", req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Room 101", result.Origin)
	assert.Equal(t, "Radiology", result.Destination)
	assert.Equal(t, "emergent", result.Priority)
	repo.AssertExpectations(t)
}

func TestService_UpdateRequest_CannotUpdateCompleted(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusCompleted,
	}

	newOrigin := "Room 101"
	req := UpdateTransportRequest{
		Origin: &newOrigin,
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil)

	result, err := svc.UpdateRequest(context.Background(), "tr-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &InvalidStateError{}, err)
	repo.AssertExpectations(t)
}

func TestService_UpdateRequest_CannotUpdateCancelled(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusCancelled,
	}

	newOrigin := "Room 101"
	req := UpdateTransportRequest{
		Origin: &newOrigin,
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil)

	result, err := svc.UpdateRequest(context.Background(), "tr-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &InvalidStateError{}, err)
	repo.AssertExpectations(t)
}

func TestService_UpdateRequest_NoFieldsProvided(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusPending,
	}

	req := UpdateTransportRequest{
		// No fields
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil)

	result, err := svc.UpdateRequest(context.Background(), "tr-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &ValidationError{}, err)
	repo.AssertExpectations(t)
}

func TestService_CompleteRequest_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusAssigned,
	}

	completedTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusCompleted,
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil).Once()
	repo.On("UpdateRequestFields", mock.Anything, "tr-1", mock.MatchedBy(func(updates map[string]any) bool {
		return updates["status"] == TransportStatusCompleted
	})).Return(nil)
	repo.On("CreateChangeEvent", mock.Anything, mock.AnythingOfType("*transport.TransportChangeEvent")).Return(nil)
	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(completedTR, nil).Once()

	result, err := svc.CompleteRequest(context.Background(), "tr-1", "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, TransportStatusCompleted, result.Status)
	repo.AssertExpectations(t)
}

func TestService_CompleteRequest_InvalidState(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	existingTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusPending,
	}

	repo.On("GetRequestByID", mock.Anything, "tr-1").Return(existingTR, nil)

	result, err := svc.CompleteRequest(context.Background(), "tr-1", "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &InvalidStateError{}, err)
	assert.Contains(t, err.Error(), "must be accepted")
	repo.AssertExpectations(t)
}

func TestService_CompleteRequest_NotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)

	repo.On("GetRequestByID", mock.Anything, "tr-999").Return(nil, gorm.ErrRecordNotFound)

	result, err := svc.CompleteRequest(context.Background(), "tr-999", "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}
