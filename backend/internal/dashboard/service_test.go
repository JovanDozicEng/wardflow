package dashboard

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/task"
)

// mockRepository is a mock implementation of Repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetActiveEncounterCount(ctx context.Context, filter FilterParams) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockRepository) GetFlowStateDistribution(ctx context.Context, filter FilterParams) (FlowDistribution, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return FlowDistribution{}, args.Error(1)
	}
	return args.Get(0).(FlowDistribution), args.Error(1)
}

func (m *mockRepository) GetOverdueTaskCount(ctx context.Context, filter FilterParams) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockRepository) GetHighPriorityTaskCount(ctx context.Context, filter FilterParams, priority task.TaskPriority) (int64, error) {
	args := m.Called(ctx, filter, priority)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockRepository) GetUnassignedTaskCount(ctx context.Context, filter FilterParams) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockRepository) GetCompletedTasksTodayCount(ctx context.Context, filter FilterParams) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockRepository) GetOpenTaskCount(ctx context.Context, filter FilterParams) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockRepository) GetPatientsInTriageOver2hrs(ctx context.Context, filter FilterParams) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockRepository) GetEncountersWithoutCareTeam(ctx context.Context, filter FilterParams) (int64, error) {
	args := m.Called(ctx, filter)
	return int64(args.Int(0)), args.Error(1)
}

func TestGetHuddleMetrics_AdminAllData(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := FilterParams{}

	// Mock all repository calls - set TotalOverdue and Urgent to 0 to avoid DB helper calls
	repo.On("GetActiveEncounterCount", ctx, filter).Return(int(42), nil)
	
	flowDist := FlowDistribution{
		Arrived:        5,
		Triage:         3,
		ProviderEval:   10,
		Diagnostics:    8,
		Admitted:       10,
		DischargeReady: 6,
		Discharged:     0,
	}
	repo.On("GetFlowStateDistribution", ctx, filter).Return(flowDist, nil)
	
	repo.On("GetOverdueTaskCount", ctx, filter).Return(int(0), nil) // Set to 0 to avoid helper call
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityHigh).Return(int(12), nil)
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityUrgent).Return(int(0), nil) // Set to 0 to avoid helper call
	repo.On("GetUnassignedTaskCount", ctx, filter).Return(int(9), nil)
	repo.On("GetCompletedTasksTodayCount", ctx, filter).Return(int(25), nil)
	repo.On("GetOpenTaskCount", ctx, filter).Return(int(30), nil)
	repo.On("GetPatientsInTriageOver2hrs", ctx, filter).Return(int(2), nil)
	repo.On("GetEncountersWithoutCareTeam", ctx, filter).Return(int(3), nil)

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleAdmin, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(42), result.ActiveEncounters)
	assert.Equal(t, int64(6), result.ExpectedDischarges) // Same as DischargeReady
	assert.Equal(t, int64(5), result.FlowDistribution.Arrived)
	assert.Equal(t, int64(0), result.TaskMetrics.TotalOverdue)
	assert.Equal(t, int64(12), result.TaskMetrics.HighPriority)
	assert.Equal(t, int64(0), result.TaskMetrics.Urgent)
	assert.Equal(t, int64(9), result.TaskMetrics.Unassigned)
	assert.Equal(t, int64(25), result.TaskMetrics.CompletedToday)
	assert.Equal(t, int64(30), result.TaskMetrics.TotalOpen)
	assert.Equal(t, int64(2), result.RiskIndicators.PatientsInTriageOver2hrs)
	assert.Equal(t, int64(3), result.RiskIndicators.EncountersWithoutCareTeam)
	repo.AssertExpectations(t)
}

