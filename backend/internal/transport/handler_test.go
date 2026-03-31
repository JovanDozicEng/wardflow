package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
	"github.com/wardflow/backend/pkg/auth"
	"gorm.io/gorm"
)

// mockService is a mock implementation of Service
type mockService struct {
	mock.Mock
}

func (m *mockService) ListRequests(ctx context.Context, filter ListTransportFilter) ([]TransportRequest, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]TransportRequest), args.Get(1).(int64), args.Error(2)
}

func (m *mockService) CreateRequest(ctx context.Context, req CreateTransportRequest, userID string) (*TransportRequest, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TransportRequest), args.Error(1)
}

func (m *mockService) GetRequest(ctx context.Context, id string) (*TransportRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TransportRequest), args.Error(1)
}

func (m *mockService) AcceptRequest(ctx context.Context, requestID string, req AcceptTransportRequest, userID string) (*TransportRequest, error) {
	args := m.Called(ctx, requestID, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TransportRequest), args.Error(1)
}

func (m *mockService) UpdateRequest(ctx context.Context, requestID string, req UpdateTransportRequest, userID string) (*TransportRequest, error) {
	args := m.Called(ctx, requestID, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TransportRequest), args.Error(1)
}

func (m *mockService) CompleteRequest(ctx context.Context, requestID, userID string) (*TransportRequest, error) {
	args := m.Called(ctx, requestID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TransportRequest), args.Error(1)
}

func TestHandler_CreateRequest_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateTransportRequest{
		EncounterID: "enc-1",
		Origin:      "ER",
		Destination: "ICU",
		Priority:    "urgent",
	}

	expectedTR := &TransportRequest{
		ID:          "tr-1",
		EncounterID: "enc-1",
		Origin:      "ER",
		Destination: "ICU",
		Priority:    "urgent",
		Status:      TransportStatusPending,
		CreatedBy:   "user-1",
	}

	svc.On("CreateRequest", mock.Anything, reqBody, "user-1").Return(expectedTR, nil)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/transport/requests", reqBody, "user-1", models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.CreateRequest(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var result TransportRequest
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, "tr-1", result.ID)
	assert.Equal(t, "ER", result.Origin)
	svc.AssertExpectations(t)
}

