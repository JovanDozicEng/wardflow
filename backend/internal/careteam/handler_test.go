package careteam

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
)

// mockService is a mock implementation of Service
type mockService struct {
	mock.Mock
}

func (m *mockService) AssignRole(ctx context.Context, r *http.Request, encounterID string, req AssignRoleRequest, currentUserID string) (*CareTeamAssignment, error) {
	args := m.Called(ctx, r, encounterID, req, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CareTeamAssignment), args.Error(1)
}

func (m *mockService) TransferRole(ctx context.Context, r *http.Request, assignmentID string, req TransferRoleRequest, currentUserID string) (*CareTeamAssignment, error) {
	args := m.Called(ctx, r, assignmentID, req, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CareTeamAssignment), args.Error(1)
}

func (m *mockService) ListCareTeam(ctx context.Context, encounterID string, activeOnly bool) ([]CareTeamAssignment, error) {
	args := m.Called(ctx, encounterID, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]CareTeamAssignment), args.Error(1)
}

func (m *mockService) GetHandoffs(ctx context.Context, encounterID string, limit, offset int) ([]HandoffNote, int64, error) {
	args := m.Called(ctx, encounterID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]HandoffNote), args.Get(1).(int64), args.Error(2)
}

func (m *mockService) GetCareTeamWithDetails(ctx context.Context, encounterID string) (*CareTeamResponse, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CareTeamResponse), args.Error(1)
}