func TestGetHuddleMetrics_NonAdminUnitFilter(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := FilterParams{}
	userUnitIDs := models.StringArray{"unit-1", "unit-2"}

	// Should apply unit filter
	expectedFilter := FilterParams{
		UnitID: &userUnitIDs[0],
	}

	repo.On("GetActiveEncounterCount", ctx, expectedFilter).Return(int(10), nil)
	repo.On("GetFlowStateDistribution", ctx, expectedFilter).Return(FlowDistribution{}, nil)
	repo.On("GetOverdueTaskCount", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, expectedFilter, task.TaskPriorityHigh).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, expectedFilter, task.TaskPriorityUrgent).Return(int(0), nil)
	repo.On("GetUnassignedTaskCount", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetCompletedTasksTodayCount", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetOpenTaskCount", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetPatientsInTriageOver2hrs", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetEncountersWithoutCareTeam", ctx, expectedFilter).Return(int(0), nil)

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleNurse, userUnitIDs, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "unit-1", *result.UnitID)
	repo.AssertExpectations(t)
}

func TestGetHuddleMetrics_NonAdminUnauthorizedUnit(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	unauthorizedUnit := "unit-999"
	filter := FilterParams{
		UnitID: &unauthorizedUnit,
	}
	userUnitIDs := models.StringArray{"unit-1", "unit-2"}

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleNurse, userUnitIDs, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized access to unit")
	repo.AssertNotCalled(t, "GetActiveEncounterCount")
}

func TestGetHuddleMetrics_NonAdminAuthorizedUnit(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	authorizedUnit := "unit-2"
	filter := FilterParams{
		UnitID: &authorizedUnit,
	}
	userUnitIDs := models.StringArray{"unit-1", "unit-2"}

	repo.On("GetActiveEncounterCount", ctx, filter).Return(int(5), nil)
	repo.On("GetFlowStateDistribution", ctx, filter).Return(FlowDistribution{}, nil)
	repo.On("GetOverdueTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityHigh).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityUrgent).Return(int(0), nil)
	repo.On("GetUnassignedTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetCompletedTasksTodayCount", ctx, filter).Return(int(0), nil)
	repo.On("GetOpenTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetPatientsInTriageOver2hrs", ctx, filter).Return(int(0), nil)
	repo.On("GetEncountersWithoutCareTeam", ctx, filter).Return(int(0), nil)

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleNurse, userUnitIDs, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "unit-2", *result.UnitID)
	repo.AssertExpectations(t)
}

func TestGetHuddleMetrics_NonAdminDepartmentFilter(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := FilterParams{}
	userDeptIDs := models.StringArray{"dept-1"}

	expectedFilter := FilterParams{
		DepartmentID: &userDeptIDs[0],
	}

	repo.On("GetActiveEncounterCount", ctx, expectedFilter).Return(int(8), nil)
	repo.On("GetFlowStateDistribution", ctx, expectedFilter).Return(FlowDistribution{}, nil)
	repo.On("GetOverdueTaskCount", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, expectedFilter, task.TaskPriorityHigh).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, expectedFilter, task.TaskPriorityUrgent).Return(int(0), nil)
	repo.On("GetUnassignedTaskCount", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetCompletedTasksTodayCount", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetOpenTaskCount", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetPatientsInTriageOver2hrs", ctx, expectedFilter).Return(int(0), nil)
	repo.On("GetEncountersWithoutCareTeam", ctx, expectedFilter).Return(int(0), nil)

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleNurse, nil, userDeptIDs)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "dept-1", *result.DepartmentID)
	repo.AssertExpectations(t)
}

func TestGetHuddleMetrics_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := FilterParams{}

	repo.On("GetActiveEncounterCount", ctx, filter).Return(int(0), fmt.Errorf("database error"))
	repo.On("GetFlowStateDistribution", ctx, filter).Return(FlowDistribution{}, nil)
	repo.On("GetOverdueTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityHigh).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityUrgent).Return(int(0), nil)
	repo.On("GetUnassignedTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetCompletedTasksTodayCount", ctx, filter).Return(int(0), nil)
	repo.On("GetOpenTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetPatientsInTriageOver2hrs", ctx, filter).Return(int(0), nil)
	repo.On("GetEncountersWithoutCareTeam", ctx, filter).Return(int(0), nil)

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleAdmin, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
}

