package flow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
)

// mockRepository is a mock implementation of Repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateTransition(ctx context.Context, transition *FlowStateTransition) error {
	args := m.Called(ctx, transition)
	return args.Error(0)
}

func (m *mockRepository) GetCurrentState(ctx context.Context, encounterID string) (*FlowStateTransition, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FlowStateTransition), args.Error(1)
}

func (m *mockRepository) GetTimeline(ctx context.Context, encounterID string) ([]FlowStateTransition, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]FlowStateTransition), args.Error(1)
}

func (m *mockRepository) GetTimelinePaginated(ctx context.Context, encounterID string, limit, offset int) ([]FlowStateTransition, int64, error) {
	args := m.Called(ctx, encounterID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]FlowStateTransition), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) GetTransitionByID(ctx context.Context, transitionID string) (*FlowStateTransition, error) {
	args := m.Called(ctx, transitionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FlowStateTransition), args.Error(1)
}

func (m *mockRepository) GetTransitionsSince(ctx context.Context, encounterID string, since time.Time) ([]FlowStateTransition, error) {
	args := m.Called(ctx, encounterID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]FlowStateTransition), args.Error(1)
}

func (m *mockRepository) GetTransitionsByState(ctx context.Context, encounterID string, state FlowState) ([]FlowStateTransition, error) {
	args := m.Called(ctx, encounterID, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]FlowStateTransition), args.Error(1)
}

func TestRecordTransition_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	// Current state is Arrived
	arrivedState := StateArrived
	currentTransition := &FlowStateTransition{
		ID:          "trans-1",
		EncounterID: "encounter-1",
		ToState:     arrivedState,
	}

	req := CreateTransitionRequest{
		ToState: StateTriage,
	}

	repo.On("GetCurrentState", ctx, "encounter-1").Return(currentTransition, nil)
	repo.On("CreateTransition", ctx, mock.MatchedBy(func(t *FlowStateTransition) bool {
		return t.EncounterID == "encounter-1" &&
			t.ToState == StateTriage &&
			*t.FromState == StateArrived &&
			t.ActorType == ActorTypeUser &&
			!t.IsOverride
	})).Return(nil)

	result, err := svc.RecordTransition(ctx, r, "encounter-1", req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StateTriage, result.ToState)
	assert.Equal(t, StateArrived, *result.FromState)
	repo.AssertExpectations(t)
}

func TestRecordTransition_InvalidTransition(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	// Current state is Arrived
	arrivedState := StateArrived
	currentTransition := &FlowStateTransition{
		ID:          "trans-1",
		EncounterID: "encounter-1",
		ToState:     arrivedState,
	}

	req := CreateTransitionRequest{
		ToState: StateDischarged, // Invalid: Arrived -> Discharged
	}

	repo.On("GetCurrentState", ctx, "encounter-1").Return(currentTransition, nil)

	result, err := svc.RecordTransition(ctx, r, "encounter-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid transition")
	repo.AssertExpectations(t)
}

func TestRecordTransition_InitialState(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	// No current state (first transition)
	repo.On("GetCurrentState", ctx, "encounter-1").Return(nil, nil)

	req := CreateTransitionRequest{
		ToState: StateArrived,
	}

	repo.On("CreateTransition", ctx, mock.MatchedBy(func(t *FlowStateTransition) bool {
		return t.EncounterID == "encounter-1" &&
			t.ToState == StateArrived &&
			t.FromState == nil // Initial state
	})).Return(nil)

	result, err := svc.RecordTransition(ctx, r, "encounter-1", req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StateArrived, result.ToState)
	assert.Nil(t, result.FromState)
	repo.AssertExpectations(t)
}

func TestOverrideTransition_AdminSuccess(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	arrivedState := StateArrived
	currentTransition := &FlowStateTransition{
		ID:          "trans-1",
		EncounterID: "encounter-1",
		ToState:     arrivedState,
	}

	req := OverrideTransitionRequest{
		ToState: StateDischarged, // Invalid transition, but override
		Reason:  "Patient left AMA",
	}

	repo.On("GetCurrentState", ctx, "encounter-1").Return(currentTransition, nil)
	repo.On("CreateTransition", ctx, mock.MatchedBy(func(t *FlowStateTransition) bool {
		return t.EncounterID == "encounter-1" &&
			t.ToState == StateDischarged &&
			*t.FromState == StateArrived &&
			t.IsOverride
	})).Return(nil)

	result, err := svc.OverrideTransition(ctx, r, "encounter-1", req, "user-1", models.RoleAdmin)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StateDischarged, result.ToState)
	assert.True(t, result.IsOverride)
	repo.AssertExpectations(t)
}

