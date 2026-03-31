package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wardflow/backend/internal/careteam"
	"github.com/wardflow/backend/internal/encounter"
	"github.com/wardflow/backend/internal/task"
	"github.com/wardflow/backend/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDashboardTestDB(t *testing.T) *database.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	assert.NoError(t, err)

	// Manually create simplified tables for SQLite (no UUID type)
	// These are minimal schemas for testing purposes
	db.Exec(`CREATE TABLE IF NOT EXISTS encounters (
		id TEXT PRIMARY KEY,
		patient_id TEXT NOT NULL,
		unit_id TEXT NOT NULL,
		department_id TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		started_at DATETIME NOT NULL,
		ended_at DATETIME,
		created_by TEXT NOT NULL,
		updated_by TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		scope_type TEXT NOT NULL,
		scope_id TEXT NOT NULL,
		title TEXT NOT NULL,
		details TEXT,
		status TEXT NOT NULL DEFAULT 'open',
		priority TEXT NOT NULL DEFAULT 'medium',
		sla_due_at DATETIME,
		current_owner_id TEXT,
		created_by TEXT NOT NULL,
		created_at DATETIME,
		completed_by TEXT,
		completed_at DATETIME,
		updated_at DATETIME
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS flow_state_transitions (
		id TEXT PRIMARY KEY,
		encounter_id TEXT NOT NULL,
		from_state TEXT,
		to_state TEXT NOT NULL,
		transitioned_at DATETIME NOT NULL,
		actor_type TEXT NOT NULL,
		actor_user_id TEXT,
		reason TEXT,
		source_event_id TEXT,
		is_override INTEGER DEFAULT 0,
		created_at DATETIME
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS care_team_assignments (
		id TEXT PRIMARY KEY,
		encounter_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		role_type TEXT NOT NULL,
		starts_at DATETIME NOT NULL,
		ends_at DATETIME,
		created_by TEXT NOT NULL,
		created_at DATETIME,
		handoff_note_id TEXT
	)`)

	return &database.DB{DB: db}
}