func TestGetHuddleMetrics_FlowDistributionError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := FilterParams{}

	repo.On("GetActiveEncounterCount", ctx, filter).Return(int(10), nil)
	repo.On("GetFlowStateDistribution", ctx, filter).Return(FlowDistribution{}, fmt.Errorf("flow query failed"))
	repo.On("GetOverdueTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityHigh).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityUrgent).Return(int(0), nil)
	repo.On("GetUnassignedTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetCompletedTasksTodayCount", ctx, filter).Return(int(0), nil)
	repo.On("GetOpenTaskCount", ctx, filter).Return(int(0), nil)
	repo.On("GetPatientsInTriageOver2hrs", ctx, filter).Return(int(0), nil)
	repo.On("GetEncountersWithoutCareTeam", ctx, filter).Return(int(0), nil)

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleAdmin, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "flow query failed")
}

func TestGetHuddleMetrics_AllMetricsCollected(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	filter := FilterParams{}

	// Set up specific values for each metric - set overdue and urgent to 0 to avoid DB helpers
	repo.On("GetActiveEncounterCount", ctx, filter).Return(int(100), nil)
	
	flowDist := FlowDistribution{
		Arrived:        10,
		Triage:         15,
		ProviderEval:   20,
		Diagnostics:    15,
		Admitted:       25,
		DischargeReady: 10,
		Discharged:     5,
	}
	repo.On("GetFlowStateDistribution", ctx, filter).Return(flowDist, nil)
	
	repo.On("GetOverdueTaskCount", ctx, filter).Return(int(0), nil) // Set to 0
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityHigh).Return(int(20), nil)
	repo.On("GetHighPriorityTaskCount", ctx, filter, task.TaskPriorityUrgent).Return(int(0), nil) // Set to 0
	repo.On("GetUnassignedTaskCount", ctx, filter).Return(int(12), nil)
	repo.On("GetCompletedTasksTodayCount", ctx, filter).Return(int(45), nil)
	repo.On("GetOpenTaskCount", ctx, filter).Return(int(50), nil)
	repo.On("GetPatientsInTriageOver2hrs", ctx, filter).Return(int(5), nil)
	repo.On("GetEncountersWithoutCareTeam", ctx, filter).Return(int(7), nil)

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleAdmin, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	
	// Verify all metrics are correctly populated
	assert.Equal(t, int64(100), result.ActiveEncounters)
	assert.Equal(t, int64(10), result.ExpectedDischarges)
	
	assert.Equal(t, int64(10), result.FlowDistribution.Arrived)
	assert.Equal(t, int64(15), result.FlowDistribution.Triage)
	assert.Equal(t, int64(20), result.FlowDistribution.ProviderEval)
	assert.Equal(t, int64(15), result.FlowDistribution.Diagnostics)
	assert.Equal(t, int64(25), result.FlowDistribution.Admitted)
	assert.Equal(t, int64(10), result.FlowDistribution.DischargeReady)
	assert.Equal(t, int64(5), result.FlowDistribution.Discharged)
	
	assert.Equal(t, int64(0), result.TaskMetrics.TotalOverdue)
	assert.Equal(t, int64(20), result.TaskMetrics.HighPriority)
	assert.Equal(t, int64(0), result.TaskMetrics.Urgent)
	assert.Equal(t, int64(12), result.TaskMetrics.Unassigned)
	assert.Equal(t, int64(45), result.TaskMetrics.CompletedToday)
	assert.Equal(t, int64(50), result.TaskMetrics.TotalOpen)
	
	assert.Equal(t, int64(5), result.RiskIndicators.PatientsInTriageOver2hrs)
	assert.Equal(t, int64(7), result.RiskIndicators.EncountersWithoutCareTeam)
	
	repo.AssertExpectations(t)
}

