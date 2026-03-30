package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/task"
	"github.com/wardflow/backend/pkg/database"
)

// Service handles dashboard business logic
type Service struct {
	repo *Repository
	db   *database.DB
}

// NewService creates a new dashboard service
func NewService(db *database.DB) *Service {
	return &Service{
		repo: NewRepository(db),
		db:   db,
	}
}

// GetHuddleMetrics aggregates all metrics for the huddle dashboard
func (s *Service) GetHuddleMetrics(ctx context.Context, filter FilterParams, userRole models.Role, userUnitIDs, userDeptIDs models.StringArray) (*HuddleMetrics, error) {
	// Apply RBAC: non-admin users can only see their authorized units/departments
	if userRole != models.RoleAdmin {
		// If user has unit restrictions, apply them
		if len(userUnitIDs) > 0 {
			if filter.UnitID == nil {
				// No filter specified - restrict to first authorized unit
				filter.UnitID = &userUnitIDs[0]
			} else {
				// Verify user can access requested unit
				allowed := false
				for _, uid := range userUnitIDs {
					if uid == *filter.UnitID {
						allowed = true
						break
					}
				}
				if !allowed {
					return nil, fmt.Errorf("unauthorized access to unit")
				}
			}
		}

		// Similar for departments
		if len(userDeptIDs) > 0 {
			if filter.DepartmentID == nil {
				filter.DepartmentID = &userDeptIDs[0]
			} else {
				allowed := false
				for _, did := range userDeptIDs {
					if did == *filter.DepartmentID {
						allowed = true
						break
					}
				}
				if !allowed {
					return nil, fmt.Errorf("unauthorized access to department")
				}
			}
		}
	}

	metrics := &HuddleMetrics{
		UnitID:       filter.UnitID,
		DepartmentID: filter.DepartmentID,
		GeneratedAt:  time.Now().UTC(),
	}

	// Aggregate all metrics concurrently for better performance
	type metricResult struct {
		name  string
		value interface{}
		err   error
	}

	results := make(chan metricResult, 10)

	// Launch goroutines for each metric
	go func() {
		count, err := s.repo.GetActiveEncounterCount(ctx, filter)
		results <- metricResult{"activeEncounters", count, err}
	}()

	go func() {
		dist, err := s.repo.GetFlowStateDistribution(ctx, filter)
		results <- metricResult{"flowDistribution", dist, err}
	}()

	go func() {
		count, err := s.repo.GetOverdueTaskCount(ctx, filter)
		results <- metricResult{"overdueTaskCount", count, err}
	}()

	go func() {
		count, err := s.repo.GetHighPriorityTaskCount(ctx, filter, task.TaskPriorityHigh)
		results <- metricResult{"highPriorityCount", count, err}
	}()

	go func() {
		count, err := s.repo.GetHighPriorityTaskCount(ctx, filter, task.TaskPriorityUrgent)
		results <- metricResult{"urgentCount", count, err}
	}()

	go func() {
		count, err := s.repo.GetUnassignedTaskCount(ctx, filter)
		results <- metricResult{"unassignedCount", count, err}
	}()

	go func() {
		count, err := s.repo.GetCompletedTasksTodayCount(ctx, filter)
		results <- metricResult{"completedToday", count, err}
	}()

	go func() {
		count, err := s.repo.GetOpenTaskCount(ctx, filter)
		results <- metricResult{"openTaskCount", count, err}
	}()

	go func() {
		count, err := s.repo.GetPatientsInTriageOver2hrs(ctx, filter)
		results <- metricResult{"triageOver2hrs", count, err}
	}()

	go func() {
		count, err := s.repo.GetEncountersWithoutCareTeam(ctx, filter)
		results <- metricResult{"withoutCareTeam", count, err}
	}()

	// Collect results
	for i := 0; i < 10; i++ {
		result := <-results
		if result.err != nil {
			return nil, fmt.Errorf("failed to get %s: %w", result.name, result.err)
		}

		switch result.name {
		case "activeEncounters":
			metrics.ActiveEncounters = result.value.(int64)
		case "flowDistribution":
			metrics.FlowDistribution = result.value.(FlowDistribution)
			// Expected discharges = discharge_ready count
			metrics.ExpectedDischarges = metrics.FlowDistribution.DischargeReady
		case "overdueTaskCount":
			metrics.TaskMetrics.TotalOverdue = result.value.(int64)
		case "highPriorityCount":
			metrics.TaskMetrics.HighPriority = result.value.(int64)
		case "urgentCount":
			metrics.TaskMetrics.Urgent = result.value.(int64)
		case "unassignedCount":
			metrics.TaskMetrics.Unassigned = result.value.(int64)
		case "completedToday":
			metrics.TaskMetrics.CompletedToday = result.value.(int64)
		case "openTaskCount":
			metrics.TaskMetrics.TotalOpen = result.value.(int64)
		case "triageOver2hrs":
			metrics.RiskIndicators.PatientsInTriageOver2hrs = result.value.(int64)
		case "withoutCareTeam":
			metrics.RiskIndicators.EncountersWithoutCareTeam = result.value.(int64)
		}
	}

	// Calculate additional risk indicators
	// Overdue high priority tasks
	if metrics.TaskMetrics.TotalOverdue > 0 {
		// Get count of overdue high/urgent priority tasks
		highPriorityOverdue, err := s.getOverdueHighPriorityCount(ctx, filter)
		if err == nil {
			metrics.RiskIndicators.OverdueHighPriorityTasks = highPriorityOverdue
		}
	}

	// Unassigned urgent tasks
	if metrics.TaskMetrics.Urgent > 0 {
		unassignedUrgent, err := s.getUnassignedUrgentCount(ctx, filter)
		if err == nil {
			metrics.RiskIndicators.UnassignedUrgentTasks = unassignedUrgent
		}
	}

	// TODO: Add drill-down lists (overdueTasks, longStayPatients, pendingDischarges)
	// These can be added as separate endpoints or included here with a flag

	return metrics, nil
}

