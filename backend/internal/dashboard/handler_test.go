package dashboard

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

func (m *mockService) GetHuddleMetrics(ctx context.Context, filter FilterParams, userRole models.Role, userUnitIDs, userDeptIDs models.StringArray) (*HuddleMetrics, error) {
	args := m.Called(ctx, filter, userRole, userUnitIDs, userDeptIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*HuddleMetrics), args.Error(1)
}

func TestHandler_GetHuddleDashboard_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc)

	metrics := &HuddleMetrics{
		GeneratedAt:      time.Now(),
		ActiveEncounters: 42,
		FlowDistribution: FlowDistribution{
			Arrived:        5,
			Triage:         3,
			ProviderEval:   10,
			DischargeReady: 6,
		},
		TaskMetrics: TaskMetrics{
			TotalOpen:      30,
			TotalOverdue:   7,
			HighPriority:   12,
			Urgent:         4,
			Unassigned:     9,
			CompletedToday: 25,
		},
		RiskIndicators: RiskIndicators{
			PatientsInTriageOver2hrs:  2,
			EncountersWithoutCareTeam: 3,
		},
	}

	svc.On("GetHuddleMetrics", mock.Anything, mock.MatchedBy(func(f FilterParams) bool {
		return f.UnitID == nil && f.DepartmentID == nil
	}), models.RoleAdmin, models.StringArray(nil), models.StringArray(nil)).Return(metrics, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/dashboard/huddle", nil, "user-1", models.RoleAdmin)

	rr := httptest.NewRecorder()
	handler.GetHuddleDashboard(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response HuddleMetrics
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, int64(42), response.ActiveEncounters)
	assert.Equal(t, int64(30), response.TaskMetrics.TotalOpen)
	assert.Equal(t, int64(2), response.RiskIndicators.PatientsInTriageOver2hrs)
	svc.AssertExpectations(t)
}

func TestHandler_GetHuddleDashboard_WithUnitFilter(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc)

	unitID := "unit-1"
	metrics := &HuddleMetrics{
		UnitID:           &unitID,
		GeneratedAt:      time.Now(),
		ActiveEncounters: 10,
		TaskMetrics: TaskMetrics{
			TotalOpen: 5,
		},
	}

	svc.On("GetHuddleMetrics", mock.Anything, mock.MatchedBy(func(f FilterParams) bool {
		return f.UnitID != nil && *f.UnitID == "unit-1"
	}), models.RoleNurse, models.StringArray(nil), models.StringArray(nil)).Return(metrics, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/dashboard/huddle?unitId=unit-1", nil, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.GetHuddleDashboard(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response HuddleMetrics
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "unit-1", *response.UnitID)
	assert.Equal(t, int64(10), response.ActiveEncounters)
	svc.AssertExpectations(t)
}

func TestHandler_GetHuddleDashboard_WithDepartmentFilter(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc)

	deptID := "dept-1"
	metrics := &HuddleMetrics{
		DepartmentID:     &deptID,
		GeneratedAt:      time.Now(),
		ActiveEncounters: 15,
	}

	svc.On("GetHuddleMetrics", mock.Anything, mock.MatchedBy(func(f FilterParams) bool {
		return f.DepartmentID != nil && *f.DepartmentID == "dept-1"
	}), models.RoleNurse, models.StringArray(nil), models.StringArray(nil)).Return(metrics, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/dashboard/huddle?departmentId=dept-1", nil, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.GetHuddleDashboard(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response HuddleMetrics
	testutil.DecodeJSON(t, rr, &response)
	assert.Equal(t, "dept-1", *response.DepartmentID)
	svc.AssertExpectations(t)
}

