package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/wardflow/backend/internal/encounter"
	"github.com/wardflow/backend/internal/flow"
	"github.com/wardflow/backend/internal/task"
	"github.com/wardflow/backend/pkg/database"
)

// Repository defines the data access interface for dashboard
type Repository interface {
	GetActiveEncounterCount(ctx context.Context, filter FilterParams) (int64, error)
	GetFlowStateDistribution(ctx context.Context, filter FilterParams) (FlowDistribution, error)
	GetOverdueTaskCount(ctx context.Context, filter FilterParams) (int64, error)
	GetHighPriorityTaskCount(ctx context.Context, filter FilterParams, priority task.TaskPriority) (int64, error)
	GetUnassignedTaskCount(ctx context.Context, filter FilterParams) (int64, error)
	GetCompletedTasksTodayCount(ctx context.Context, filter FilterParams) (int64, error)
	GetOpenTaskCount(ctx context.Context, filter FilterParams) (int64, error)
	GetPatientsInTriageOver2hrs(ctx context.Context, filter FilterParams) (int64, error)
	GetEncountersWithoutCareTeam(ctx context.Context, filter FilterParams) (int64, error)
}

// repository handles data aggregation for dashboard
type repository struct {
	db *database.DB
}

// NewRepository creates a new dashboard repository
func NewRepository(db *database.DB) Repository {
	return &repository{db: db}
}

// GetActiveEncounterCount returns count of active encounters
func (r *repository) GetActiveEncounterCount(ctx context.Context, filter FilterParams) (int64, error) {
	query := r.db.WithContext(ctx).Model(&encounter.Encounter{}).
		Where("status = ?", encounter.EncounterStatusActive)

	if filter.UnitID != nil {
		query = query.Where("unit_id = ?", *filter.UnitID)
	}
	if filter.DepartmentID != nil {
		query = query.Where("department_id = ?", *filter.DepartmentID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active encounters: %w", err)
	}
	return count, nil
}

// GetFlowStateDistribution returns count of encounters by flow state
func (r *repository) GetFlowStateDistribution(ctx context.Context, filter FilterParams) (FlowDistribution, error) {
	dist := FlowDistribution{}

	// Query to get latest state per encounter
	// This uses a subquery to get the most recent transition for each encounter
	subquery := r.db.WithContext(ctx).
		Model(&flow.FlowStateTransition{}).
		Select("encounter_id, to_state, ROW_NUMBER() OVER (PARTITION BY encounter_id ORDER BY transitioned_at DESC, created_at DESC) as rn").
		Where("encounter_id IN (?)",
			r.db.Model(&encounter.Encounter{}).
				Select("id").
				Where("status = ?", encounter.EncounterStatusActive))

	if filter.UnitID != nil || filter.DepartmentID != nil {
		encounterQuery := r.db.Model(&encounter.Encounter{}).Select("id")
		if filter.UnitID != nil {
			encounterQuery = encounterQuery.Where("unit_id = ?", *filter.UnitID)
		}
		if filter.DepartmentID != nil {
			encounterQuery = encounterQuery.Where("department_id = ?", *filter.DepartmentID)
		}
		subquery = subquery.Where("encounter_id IN (?)", encounterQuery)
	}

	// Count by state
	type stateCount struct {
		ToState flow.FlowState
		Count   int64
	}

	var counts []stateCount
	err := r.db.WithContext(ctx).
		Table("(?) as latest_states", subquery).
		Select("to_state, COUNT(*) as count").
		Where("rn = 1").
		Group("to_state").
		Scan(&counts).Error

	if err != nil {
		return dist, fmt.Errorf("failed to get flow distribution: %w", err)
	}

	// Map to distribution struct
	for _, c := range counts {
		switch c.ToState {
		case flow.StateArrived:
			dist.Arrived = c.Count
		case flow.StateTriage:
			dist.Triage = c.Count
		case flow.StateProviderEval:
			dist.ProviderEval = c.Count
		case flow.StateDiagnostics:
			dist.Diagnostics = c.Count
		case flow.StateAdmitted:
			dist.Admitted = c.Count
		case flow.StateDischargeReady:
			dist.DischargeReady = c.Count
		case flow.StateDischarged:
			dist.Discharged = c.Count
		}
	}

	return dist, nil
}

// GetOverdueTaskCount returns count of overdue tasks
func (r *repository) GetOverdueTaskCount(ctx context.Context, filter FilterParams) (int64, error) {
	query := r.db.WithContext(ctx).Model(&task.Task{}).
		Where("sla_due_at IS NOT NULL").
		Where("sla_due_at < ?", time.Now().UTC()).
		Where("status NOT IN (?, ?)", task.TaskStatusCompleted, task.TaskStatusCancelled)

	// Join with encounters to filter by unit/department if needed
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
		return 0, fmt.Errorf("failed to count overdue tasks: %w", err)
	}
	return count, nil
}

