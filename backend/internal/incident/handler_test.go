package incident

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
	"github.com/wardflow/backend/pkg/auth"
)

// mockService is a mock implementation of Service
type mockService struct {
	mock.Mock
}

func (m *mockService) Create(ctx context.Context, req *CreateIncidentRequest, byUserID string) (*Incident, error) {
	args := m.Called(ctx, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Incident), args.Error(1)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*Incident, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Incident), args.Error(1)
}

func (m *mockService) List(ctx context.Context, f ListIncidentsFilter) ([]*Incident, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Incident), args.Get(1).(int64), args.Error(2)
}

func (m *mockService) UpdateStatus(ctx context.Context, id string, req *UpdateIncidentStatusRequest, byUserID string) (*Incident, error) {
	args := m.Called(ctx, id, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Incident), args.Error(1)
}

func (m *mockService) GetStatusHistory(ctx context.Context, incidentID string) ([]*IncidentStatusEvent, error) {
	args := m.Called(ctx, incidentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*IncidentStatusEvent), args.Error(1)
}

func TestHandler_Create_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateIncidentRequest{
		Type:      "fall",
		EventTime: time.Now().UTC(),
	}

	expectedIncident := &Incident{
		ID:         "inc-1",
		Type:       "fall",
		Status:     IncidentStatusSubmitted,
		ReportedBy: "user-1",
	}

	svc.On("Create", mock.Anything, mock.AnythingOfType("*incident.CreateIncidentRequest"), "user-1").
		Return(expectedIncident, nil)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/incidents", reqBody, "user-1", models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var result Incident
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, "inc-1", result.ID)
	assert.Equal(t, "fall", result.Type)
	svc.AssertExpectations(t)
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/incidents", "invalid-json")
	req = req.WithContext(auth.SetUserContext(req.Context(), &auth.Claims{
		UserID: "user-1",
		Role:   models.RoleNurse,
	}))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Create_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateIncidentRequest{
		Type:      "fall",
		EventTime: time.Now().UTC(),
	}

	svc.On("Create", mock.Anything, mock.AnythingOfType("*incident.CreateIncidentRequest"), "user-1").
		Return(nil, errors.New("validation error"))

	req := testutil.NewRequest(http.MethodPost, "/api/v1/incidents", reqBody, "user-1", models.RoleNurse)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetByID_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	expectedIncident := &Incident{
		ID:     "inc-1",
		Type:   "fall",
		Status: IncidentStatusSubmitted,
	}

	svc.On("GetByID", mock.Anything, "inc-1").Return(expectedIncident, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents/inc-1", nil, "user-1", models.RoleNurse)
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result Incident
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, "inc-1", result.ID)
	svc.AssertExpectations(t)
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetByID", mock.Anything, "inc-999").Return(nil, ErrNotFound)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents/inc-999", nil, "user-1", models.RoleNurse)
	req.SetPathValue("incidentId", "inc-999")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetByID_MissingID(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents/", nil, "user-1", models.RoleNurse)
	// Do not set PathValue
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_List_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	incidents := []*Incident{
		{ID: "inc-1", Type: "fall", Status: IncidentStatusSubmitted},
		{ID: "inc-2", Type: "medication_error", Status: IncidentStatusUnderReview},
	}

	svc.On("List", mock.Anything, mock.MatchedBy(func(f ListIncidentsFilter) bool {
		return f.Status == IncidentStatusSubmitted && f.Limit == 20
	})).Return(incidents, int64(2), nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents?status=submitted", nil, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result models.PaginatedResponse
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, int64(2), result.Total)
	svc.AssertExpectations(t)
}