func TestHandler_CreateRequest_InvalidJSON(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/transport/requests", "invalid-json")
	req = req.WithContext(auth.SetUserContext(req.Context(), &auth.Claims{
		UserID: "user-1",
		Role:   models.RoleNurse,
	}))
	rr := httptest.NewRecorder()

	handler.CreateRequest(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_CreateRequest_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateTransportRequest{
		EncounterID: "enc-1",
		Origin:      "ER",
		Destination: "ICU",
	}

	svc.On("CreateRequest", mock.Anything, reqBody, "user-1").Return(nil, errors.New("validation error"))

	req := testutil.NewRequest(http.MethodPost, "/api/v1/transport/requests", reqBody, "user-1", models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.CreateRequest(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_ListRequests_Success_Admin(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	requests := []TransportRequest{
		{ID: "tr-1", Status: TransportStatusPending},
		{ID: "tr-2", Status: TransportStatusAssigned},
	}

	svc.On("ListRequests", mock.Anything, mock.MatchedBy(func(f ListTransportFilter) bool {
		return f.Status == "pending" && f.Limit == 50
	})).Return(requests, int64(2), nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/transport/requests?status=pending", nil, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.ListRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_ListRequests_NonAdmin_NoUnitFilter(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	requests := []TransportRequest{
		{ID: "tr-1", Status: TransportStatusPending},
	}

	// Non-admin without unit filter should have their unit IDs applied
	svc.On("ListRequests", mock.Anything, mock.MatchedBy(func(f ListTransportFilter) bool {
		return len(f.UnitIDs) == 2 && f.UnitID == ""
	})).Return(requests, int64(1), nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/transport/requests", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ListRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_ListRequests_NonAdmin_WithAllowedUnitFilter(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	requests := []TransportRequest{
		{ID: "tr-1", Status: TransportStatusPending},
	}

	svc.On("ListRequests", mock.Anything, mock.MatchedBy(func(f ListTransportFilter) bool {
		return f.UnitID == "unit-1" && len(f.UnitIDs) == 0
	})).Return(requests, int64(1), nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/transport/requests?unitId=unit-1", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ListRequests(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_ListRequests_NonAdmin_WithDisallowedUnitFilter(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/transport/requests?unitId=unit-3", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ListRequests(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_ListRequests_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("ListRequests", mock.Anything, mock.AnythingOfType("transport.ListTransportFilter")).
		Return(nil, int64(0), errors.New("database error"))

	req := testutil.NewRequest(http.MethodGet, "/api/v1/transport/requests", nil, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.ListRequests(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_AcceptRequest_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := AcceptTransportRequest{
		AssignedTo: "transport-user",
	}

	expectedTR := &TransportRequest{
		ID:         "tr-1",
		Status:     TransportStatusAssigned,
		AssignedTo: strPtr("transport-user"),
	}

	svc.On("AcceptRequest", mock.Anything, "tr-1", reqBody, "user-1").Return(expectedTR, nil)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/transport/requests/tr-1/accept", reqBody, "user-1", models.RoleTransport)
	req.SetPathValue("requestId", "tr-1")
	rr := httptest.NewRecorder()

	handler.AcceptRequest(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result TransportRequest
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, TransportStatusAssigned, result.Status)
	svc.AssertExpectations(t)
}

func TestHandler_AcceptRequest_InvalidJSON(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/transport/requests/tr-1/accept", "invalid-json")
	req = req.WithContext(auth.SetUserContext(req.Context(), &auth.Claims{
		UserID: "user-1",
		Role:   models.RoleTransport,
	}))
	req.SetPathValue("requestId", "tr-1")
	rr := httptest.NewRecorder()

	handler.AcceptRequest(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_AcceptRequest_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := AcceptTransportRequest{
		AssignedTo: "transport-user",
	}

	svc.On("AcceptRequest", mock.Anything, "tr-1", reqBody, "user-1").
		Return(nil, &InvalidStateError{Message: "not pending"})

	req := testutil.NewRequest(http.MethodPost, "/api/v1/transport/requests/tr-1/accept", reqBody, "user-1", models.RoleTransport)
	req.SetPathValue("requestId", "tr-1")
	rr := httptest.NewRecorder()

	handler.AcceptRequest(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateRequest_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	newOrigin := "Room 101"
	reqBody := UpdateTransportRequest{
		Origin: &newOrigin,
	}

	expectedTR := &TransportRequest{
		ID:     "tr-1",
		Origin: "Room 101",
		Status: TransportStatusPending,
	}

	svc.On("UpdateRequest", mock.Anything, "tr-1", reqBody, "user-1").Return(expectedTR, nil)

	req := testutil.NewRequest(http.MethodPatch, "/api/v1/transport/requests/tr-1", reqBody, "user-1", models.RoleNurse)
	req.SetPathValue("requestId", "tr-1")
	rr := httptest.NewRecorder()

	handler.UpdateRequest(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result TransportRequest
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, "Room 101", result.Origin)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateRequest_InvalidJSON(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequestNoAuth(http.MethodPatch, "/api/v1/transport/requests/tr-1", "invalid-json")
	req = req.WithContext(auth.SetUserContext(req.Context(), &auth.Claims{
		UserID: "user-1",
		Role:   models.RoleNurse,
	}))
	req.SetPathValue("requestId", "tr-1")
	rr := httptest.NewRecorder()

	handler.UpdateRequest(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_UpdateRequest_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	newOrigin := "Room 101"
	reqBody := UpdateTransportRequest{
		Origin: &newOrigin,
	}

	svc.On("UpdateRequest", mock.Anything, "tr-1", reqBody, "user-1").
		Return(nil, errors.New("update failed"))

	req := testutil.NewRequest(http.MethodPatch, "/api/v1/transport/requests/tr-1", reqBody, "user-1", models.RoleNurse)
	req.SetPathValue("requestId", "tr-1")
	rr := httptest.NewRecorder()

	handler.UpdateRequest(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_CompleteRequest_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	expectedTR := &TransportRequest{
		ID:     "tr-1",
		Status: TransportStatusCompleted,
	}

	svc.On("CompleteRequest", mock.Anything, "tr-1", "user-1").Return(expectedTR, nil)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/transport/requests/tr-1/complete", nil, "user-1", models.RoleTransport)
	req.SetPathValue("requestId", "tr-1")
	rr := httptest.NewRecorder()

	handler.CompleteRequest(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result TransportRequest
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, TransportStatusCompleted, result.Status)
	svc.AssertExpectations(t)
}

func TestHandler_CompleteRequest_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("CompleteRequest", mock.Anything, "tr-1", "user-1").
		Return(nil, &InvalidStateError{Message: "not assigned"})

	req := testutil.NewRequest(http.MethodPost, "/api/v1/transport/requests/tr-1/complete", nil, "user-1", models.RoleTransport)
	req.SetPathValue("requestId", "tr-1")
	rr := httptest.NewRecorder()

	handler.CompleteRequest(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_CompleteRequest_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("CompleteRequest", mock.Anything, "tr-999", "user-1").
		Return(nil, gorm.ErrRecordNotFound)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/transport/requests/tr-999/complete", nil, "user-1", models.RoleTransport)
	req.SetPathValue("requestId", "tr-999")
	rr := httptest.NewRecorder()

	handler.CompleteRequest(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}
