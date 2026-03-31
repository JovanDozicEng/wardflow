package task

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/wardflow/backend/pkg/database"
)

func newRepositoryTestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	require.NoError(t, err)
	// Manually create tables for SQLite (UUID constraints don't work)
	err = db.Exec(`CREATE TABLE tasks (
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
	)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE TABLE task_assignment_events (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		from_owner_id TEXT,
		to_owner_id TEXT,
		assigned_at DATETIME NOT NULL,
		assigned_by TEXT NOT NULL,
		reason TEXT,
		created_at DATETIME
	)`).Error
	require.NoError(t, err)
	return &database.DB{DB: db}
}

func TestRepositoryCreate(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	ownerID := "user-1"
	slaAt := now.Add(24 * time.Hour)

	task := &Task{
		ID:             "task-1",
		ScopeType:      ScopeTypeEncounter,
		ScopeID:        "encounter-1",
		Title:          "Review lab results",
		Status:         TaskStatusOpen,
		Priority:       TaskPriorityMedium,
		SLADueAt:       &slaAt,
		CurrentOwnerID: &ownerID,
		CreatedBy:      "user-2",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(ctx, task)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, "task-1")
	require.NoError(t, err)
	assert.Equal(t, "task-1", retrieved.ID)
	assert.Equal(t, "Review lab results", retrieved.Title)
	assert.Equal(t, TaskStatusOpen, retrieved.Status)
}

func TestRepositoryGetByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	task := &Task{
		ID:        "task-by-id-1",
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Check medication",
		Status:    TaskStatusOpen,
		Priority:  TaskPriorityHigh,
		CreatedBy: "user-1",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, repo.Create(ctx, task))

	t.Run("found", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, "task-by-id-1")
		require.NoError(t, err)
		assert.Equal(t, "task-by-id-1", retrieved.ID)
		assert.Equal(t, "Check medication", retrieved.Title)
	})

	t.Run("not found", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, retrieved)
	})
}