func TestHandler_List_WithUnitFilter_AdminAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	incidents := []*Incident{
		{ID: "inc-1", Type: "fall"},
	}

	svc.On("List", mock.Anything, mock.AnythingOfType("incident.ListIncidentsFilter")).
		Return(incidents, int64(1), nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents?unitId=unit-1", nil, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_List_WithUnitFilter_NonAdminWithAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	incidents := []*Incident{
		{ID: "inc-1", Type: "fall"},
	}

	svc.On("List", mock.Anything, mock.AnythingOfType("incident.ListIncidentsFilter")).
		Return(incidents, int64(1), nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents?unitId=unit-1", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_List_WithUnitFilter_NonAdminWithoutAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents?unitId=unit-1", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-2", "unit-3"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_List_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("List", mock.Anything, mock.AnythingOfType("incident.ListIncidentsFilter")).
		Return(nil, int64(0), errors.New("database error"))

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents", nil, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateStatus_Success_QualitySafety(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := UpdateIncidentStatusRequest{
		Status: IncidentStatusUnderReview,
	}

	existingIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusSubmitted,
	}

	updatedIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusUnderReview,
	}

	svc.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	svc.On("UpdateStatus", mock.Anything, "inc-1", mock.AnythingOfType("*incident.UpdateIncidentStatusRequest"), "user-1").
		Return(updatedIncident, nil)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/incidents/inc-1/status", reqBody, "user-1", models.RoleQualitySafety)
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result Incident
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, IncidentStatusUnderReview, result.Status)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateStatus_Success_Admin(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := UpdateIncidentStatusRequest{
		Status: IncidentStatusClosed,
	}

	existingIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusUnderReview,
	}

	updatedIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusClosed,
	}

	svc.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	svc.On("UpdateStatus", mock.Anything, "inc-1", mock.AnythingOfType("*incident.UpdateIncidentStatusRequest"), "user-1").
		Return(updatedIncident, nil)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/incidents/inc-1/status", reqBody, "user-1", models.RoleAdmin)
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateStatus_InsufficientPermissions(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := UpdateIncidentStatusRequest{
		Status: IncidentStatusUnderReview,
	}

	req := testutil.NewRequest(http.MethodPost, "/api/v1/incidents/inc-1/status", reqBody, "user-1", models.RoleNurse)
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_UpdateStatus_ProviderInsufficientPermissions(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := UpdateIncidentStatusRequest{
		Status: IncidentStatusUnderReview,
	}

	req := testutil.NewRequest(http.MethodPost, "/api/v1/incidents/inc-1/status", reqBody, "user-1", models.RoleProvider)
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_UpdateStatus_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := UpdateIncidentStatusRequest{
		Status: IncidentStatusUnderReview,
	}

	svc.On("GetByID", mock.Anything, "inc-999").Return(nil, ErrNotFound)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/incidents/inc-999/status", reqBody, "user-1", models.RoleQualitySafety)
	req.SetPathValue("incidentId", "inc-999")
	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_UpdateStatus_InvalidJSON(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/incidents/inc-1/status", "invalid-json")
	req = req.WithContext(auth.SetUserContext(req.Context(), &auth.Claims{
		UserID: "user-1",
		Role:   models.RoleQualitySafety,
	}))
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_UpdateStatus_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := UpdateIncidentStatusRequest{
		Status: IncidentStatusUnderReview,
	}

	existingIncident := &Incident{
		ID:     "inc-1",
		Status: IncidentStatusSubmitted,
	}

	svc.On("GetByID", mock.Anything, "inc-1").Return(existingIncident, nil)
	svc.On("UpdateStatus", mock.Anything, "inc-1", mock.AnythingOfType("*incident.UpdateIncidentStatusRequest"), "user-1").
		Return(nil, errors.New("update failed"))

	req := testutil.NewRequest(http.MethodPost, "/api/v1/incidents/inc-1/status", reqBody, "user-1", models.RoleQualitySafety)
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.UpdateStatus(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetStatusHistory_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	fromStatus := IncidentStatusSubmitted
	events := []*IncidentStatusEvent{
		{
			ID:         "event-1",
			IncidentID: "inc-1",
			FromStatus: &fromStatus,
			ToStatus:   IncidentStatusUnderReview,
			ChangedBy:  "user-1",
		},
		{
			ID:         "event-2",
			IncidentID: "inc-1",
			FromStatus: ptrTo(IncidentStatusUnderReview),
			ToStatus:   IncidentStatusClosed,
			ChangedBy:  "user-2",
		},
	}

	svc.On("GetStatusHistory", mock.Anything, "inc-1").Return(events, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents/inc-1/status/history", nil, "user-1", models.RoleNurse)
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.GetStatusHistory(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result []*IncidentStatusEvent
	testutil.DecodeJSON(t, rr, &result)
	assert.Len(t, result, 2)
	assert.Equal(t, "event-1", result[0].ID)
	svc.AssertExpectations(t)
}

func TestHandler_GetStatusHistory_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetStatusHistory", mock.Anything, "inc-999").Return(nil, ErrNotFound)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents/inc-999/status/history", nil, "user-1", models.RoleNurse)
	req.SetPathValue("incidentId", "inc-999")
	rr := httptest.NewRecorder()

	handler.GetStatusHistory(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetStatusHistory_MissingID(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents//status/history", nil, "user-1", models.RoleNurse)
	// Do not set PathValue
	rr := httptest.NewRecorder()

	handler.GetStatusHistory(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_GetStatusHistory_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetStatusHistory", mock.Anything, "inc-1").Return(nil, errors.New("database error"))

	req := testutil.NewRequest(http.MethodGet, "/api/v1/incidents/inc-1/status/history", nil, "user-1", models.RoleNurse)
	req.SetPathValue("incidentId", "inc-1")
	rr := httptest.NewRecorder()

	handler.GetStatusHistory(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}