func TestRepository_GetActiveEncounterCount_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetActiveEncounterCount(ctx, FilterParams{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepository_GetActiveEncounterCount_WithData(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Insert test encounters
	now := time.Now().UTC()
	encounters := []encounter.Encounter{
		{ID: "enc-1", PatientID: "pat-1", UnitID: "unit-1", DepartmentID: "dept-1", Status: encounter.EncounterStatusActive, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1"},
		{ID: "enc-2", PatientID: "pat-2", UnitID: "unit-2", DepartmentID: "dept-1", Status: encounter.EncounterStatusActive, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1"},
		{ID: "enc-3", PatientID: "pat-3", UnitID: "unit-1", DepartmentID: "dept-2", Status: encounter.EncounterStatusDischarged, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1"},
	}

	for _, enc := range encounters {
		err := db.Create(&enc).Error
		assert.NoError(t, err)
	}

	// Test: all active encounters
	count, err := repo.GetActiveEncounterCount(ctx, FilterParams{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Test: filter by unit
	unit1 := "unit-1"
	count, err = repo.GetActiveEncounterCount(ctx, FilterParams{UnitID: &unit1})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Test: filter by department
	dept1 := "dept-1"
	count, err = repo.GetActiveEncounterCount(ctx, FilterParams{DepartmentID: &dept1})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRepository_GetFlowStateDistribution_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	dist, err := repo.GetFlowStateDistribution(ctx, FilterParams{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), dist.Arrived)
	assert.Equal(t, int64(0), dist.Triage)
}

func TestRepository_GetOverdueTaskCount_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetOverdueTaskCount(ctx, FilterParams{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepository_GetOverdueTaskCount_WithData(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Create encounter first
	now := time.Now().UTC()
	enc := encounter.Encounter{
		ID: "enc-1", PatientID: "pat-1", UnitID: "unit-1", DepartmentID: "dept-1",
		Status: encounter.EncounterStatusActive, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1",
	}
	err := db.Create(&enc).Error
	assert.NoError(t, err)

	// Insert overdue and non-overdue tasks
	pastDue := now.Add(-2 * time.Hour)
	futureDue := now.Add(2 * time.Hour)

	tasks := []task.Task{
		{ID: "task-1", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "Overdue", Status: task.TaskStatusOpen, Priority: task.TaskPriorityMedium, SLADueAt: &pastDue, CreatedBy: "user-1"},
		{ID: "task-2", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "Not overdue", Status: task.TaskStatusOpen, Priority: task.TaskPriorityMedium, SLADueAt: &futureDue, CreatedBy: "user-1"},
		{ID: "task-3", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "Overdue completed", Status: task.TaskStatusCompleted, Priority: task.TaskPriorityMedium, SLADueAt: &pastDue, CreatedBy: "user-1"},
		{ID: "task-4", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "No SLA", Status: task.TaskStatusOpen, Priority: task.TaskPriorityMedium, CreatedBy: "user-1"},
	}

	for _, tsk := range tasks {
		err := db.Create(&tsk).Error
		assert.NoError(t, err)
	}

	count, err := repo.GetOverdueTaskCount(ctx, FilterParams{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count) // Only task-1 is overdue and not completed
}

func TestRepository_GetHighPriorityTaskCount_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetHighPriorityTaskCount(ctx, FilterParams{}, task.TaskPriorityHigh)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepository_GetHighPriorityTaskCount_WithData(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Create encounter
	now := time.Now().UTC()
	enc := encounter.Encounter{
		ID: "enc-1", PatientID: "pat-1", UnitID: "unit-1", DepartmentID: "dept-1",
		Status: encounter.EncounterStatusActive, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1",
	}
	err := db.Create(&enc).Error
	assert.NoError(t, err)

	tasks := []task.Task{
		{ID: "task-1", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "High priority open", Status: task.TaskStatusOpen, Priority: task.TaskPriorityHigh, CreatedBy: "user-1"},
		{ID: "task-2", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "High priority in progress", Status: task.TaskStatusInProgress, Priority: task.TaskPriorityHigh, CreatedBy: "user-1"},
		{ID: "task-3", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "High priority completed", Status: task.TaskStatusCompleted, Priority: task.TaskPriorityHigh, CreatedBy: "user-1"},
		{ID: "task-4", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "Medium priority", Status: task.TaskStatusOpen, Priority: task.TaskPriorityMedium, CreatedBy: "user-1"},
	}

	for _, tsk := range tasks {
		err := db.Create(&tsk).Error
		assert.NoError(t, err)
	}

	count, err := repo.GetHighPriorityTaskCount(ctx, FilterParams{}, task.TaskPriorityHigh)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count) // task-1 and task-2
}

func TestRepository_GetUnassignedTaskCount_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetUnassignedTaskCount(ctx, FilterParams{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepository_GetUnassignedTaskCount_WithData(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Create encounter
	now := time.Now().UTC()
	enc := encounter.Encounter{
		ID: "enc-1", PatientID: "pat-1", UnitID: "unit-1", DepartmentID: "dept-1",
		Status: encounter.EncounterStatusActive, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1",
	}
	err := db.Create(&enc).Error
	assert.NoError(t, err)

	ownerID := "owner-1"
	tasks := []task.Task{
		{ID: "task-1", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "Unassigned open", Status: task.TaskStatusOpen, Priority: task.TaskPriorityMedium, CreatedBy: "user-1"},
		{ID: "task-2", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "Assigned open", Status: task.TaskStatusOpen, Priority: task.TaskPriorityMedium, CurrentOwnerID: &ownerID, CreatedBy: "user-1"},
		{ID: "task-3", ScopeType: task.ScopeTypeEncounter, ScopeID: "enc-1", Title: "Unassigned completed", Status: task.TaskStatusCompleted, Priority: task.TaskPriorityMedium, CreatedBy: "user-1"},
	}

	for _, tsk := range tasks {
		err := db.Create(&tsk).Error
		assert.NoError(t, err)
	}

	count, err := repo.GetUnassignedTaskCount(ctx, FilterParams{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count) // Only task-1
}

func TestRepository_GetCompletedTasksTodayCount_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetCompletedTasksTodayCount(ctx, FilterParams{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepository_GetOpenTaskCount_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetOpenTaskCount(ctx, FilterParams{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepository_GetPatientsInTriageOver2hrs_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetPatientsInTriageOver2hrs(ctx, FilterParams{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepository_GetEncountersWithoutCareTeam_EmptyDB(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	count, err := repo.GetEncountersWithoutCareTeam(ctx, FilterParams{})

	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRepository_GetEncountersWithoutCareTeam_WithData(t *testing.T) {
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Create active encounters
	now := time.Now().UTC()
	encounters := []encounter.Encounter{
		{ID: "enc-1", PatientID: "pat-1", UnitID: "unit-1", DepartmentID: "dept-1", Status: encounter.EncounterStatusActive, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1"},
		{ID: "enc-2", PatientID: "pat-2", UnitID: "unit-1", DepartmentID: "dept-1", Status: encounter.EncounterStatusActive, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1"},
		{ID: "enc-3", PatientID: "pat-3", UnitID: "unit-1", DepartmentID: "dept-1", Status: encounter.EncounterStatusDischarged, StartedAt: now, CreatedBy: "user-1", UpdatedBy: "user-1"},
	}

	for _, enc := range encounters {
		err := db.Create(&enc).Error
		assert.NoError(t, err)
	}

	// Assign care team to enc-2
	assignment := careteam.CareTeamAssignment{
		ID:          "assign-1",
		EncounterID: "enc-2",
		UserID:      "user-1",
		RoleType:    careteam.RolePrimaryNurse,
		StartsAt:    now,
		CreatedBy:   "user-1",
	}
	err := db.Create(&assignment).Error
	assert.NoError(t, err)

	count, err := repo.GetEncountersWithoutCareTeam(ctx, FilterParams{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count) // Only enc-1 is active without care team
}

func TestRepository_AllMethods_RunWithoutError(t *testing.T) {
	// Integration test: all methods should run without error on empty DB
	db := newDashboardTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	filter := FilterParams{}

	t.Run("GetActiveEncounterCount", func(t *testing.T) {
		_, err := repo.GetActiveEncounterCount(ctx, filter)
		assert.NoError(t, err)
	})

	t.Run("GetFlowStateDistribution", func(t *testing.T) {
		_, err := repo.GetFlowStateDistribution(ctx, filter)
		assert.NoError(t, err)
	})

	t.Run("GetOverdueTaskCount", func(t *testing.T) {
		_, err := repo.GetOverdueTaskCount(ctx, filter)
		assert.NoError(t, err)
	})

	t.Run("GetHighPriorityTaskCount", func(t *testing.T) {
		_, err := repo.GetHighPriorityTaskCount(ctx, filter, task.TaskPriorityHigh)
		assert.NoError(t, err)
	})

	t.Run("GetUnassignedTaskCount", func(t *testing.T) {
		_, err := repo.GetUnassignedTaskCount(ctx, filter)
		assert.NoError(t, err)
	})

	t.Run("GetCompletedTasksTodayCount", func(t *testing.T) {
		_, err := repo.GetCompletedTasksTodayCount(ctx, filter)
		assert.NoError(t, err)
	})

	t.Run("GetOpenTaskCount", func(t *testing.T) {
		_, err := repo.GetOpenTaskCount(ctx, filter)
		assert.NoError(t, err)
	})

	t.Run("GetPatientsInTriageOver2hrs", func(t *testing.T) {
		_, err := repo.GetPatientsInTriageOver2hrs(ctx, filter)
		assert.NoError(t, err)
	})

	t.Run("GetEncountersWithoutCareTeam", func(t *testing.T) {
		_, err := repo.GetEncountersWithoutCareTeam(ctx, filter)
		assert.NoError(t, err)
	})
}