// GetHighPriorityTaskCount returns count of high priority open tasks
func (r *repository) GetHighPriorityTaskCount(ctx context.Context, filter FilterParams, priority task.TaskPriority) (int64, error) {
	query := r.db.WithContext(ctx).Model(&task.Task{}).
		Where("priority = ?", priority).
		Where("status IN (?, ?)", task.TaskStatusOpen, task.TaskStatusInProgress)

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
		return 0, fmt.Errorf("failed to count priority tasks: %w", err)
	}
	return count, nil
}

// GetUnassignedTaskCount returns count of unassigned tasks
func (r *repository) GetUnassignedTaskCount(ctx context.Context, filter FilterParams) (int64, error) {
	query := r.db.WithContext(ctx).Model(&task.Task{}).
		Where("current_owner_id IS NULL").
		Where("status IN (?, ?)", task.TaskStatusOpen, task.TaskStatusInProgress)

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
		return 0, fmt.Errorf("failed to count unassigned tasks: %w", err)
	}
	return count, nil
}

// GetCompletedTasksTodayCount returns count of tasks completed today
func (r *repository) GetCompletedTasksTodayCount(ctx context.Context, filter FilterParams) (int64, error) {
	startOfDay := time.Now().UTC().Truncate(24 * time.Hour)

	query := r.db.WithContext(ctx).Model(&task.Task{}).
		Where("status = ?", task.TaskStatusCompleted).
		Where("completed_at >= ?", startOfDay)

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
		return 0, fmt.Errorf("failed to count completed tasks: %w", err)
	}
	return count, nil
}

// GetOpenTaskCount returns count of open tasks
func (r *repository) GetOpenTaskCount(ctx context.Context, filter FilterParams) (int64, error) {
	query := r.db.WithContext(ctx).Model(&task.Task{}).
		Where("status IN (?, ?)", task.TaskStatusOpen, task.TaskStatusInProgress)

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
		return 0, fmt.Errorf("failed to count open tasks: %w", err)
	}
	return count, nil
}

// GetPatientsInTriageOver2hrs returns count of patients in triage for over 2 hours
func (r *repository) GetPatientsInTriageOver2hrs(ctx context.Context, filter FilterParams) (int64, error) {
	twoHoursAgo := time.Now().UTC().Add(-2 * time.Hour)

	subquery := r.db.WithContext(ctx).
		Model(&flow.FlowStateTransition{}).
		Select("encounter_id, to_state, transitioned_at, ROW_NUMBER() OVER (PARTITION BY encounter_id ORDER BY transitioned_at DESC) as rn").
		Where("encounter_id IN (?)",
			r.db.Model(&encounter.Encounter{}).Select("id").Where("status = ?", encounter.EncounterStatusActive))

	query := r.db.WithContext(ctx).
		Table("(?) as latest_states", subquery).
		Where("rn = 1").
		Where("to_state = ?", flow.StateTriage).
		Where("transitioned_at < ?", twoHoursAgo)

	if filter.UnitID != nil || filter.DepartmentID != nil {
		encounterQuery := r.db.Model(&encounter.Encounter{}).Select("id")
		if filter.UnitID != nil {
			encounterQuery = encounterQuery.Where("unit_id = ?", *filter.UnitID)
		}
		if filter.DepartmentID != nil {
			encounterQuery = encounterQuery.Where("department_id = ?", *filter.DepartmentID)
		}
		query = query.Where("encounter_id IN (?)", encounterQuery)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count patients in triage: %w", err)
	}
	return count, nil
}

// GetEncountersWithoutCareTeam returns count of active encounters with no care team
func (r *repository) GetEncountersWithoutCareTeam(ctx context.Context, filter FilterParams) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("encounters").
		Where("status = ?", encounter.EncounterStatusActive).
		Where("NOT EXISTS (SELECT 1 FROM care_team_assignments WHERE care_team_assignments.encounter_id = encounters.id AND care_team_assignments.ends_at IS NULL)")

	if filter.UnitID != nil {
		query = query.Where("unit_id = ?", *filter.UnitID)
	}
	if filter.DepartmentID != nil {
		query = query.Where("department_id = ?", *filter.DepartmentID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count encounters without care team: %w", err)
	}
	return count, nil
}