// Helper methods for specific calculations

func (s *Service) getOverdueHighPriorityCount(ctx context.Context, filter FilterParams) (int64, error) {
	query := s.db.WithContext(ctx).Model(&task.Task{}).
		Where("sla_due_at IS NOT NULL").
		Where("sla_due_at < ?", time.Now().UTC()).
		Where("status NOT IN (?, ?)", task.TaskStatusCompleted, task.TaskStatusCancelled).
		Where("priority IN (?, ?)", task.TaskPriorityHigh, task.TaskPriorityUrgent)

	if filter.UnitID != nil || filter.DepartmentID != nil {
		query = query.Joins("LEFT JOIN encounters ON tasks.scope_type = ? AND tasks.scope_id = encounters.id", task.ScopeTypeEncounter)
		if filter.UnitID != nil {
			query = query.Where("encounters.unit_id = ? OR tasks.scope_type != ?", *filter.UnitID, task.ScopeTypeEncounter)
		}
		if filter.DepartmentID != nil {
			query = query.Where("encounters.department_id = ? OR tasks.scope_type != ?", *filter.DepartmentID, task.ScopeTypeEncounter)
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Service) getUnassignedUrgentCount(ctx context.Context, filter FilterParams) (int64, error) {
	query := s.db.WithContext(ctx).Model(&task.Task{}).
		Where("current_owner_id IS NULL").
		Where("status IN (?, ?)", task.TaskStatusOpen, task.TaskStatusInProgress).
		Where("priority = ?", task.TaskPriorityUrgent)

	if filter.UnitID != nil || filter.DepartmentID != nil {
		query = query.Joins("LEFT JOIN encounters ON tasks.scope_type = ? AND tasks.scope_id = encounters.id", task.ScopeTypeEncounter)
		if filter.UnitID != nil {
			query = query.Where("encounters.unit_id = ? OR tasks.scope_type != ?", *filter.UnitID, task.ScopeTypeEncounter)
		}
		if filter.DepartmentID != nil {
			query = query.Where("encounters.department_id = ? OR tasks.scope_type != ?", *filter.DepartmentID, task.ScopeTypeEncounter)
		}
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