func TestRepositoryList(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	ownerID := "user-1"
	slaOverdue := now.Add(-2 * time.Hour) // Clearly in the past
	slaNormal := now.Add(24 * time.Hour)

	tasks := []*Task{
		{
			ID:        "task-list-1",
			ScopeType: ScopeTypeEncounter,
			ScopeID:   "encounter-1",
			Title:     "Task 1",
			Status:    TaskStatusOpen,
			Priority:  TaskPriorityMedium,
			CreatedBy: "user-1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:             "task-list-2",
			ScopeType:      ScopeTypeEncounter,
			ScopeID:        "encounter-1",
			Title:          "Task 2",
			Status:         TaskStatusInProgress,
			Priority:       TaskPriorityHigh,
			CurrentOwnerID: &ownerID,
			SLADueAt:       &slaOverdue,
			CreatedBy:      "user-1",
			CreatedAt:      now.Add(1 * time.Hour),
			UpdatedAt:      now.Add(1 * time.Hour),
		},
		{
			ID:        "task-list-3",
			ScopeType: ScopeTypePatient,
			ScopeID:   "patient-1",
			Title:     "Task 3",
			Status:    TaskStatusCompleted,
			Priority:  TaskPriorityLow,
			SLADueAt:  &slaNormal,
			CreatedBy: "user-2",
			CreatedAt: now.Add(2 * time.Hour),
			UpdatedAt: now.Add(2 * time.Hour),
		},
	}

	for _, task := range tasks {
		require.NoError(t, repo.Create(ctx, task))
	}

	t.Run("no filter - all results", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListTasksFilter{Limit: -1})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
		// Ordered by created_at DESC
		assert.Equal(t, "task-list-3", results[0].ID)
	})

	t.Run("filter by scope type", func(t *testing.T) {
		scopeType := ScopeTypeEncounter
		results, total, err := repo.List(ctx, ListTasksFilter{
			ScopeType: &scopeType,
			Limit:     -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
	})

	t.Run("filter by status", func(t *testing.T) {
		status := TaskStatusOpen
		results, total, err := repo.List(ctx, ListTasksFilter{
			Status: &status,
			Limit:  -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "task-list-1", results[0].ID)
	})

	t.Run("filter by priority", func(t *testing.T) {
		priority := TaskPriorityHigh
		results, total, err := repo.List(ctx, ListTasksFilter{
			Priority: &priority,
			Limit:    -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "task-list-2", results[0].ID)
	})

	t.Run("filter by owner", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListTasksFilter{
			OwnerID: &ownerID,
			Limit:   -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "task-list-2", results[0].ID)
	})

	t.Run("filter by overdue", func(t *testing.T) {
		overdue := true
		results, total, err := repo.List(ctx, ListTasksFilter{
			Overdue: &overdue,
			Limit:   -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "task-list-2", results[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListTasksFilter{
			Limit:  1,
			Offset: 1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "task-list-2", results[0].ID)
	})
}

func TestRepositoryUpdate(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	task := &Task{
		ID:        "task-update-1",
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Original title",
		Status:    TaskStatusOpen,
		Priority:  TaskPriorityMedium,
		CreatedBy: "user-1",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, repo.Create(ctx, task))

	// Update task
	task.Title = "Updated title"
	task.Status = TaskStatusInProgress
	task.UpdatedAt = now.Add(10 * time.Minute)

	err := repo.Update(ctx, task)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, "task-update-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated title", updated.Title)
	assert.Equal(t, TaskStatusInProgress, updated.Status)
}

func TestRepositoryCreateAssignmentEvent(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create parent task
	task := &Task{
		ID:        "task-event-1",
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Task with events",
		Status:    TaskStatusOpen,
		Priority:  TaskPriorityMedium,
		CreatedBy: "user-1",
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, task))

	// Create assignment event
	toOwnerID := "user-2"
	reason := "Reassigning to specialist"
	event := &TaskAssignmentEvent{
		ID:         "event-1",
		TaskID:     "task-event-1",
		ToOwnerID:  &toOwnerID,
		AssignedAt: now.Add(1 * time.Hour),
		AssignedBy: "user-3",
		Reason:     &reason,
		CreatedAt:  now.Add(1 * time.Hour),
	}

	err := repo.CreateAssignmentEvent(ctx, event)
	require.NoError(t, err)

	// Verify by getting history
	history, err := repo.GetAssignmentHistory(ctx, "task-event-1")
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "event-1", history[0].ID)
	assert.NotNil(t, history[0].ToOwnerID)
	assert.Equal(t, "user-2", *history[0].ToOwnerID)
	assert.NotNil(t, history[0].Reason)
	assert.Equal(t, "Reassigning to specialist", *history[0].Reason)
}

func TestRepositoryGetAssignmentHistory(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create task
	task := &Task{
		ID:        "task-history-1",
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Task with history",
		Status:    TaskStatusOpen,
		Priority:  TaskPriorityMedium,
		CreatedBy: "user-1",
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, task))

	// Create multiple assignment events
	toOwner1 := "user-2"
	toOwner2 := "user-3"
	events := []*TaskAssignmentEvent{
		{
			ID:         "event-h-1",
			TaskID:     "task-history-1",
			ToOwnerID:  &toOwner1,
			AssignedAt: now.Add(1 * time.Hour),
			AssignedBy: "user-1",
			CreatedAt:  now.Add(1 * time.Hour),
		},
		{
			ID:         "event-h-2",
			TaskID:     "task-history-1",
			FromOwnerID: &toOwner1,
			ToOwnerID:  &toOwner2,
			AssignedAt: now.Add(2 * time.Hour),
			AssignedBy: "user-2",
			CreatedAt:  now.Add(2 * time.Hour),
		},
	}

	for _, e := range events {
		require.NoError(t, repo.CreateAssignmentEvent(ctx, e))
	}

	// Get history
	history, err := repo.GetAssignmentHistory(ctx, "task-history-1")
	require.NoError(t, err)
	assert.Len(t, history, 2)
	// Should be ordered by assigned_at DESC
	assert.Equal(t, "event-h-2", history[0].ID)
	assert.Equal(t, "event-h-1", history[1].ID)

	t.Run("no history for non-existent task", func(t *testing.T) {
		history, err := repo.GetAssignmentHistory(ctx, "non-existent")
		require.NoError(t, err)
		assert.Empty(t, history)
	})
}

func TestRepositoryGetAssignmentHistoryPaginated(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create task
	task := &Task{
		ID:        "task-page-history-1",
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Task with paginated history",
		Status:    TaskStatusOpen,
		Priority:  TaskPriorityMedium,
		CreatedBy: "user-1",
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, task))

	// Create multiple assignment events
	toOwner1 := "user-2"
	toOwner2 := "user-3"
	toOwner3 := "user-4"
	events := []*TaskAssignmentEvent{
		{
			ID:         "event-page-1",
			TaskID:     "task-page-history-1",
			ToOwnerID:  &toOwner1,
			AssignedAt: now.Add(1 * time.Hour),
			AssignedBy: "user-1",
			CreatedAt:  now.Add(1 * time.Hour),
		},
		{
			ID:         "event-page-2",
			TaskID:     "task-page-history-1",
			FromOwnerID: &toOwner1,
			ToOwnerID:  &toOwner2,
			AssignedAt: now.Add(2 * time.Hour),
			AssignedBy: "user-2",
			CreatedAt:  now.Add(2 * time.Hour),
		},
		{
			ID:         "event-page-3",
			TaskID:     "task-page-history-1",
			FromOwnerID: &toOwner2,
			ToOwnerID:  &toOwner3,
			AssignedAt: now.Add(3 * time.Hour),
			AssignedBy: "user-3",
			CreatedAt:  now.Add(3 * time.Hour),
		},
	}

	for _, e := range events {
		require.NoError(t, repo.CreateAssignmentEvent(ctx, e))
	}

	t.Run("all results", func(t *testing.T) {
		results, total, err := repo.GetAssignmentHistoryPaginated(ctx, "task-page-history-1", -1, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
		// Ordered by assigned_at DESC
		assert.Equal(t, "event-page-3", results[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.GetAssignmentHistoryPaginated(ctx, "task-page-history-1", 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "event-page-2", results[0].ID)
	})
}

func TestRepositoryDelete(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	task := &Task{
		ID:        "task-delete-1",
		ScopeType: ScopeTypeEncounter,
		ScopeID:   "encounter-1",
		Title:     "Task to delete",
		Status:    TaskStatusOpen,
		Priority:  TaskPriorityMedium,
		CreatedBy: "user-1",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, repo.Create(ctx, task))

	t.Run("delete existing task", func(t *testing.T) {
		err := repo.Delete(ctx, "task-delete-1")
		require.NoError(t, err)

		// Verify it's deleted
		retrieved, err := repo.GetByID(ctx, "task-delete-1")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existent task", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
