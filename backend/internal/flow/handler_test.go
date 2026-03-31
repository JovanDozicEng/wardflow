package flow

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
)

// mockService is a mock implementation of Service
type mockService struct {
	mock.Mock
}

func (m *mockService) RecordTransition(ctx context.Context, r *http.Request, encounterID string, req CreateTransitionRequest, currentUserID string) (*FlowStateTransition, error) {
	args := m.Called(ctx, r, encounterID, req, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FlowStateTransition), args.Error(1)
}

func (m *mockService) OverrideTransition(ctx context.Context, r *http.Request, encounterID string, req OverrideTransitionRequest, currentUserID string, userRole models.Role) (*FlowStateTransition, error) {
	args := m.Called(ctx, r, encounterID, req, currentUserID, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FlowStateTransition), args.Error(1)
}

func (m *mockService) GetTimeline(ctx context.Context, encounterID string) (*FlowTimelineResponse, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FlowTimelineResponse), args.Error(1)
}

func (m *mockService) GetTimelinePaginated(ctx context.Context, encounterID string, limit, offset int) (*FlowTimelineResponse, error) {
	args := m.Called(ctx, encounterID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FlowTimelineResponse), args.Error(1)
}

func (m *mockService) GetTimelineWithActors(ctx context.Context, encounterID string) (*FlowTimelineDetailResponse, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FlowTimelineDetailResponse), args.Error(1)
}

func (m *mockService) GetCurrentState(ctx context.Context, encounterID string) (*FlowState, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FlowState), args.Error(1)
}