func TestHandler_GetCareTeam_ActiveOnly(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	activeAssignments := []CareTeamAssignment{
		{ID: "a1", UserID: "user-1", RoleType: RolePrimaryNurse},
		{ID: "a2", UserID: "user-2", RoleType: RoleAttendingProvider},
	}

	svc.On("ListCareTeam", mock.Anything, "encounter-1", true).Return(activeAssignments, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/care-team?activeOnly=true", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetCareTeam(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListAssignmentsResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Len(t, response.Assignments, 2)
	assert.Equal(t, int64(2), response.Total)
	svc.AssertExpectations(t)
}

func TestHandler_GetCareTeam_WithDetails(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	careTeamResponse := &CareTeamResponse{
		EncounterID: "encounter-1",
		Members: []CareTeamMember{
			{
				Assignment: CareTeamAssignment{ID: "a1", UserID: "user-1", RoleType: RolePrimaryNurse},
				UserName:   "John Doe",
				UserEmail:  "john@example.com",
			},
		},
	}

	svc.On("GetCareTeamWithDetails", mock.Anything, "encounter-1").Return(careTeamResponse, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/care-team?withDetails=true", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetCareTeam(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response CareTeamResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "encounter-1", response.EncounterID)
	assert.Len(t, response.Members, 1)
	assert.Equal(t, "John Doe", response.Members[0].UserName)
	svc.AssertExpectations(t)
}

func TestHandler_GetCareTeam_MissingEncounterID(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters//care-team", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "")

	rr := httptest.NewRecorder()
	handler.GetCareTeam(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_AssignRole_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := AssignRoleRequest{
		UserID:   "user-2",
		RoleType: RolePrimaryNurse,
	}

	assignment := &CareTeamAssignment{
		ID:          "assignment-1",
		EncounterID: "encounter-1",
		UserID:      "user-2",
		RoleType:    RolePrimaryNurse,
		StartsAt:    time.Now(),
		CreatedBy:   "user-1",
	}

	svc.On("AssignRole", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1").Return(assignment, nil)

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/care-team/assignments", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.AssignRole(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response CareTeamAssignment
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "assignment-1", response.ID)
	assert.Equal(t, "user-2", response.UserID)
	svc.AssertExpectations(t)
}

func TestHandler_AssignRole_AlreadyAssigned(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := AssignRoleRequest{
		UserID:   "user-2",
		RoleType: RolePrimaryNurse,
	}

	svc.On("AssignRole", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1").
		Return(nil, fmt.Errorf("role primary_nurse is already assigned"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/care-team/assignments", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.AssignRole(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_TransferRole_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := TransferRoleRequest{
		ToUserID:    "user-3",
		HandoffNote: "Patient is stable",
	}

	newAssignment := &CareTeamAssignment{
		ID:          "assignment-2",
		EncounterID: "encounter-1",
		UserID:      "user-3",
		RoleType:    RolePrimaryNurse,
		StartsAt:    time.Now(),
		CreatedBy:   "user-1",
	}

	svc.On("TransferRole", mock.Anything, mock.Anything, "assignment-1", reqBody, "user-1").Return(newAssignment, nil)

	r := testutil.NewRequest(http.MethodPost, "/api/v1/care-team/assignments/assignment-1/transfer", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("assignmentId", "assignment-1")

	rr := httptest.NewRecorder()
	handler.TransferRole(rr, r)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response CareTeamAssignment
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "assignment-2", response.ID)
	assert.Equal(t, "user-3", response.UserID)
	svc.AssertExpectations(t)
}

func TestHandler_TransferRole_MissingAssignmentID(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := TransferRoleRequest{
		ToUserID:    "user-3",
		HandoffNote: "Transfer",
	}

	r := testutil.NewRequest(http.MethodPost, "/api/v1/care-team/assignments//transfer", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("assignmentId", "")

	rr := httptest.NewRecorder()
	handler.TransferRole(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_GetHandoffs_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	handoffs := []HandoffNote{
		{ID: "h1", Note: "First handoff"},
		{ID: "h2", Note: "Second handoff"},
	}

	svc.On("GetHandoffs", mock.Anything, "encounter-1", 30, 0).Return(handoffs, int64(2), nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/handoffs", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetHandoffs(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListHandoffsResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Len(t, response.Handoffs, 2)
	assert.Equal(t, int64(2), response.Total)
	svc.AssertExpectations(t)
}

func TestHandler_GetHandoffs_WithPagination(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	handoffs := []HandoffNote{
		{ID: "h1", Note: "First handoff"},
	}

	svc.On("GetHandoffs", mock.Anything, "encounter-1", 10, 5).Return(handoffs, int64(15), nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-1/handoffs?limit=10&offset=5", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.GetHandoffs(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response ListHandoffsResponse
	testutil.DecodeJSON(t, rr, &response)
	assert.Len(t, response.Handoffs, 1)
	assert.Equal(t, int64(15), response.Total)
	svc.AssertExpectations(t)
}

// Additional tests for improved coverage

func TestHandler_GetCareTeam_EncounterNotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("ListCareTeam", mock.Anything, "encounter-999", false).
		Return(nil, fmt.Errorf("encounter not found"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-999/care-team", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-999")

	rr := httptest.NewRecorder()
	handler.GetCareTeam(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_AssignRole_MissingRole(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := AssignRoleRequest{
		UserID: "user-2",
		// Missing RoleType
	}

	svc.On("AssignRole", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1").
		Return(nil, fmt.Errorf("roleType is required"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/care-team/assignments", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.AssignRole(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_AssignRole_MissingUserID(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := AssignRoleRequest{
		RoleType: RolePrimaryNurse,
		// Missing UserID
	}

	svc.On("AssignRole", mock.Anything, mock.Anything, "encounter-1", reqBody, "user-1").
		Return(nil, fmt.Errorf("userId is required"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/encounter-1/care-team/assignments", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("encounterId", "encounter-1")

	rr := httptest.NewRecorder()
	handler.AssignRole(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_TransferRole_AssignmentNotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := TransferRoleRequest{
		ToUserID:    "user-3",
		HandoffNote: "Transfer",
	}

	svc.On("TransferRole", mock.Anything, mock.Anything, "assignment-999", reqBody, "user-1").
		Return(nil, fmt.Errorf("assignment not found"))

	r := testutil.NewRequest(http.MethodPost, "/api/v1/care-team/assignments/assignment-999/transfer", reqBody, "user-1", models.RoleChargeNurse)
	r.SetPathValue("assignmentId", "assignment-999")

	rr := httptest.NewRecorder()
	handler.TransferRole(rr, r)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetHandoffs_EncounterNotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetHandoffs", mock.Anything, "encounter-999", 30, 0).
		Return(nil, int64(0), fmt.Errorf("encounter not found"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/encounter-999/handoffs", nil, "user-1", models.RoleNurse)
	r.SetPathValue("encounterId", "encounter-999")

	rr := httptest.NewRecorder()
	handler.GetHandoffs(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}