func TestHandler_GetHuddleDashboard_UnauthorizedUnit(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc)

	svc.On("GetHuddleMetrics", mock.Anything, mock.Anything, models.RoleNurse, models.StringArray(nil), models.StringArray(nil)).
		Return(nil, fmt.Errorf("unauthorized access to unit"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/dashboard/huddle?unitId=unit-999", nil, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.GetHuddleDashboard(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetHuddleDashboard_UnauthorizedDepartment(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc)

	svc.On("GetHuddleMetrics", mock.Anything, mock.Anything, models.RoleNurse, models.StringArray(nil), models.StringArray(nil)).
		Return(nil, fmt.Errorf("unauthorized access to department"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/dashboard/huddle?departmentId=dept-999", nil, "user-1", models.RoleNurse)

	rr := httptest.NewRecorder()
	handler.GetHuddleDashboard(rr, r)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetHuddleDashboard_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc)

	svc.On("GetHuddleMetrics", mock.Anything, mock.Anything, models.RoleAdmin, models.StringArray(nil), models.StringArray(nil)).
		Return(nil, fmt.Errorf("database connection failed"))

	r := testutil.NewRequest(http.MethodGet, "/api/v1/dashboard/huddle", nil, "user-1", models.RoleAdmin)

	rr := httptest.NewRecorder()
	handler.GetHuddleDashboard(rr, r)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetHuddleDashboard_CompleteMetrics(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc)

	metrics := &HuddleMetrics{
		GeneratedAt:        time.Now(),
		ActiveEncounters:   100,
		ExpectedDischarges: 10,
		ExpectedAdmissions: 5,
		FlowDistribution: FlowDistribution{
			Arrived:        10,
			Triage:         15,
			ProviderEval:   20,
			Diagnostics:    15,
			Admitted:       25,
			DischargeReady: 10,
			Discharged:     5,
		},
		RiskIndicators: RiskIndicators{
			PatientsInTriageOver2hrs:          5,
			PatientsWaitingForBedOver4hrs:     3,
			OverdueHighPriorityTasks:          8,
			UnassignedUrgentTasks:             4,
			EncountersWithoutCareTeam:         7,
		},
		TaskMetrics: TaskMetrics{
			TotalOpen:      50,
			TotalOverdue:   15,
			HighPriority:   20,
			Urgent:         8,
			Unassigned:     12,
			CompletedToday: 45,
		},
	}

	svc.On("GetHuddleMetrics", mock.Anything, mock.Anything, models.RoleAdmin, models.StringArray(nil), models.StringArray(nil)).
		Return(metrics, nil)

	r := testutil.NewRequest(http.MethodGet, "/api/v1/dashboard/huddle", nil, "user-1", models.RoleAdmin)

	rr := httptest.NewRecorder()
	handler.GetHuddleDashboard(rr, r)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response HuddleMetrics
	testutil.DecodeJSON(t, rr, &response)
	
	// Verify all metrics are present
	assert.Equal(t, int64(100), response.ActiveEncounters)
	assert.Equal(t, int64(10), response.ExpectedDischarges)
	
	assert.Equal(t, int64(10), response.FlowDistribution.Arrived)
	assert.Equal(t, int64(15), response.FlowDistribution.Triage)
	assert.Equal(t, int64(20), response.FlowDistribution.ProviderEval)
	assert.Equal(t, int64(15), response.FlowDistribution.Diagnostics)
	assert.Equal(t, int64(25), response.FlowDistribution.Admitted)
	assert.Equal(t, int64(10), response.FlowDistribution.DischargeReady)
	assert.Equal(t, int64(5), response.FlowDistribution.Discharged)
	
	assert.Equal(t, int64(5), response.RiskIndicators.PatientsInTriageOver2hrs)
	assert.Equal(t, int64(3), response.RiskIndicators.PatientsWaitingForBedOver4hrs)
	assert.Equal(t, int64(8), response.RiskIndicators.OverdueHighPriorityTasks)
	assert.Equal(t, int64(4), response.RiskIndicators.UnassignedUrgentTasks)
	assert.Equal(t, int64(7), response.RiskIndicators.EncountersWithoutCareTeam)
	
	assert.Equal(t, int64(50), response.TaskMetrics.TotalOpen)
	assert.Equal(t, int64(15), response.TaskMetrics.TotalOverdue)
	assert.Equal(t, int64(20), response.TaskMetrics.HighPriority)
	assert.Equal(t, int64(8), response.TaskMetrics.Urgent)
	assert.Equal(t, int64(12), response.TaskMetrics.Unassigned)
	assert.Equal(t, int64(45), response.TaskMetrics.CompletedToday)
	
	svc.AssertExpectations(t)
}