func TestHandler_GetFlowTimeline_Simple(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	arrivedState := StateArrived
	triageState := StateTriage

	timeline := &FlowTimelineResponse{
		EncounterID:  "encounter-1",
		CurrentState: &triageState,
		Transitions: []FlowStateTransition{
			{ID: "t1", ToState: arrivedState},
			{ID: "t2", ToState: triageState},
		},
		Total: 2,
	}

	svc.On("GetTimeline", mock.Anything, "encounter-1").Return(timeline, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/flow", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetFlowTimeline(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response FlowTimelineResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "encounter-1", response.EncounterID)
	assert.Equal(t, triageState, *response.CurrentState)
	assert.Len(t, response.Transitions, 2)
	svc.AssertExpectations(t)
}

func TestHandler_GetFlowTimeline_WithActors(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	triageState := StateTriage
	actorName := "John Doe"

	timeline := &FlowTimelineDetailResponse{
		EncounterID:  "encounter-1",
		CurrentState: &triageState,
		Transitions: []TransitionWithActor{
			{
				FlowStateTransition: FlowStateTransition{ID: "t1", ToState: triageState},
				ActorName:           &actorName,
			},
		},
		Total: 1,
	}

	svc.On("GetTimelineWithActors", mock.Anything, "encounter-1").Return(timeline, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/flow?withActors=true", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetFlowTimeline(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response FlowTimelineDetailResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "encounter-1", response.EncounterID)
	assert.Len(t, response.Transitions, 1)
	assert.Equal(t, "John Doe", *response.Transitions[0].ActorName)
	svc.AssertExpectations(t)
}

func TestHandler_GetFlowTimeline_Paginated(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	triageState := StateTriage
	timeline := &FlowTimelineResponse{
		EncounterID:  "encounter-1",
		CurrentState: &triageState,
		Transitions:  []FlowStateTransition{{ID: "t1", ToState: triageState}},
		Total:        10,
	}

	svc.On("GetTimelinePaginated", mock.Anything, "encounter-1", 5, 0).Return(timeline, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/flow?paginated=true&limit=5", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetFlowTimeline(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response FlowTimelineResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, int64(10), response.Total)
	svc.AssertExpectations(t)
}

func TestHandler_GetFlowTimeline_MissingEncounterID(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters//flow", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "")

	rr := httptest.NewRecorder()
	handler.GetFlowTimeline(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_RecordTransition_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateTransitionRequest{
		ToState: StateTriage,
	}

	triageState := StateTriage
	arrivedState := StateArrived
	transition := &FlowStateTransition{
		ID:          "trans-1",
		EncounterID: "encounter-1",
		FromState:   &arrivedState,
		ToState:     triageState,
	}

	svc.On("RecordTransition", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1").Return(transition, nil)

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/flow/transitions", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.RecordTransition(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response FlowStateTransition
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "trans-1", response.ID)
	assert.Equal(t, triageState, response.ToState)
	svc.AssertExpectations(t)
}

func TestHandler_RecordTransition_InvalidTransition(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateTransitionRequest{
		ToState: StateDischarged,
	}

	svc.On("RecordTransition", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1").
		Return(nil, fmt.Errorf("invalid transition from arrived to discharged"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/flow/transitions", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.RecordTransition(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_OverrideTransition_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := OverrideTransitionRequest{
		ToState: StateDischarged,
		Reason:  "Patient left AMA",
	}

	transition := &FlowStateTransition{
		ID:         "trans-1",
		ToState:    StateDischarged,
		IsOverride: true,
	}

	svc.On("OverrideTransition", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1", models.RoleAdmin).
		Return(transition, nil)

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/flow/override", reqBody, "user-1", models.RoleAdmin)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.OverrideTransition(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response FlowStateTransition
	testutil.DecodeJSON(t, rr, &response)
	assert.True(t, response.IsOverride)
	svc.AssertExpectations(t)
}

func TestHandler_OverrideTransition_Forbidden(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := OverrideTransitionRequest{
		ToState: StateDischarged,
		Reason:  "Override",
	}

	svc.On("OverrideTransition", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1", models.RoleNurse).
		Return(nil, fmt.Errorf("insufficient permissions to override flow transitions; requires admin or operations role"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/flow/override", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.OverrideTransition(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetCurrentState_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	triageState := StateTriage
	svc.On("GetCurrentState", mock.Anything, "encounter-1").Return(&triageState, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/flow/current", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetCurrentState(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "encounter-1", response["encounterId"])
	assert.Equal(t, string(triageState), response["currentState"])
	svc.AssertExpectations(t)
}

func TestHandler_GetCurrentState_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetCurrentState", mock.Anything, "encounter-1").Return(nil, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/flow/current", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetCurrentState(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "encounter-1", response["encounterId"])
	assert.Nil(t, response["currentState"])
	svc.AssertExpectations(t)
}

// Additional tests for improved coverage

func TestHandler_GetFlowTimeline_EncounterNotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetTimeline", mock.Anything, "encounter-999").
		Return(nil, fmt.Errorf("encounter not found"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-999/flow", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-999")

	rr := httptest.NewRecorder()
	handler.GetFlowTimeline(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetFlowTimeline_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetTimeline", mock.Anything, "encounter-1").
		Return(nil, fmt.Errorf("database error"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/flow", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetFlowTimeline(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetFlowTimeline_PaginatedWithOffset(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	triageState := StateTriage
	timeline := &FlowTimelineResponse{
		EncounterID:  "encounter-1",
		CurrentState: &triageState,
		Transitions:  []FlowStateTransition{{ID: "t1", ToState: triageState}},
		Total:        25,
	}

	svc.On("GetTimelinePaginated", mock.Anything, "encounter-1", 10, 10).Return(timeline, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/flow?paginated=true&limit=10&offset=10", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetFlowTimeline(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response FlowTimelineResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, int64(25), response.Total)
	svc.AssertExpectations(t)
}

func TestHandler_RecordTransition_EncounterNotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateTransitionRequest{
		ToState: StateTriage,
	}

	svc.On("RecordTransition", mock.Anything, mock.Anything, "encounter-999", reqBody, "user-1").
		Return(nil, fmt.Errorf("encounter not found"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-999/flow/transitions", reqBody, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-999")

	rr := httptest.NewRecorder()
	handler.RecordTransition(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_OverrideTransition_RoleNotAllowed(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := OverrideTransitionRequest{
		ToState: StateDischarged,
		Reason:  "Override",
	}

	svc.On("OverrideTransition", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1", models.RoleChargeNurse).
		Return(nil, fmt.Errorf("insufficient permissions to override flow transitions; requires admin or operations role"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/flow/override", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.OverrideTransition(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetCurrentState_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetCurrentState", mock.Anything, "encounter-1").
		Return(nil, fmt.Errorf("database error"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/flow/current", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetCurrentState(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}