func TestOverrideTransition_OperationsSuccess(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	triageState := StateTriage
	currentTransition := &FlowStateTransition{
		ID:          "trans-1",
		EncounterID: "encounter-1",
		ToState:     triageState,
	}

	req := OverrideTransitionRequest{
		ToState: StateArrived, // Backwards transition
		Reason:  "Correction needed",
	}

	repo.On("GetCurrentState", ctx, "encounter-1").Return(currentTransition, nil)
	repo.On("CreateTransition", ctx, mock.Anything).Return(nil)

	result, err := svc.OverrideTransition(ctx, r, "encounter-1", req, "user-1", models.RoleOperations)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsOverride)
	repo.AssertExpectations(t)
}

func TestOverrideTransition_InsufficientPermissions(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	req := OverrideTransitionRequest{
		ToState: StateDischarged,
		Reason:  "Override needed",
	}

	result, err := svc.OverrideTransition(ctx, r, "encounter-1", req, "user-1", models.RoleNurse)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insufficient permissions")
	repo.AssertNotCalled(t, "GetCurrentState")
}

func TestOverrideTransition_MissingReason(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	req := OverrideTransitionRequest{
		ToState: StateDischarged,
		Reason:  "", // Empty reason
	}

	result, err := svc.OverrideTransition(ctx, r, "encounter-1", req, "user-1", models.RoleAdmin)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "reason is required")
	repo.AssertNotCalled(t, "CreateTransition")
}

func TestGetTimeline_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	transitions := []FlowStateTransition{
		{ID: "t1", ToState: StateArrived},
		{ID: "t2", ToState: StateTriage},
		{ID: "t3", ToState: StateProviderEval},
	}

	repo.On("GetTimeline", ctx, "encounter-1").Return(transitions, nil)

	result, err := svc.GetTimeline(ctx, "encounter-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "encounter-1", result.EncounterID)
	assert.Len(t, result.Transitions, 3)
	assert.Equal(t, StateProviderEval, *result.CurrentState)
	assert.Equal(t, int64(3), result.Total)
	repo.AssertExpectations(t)
}

func TestGetTimeline_Empty(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	repo.On("GetTimeline", ctx, "encounter-1").Return([]FlowStateTransition{}, nil)

	result, err := svc.GetTimeline(ctx, "encounter-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Transitions, 0)
	assert.Nil(t, result.CurrentState)
	repo.AssertExpectations(t)
}

func TestGetTimelinePaginated_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	transitions := []FlowStateTransition{
		{ID: "t1", ToState: StateArrived},
		{ID: "t2", ToState: StateTriage},
	}

	currentTransition := &FlowStateTransition{
		ID:      "t3",
		ToState: StateProviderEval,
	}

	repo.On("GetTimelinePaginated", ctx, "encounter-1", 10, 0).Return(transitions, int64(5), nil)
	repo.On("GetCurrentState", ctx, "encounter-1").Return(currentTransition, nil)

	result, err := svc.GetTimelinePaginated(ctx, "encounter-1", 10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Transitions, 2)
	assert.Equal(t, StateProviderEval, *result.CurrentState)
	assert.Equal(t, int64(5), result.Total)
	repo.AssertExpectations(t)
}

func TestGetCurrentState_Exists(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	currentTransition := &FlowStateTransition{
		ID:      "trans-1",
		ToState: StateTriage,
	}

	repo.On("GetCurrentState", ctx, "encounter-1").Return(currentTransition, nil)

	result, err := svc.GetCurrentState(ctx, "encounter-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StateTriage, *result)
	repo.AssertExpectations(t)
}

func TestGetCurrentState_NotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	repo.On("GetCurrentState", ctx, "encounter-1").Return(nil, nil)

	result, err := svc.GetCurrentState(ctx, "encounter-1")

	assert.NoError(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestGetTimelineWithActors_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	// Return empty timeline to avoid DB lookup
	transitions := []FlowStateTransition{}

	repo.On("GetTimeline", ctx, "encounter-1").Return(transitions, nil)

	result, err := svc.GetTimelineWithActors(ctx, "encounter-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "encounter-1", result.EncounterID)
	assert.Len(t, result.Transitions, 0)
	repo.AssertExpectations(t)
}

func TestGetTimelineWithActors_ErrorFromRepo(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	repo.On("GetTimeline", ctx, "encounter-1").Return(nil, assert.AnError)

	result, err := svc.GetTimelineWithActors(ctx, "encounter-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestGetTimelineWithActors_WithSystemActor(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	now := time.Now().UTC()
	fromState := StateArrived
	transitions := []FlowStateTransition{
		{
			ID:             "trans-1",
			EncounterID:    "encounter-1",
			FromState:      &fromState,
			ToState:        StateTriage,
			TransitionedAt: now,
			ActorType:      ActorTypeSystem,
			ActorUserID:    nil,
		},
	}

	repo.On("GetTimeline", ctx, "encounter-1").Return(transitions, nil)

	result, err := svc.GetTimelineWithActors(ctx, "encounter-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Transitions, 1)
	assert.Equal(t, ActorTypeSystem, result.Transitions[0].ActorType)
	assert.Nil(t, result.Transitions[0].ActorName)
	assert.Nil(t, result.Transitions[0].ActorEmail)
	repo.AssertExpectations(t)
}