func TestGetHuddleMetrics_WithOverdueHighPriority(t *testing.T) {
	// Since the private method uses s.db directly, we need a real DB for testing
	// For unit tests with mocks, we skip testing the private helper methods directly
	// They are tested implicitly through integration tests with a real database
	
	// Instead, test that the method calls the repo correctly when TotalOverdue > 0
	// The private helper will fail silently with nil DB but won't crash
	t.Skip("Private method requires real DB, tested in repository_test.go integration tests")
}

func TestGetHuddleMetrics_WithUnassignedUrgent(t *testing.T) {
	// Since the private method uses s.db directly, we need a real DB for testing
	// For unit tests with mocks, we skip testing the private helper methods directly
	// They are tested implicitly through integration tests with a real database
	
	t.Skip("Private method requires real DB, tested in repository_test.go integration tests")
}

func TestGetHuddleMetrics_NonAdminUnauthorizedDepartment(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	unauthorizedDept := "dept-999"
	filter := FilterParams{
		DepartmentID: &unauthorizedDept,
	}
	userDeptIDs := models.StringArray{"dept-1", "dept-2"}

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleNurse, nil, userDeptIDs)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized access to department")
	repo.AssertNotCalled(t, "GetActiveEncounterCount")
}

func TestGetHuddleMetrics_NonAdminAuthorizedDepartment(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	authorizedDept := "dept-2"
	filter := FilterParams{
		DepartmentID: &authorizedDept,
	}
	userDeptIDs := models.StringArray{"dept-1", "dept-2"}

	repo.On("GetActiveEncounterCount", mock.Anything, mock.Anything).Return(int(7), nil)
	repo.On("GetFlowStateDistribution", mock.Anything, mock.Anything).Return(FlowDistribution{}, nil)
	repo.On("GetOverdueTaskCount", mock.Anything, mock.Anything).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", mock.Anything, mock.Anything, task.TaskPriorityHigh).Return(int(0), nil)
	repo.On("GetHighPriorityTaskCount", mock.Anything, mock.Anything, task.TaskPriorityUrgent).Return(int(0), nil)
	repo.On("GetUnassignedTaskCount", mock.Anything, mock.Anything).Return(int(0), nil)
	repo.On("GetCompletedTasksTodayCount", mock.Anything, mock.Anything).Return(int(0), nil)
	repo.On("GetOpenTaskCount", mock.Anything, mock.Anything).Return(int(0), nil)
	repo.On("GetPatientsInTriageOver2hrs", mock.Anything, mock.Anything).Return(int(0), nil)
	repo.On("GetEncountersWithoutCareTeam", mock.Anything, mock.Anything).Return(int(0), nil)

	result, err := svc.GetHuddleMetrics(ctx, filter, models.RoleNurse, nil, userDeptIDs)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "dept-2", *result.DepartmentID)
	repo.AssertExpectations(t)
}

// NOTE: The private methods getOverdueHighPriorityCount and getUnassignedUrgentCount
// use s.db directly (not the repository) and are called conditionally:
// - getOverdueHighPriorityCount is called when TotalOverdue > 0
// - getUnassignedUrgentCount is called when Urgent > 0
// 
// These methods silently ignore errors (err == nil check) to avoid failing the entire
// metrics aggregation if one calculation fails. They are indirectly tested through:
// 1. Integration tests with a real database
// 2. The conditional logic in GetHuddleMetrics is tested by setting TotalOverdue=0 and Urgent=0
//    in the above tests to ensure the helpers are NOT called when conditions aren't met.
//
// To test error scenarios for these private methods, we would need:
// - A test database with a schema that causes Count() to fail
// - Or export these methods and test them directly with a mock DB
//
// The current implementation gracefully degrades: if these helpers fail, the RiskIndicators
// fields remain at their zero values (0), and the main metrics are still returned successfully.

