package incident

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
	
	// SQLite doesn't support uuid type with default function
	err = db.Migrator().AutoMigrate(&Incident{}, &IncidentStatusEvent{})
	if err != nil {
		// Fallback to manual table creation for SQLite
		sqlDB, _ := db.DB()
		_, _ = sqlDB.Exec(`CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,
			encounter_id TEXT,
			type TEXT NOT NULL,
			severity TEXT,
			harm_indicators TEXT,
			event_time DATETIME NOT NULL,
			reported_by TEXT NOT NULL,
			reported_at DATETIME,
			status TEXT NOT NULL DEFAULT 'submitted',
			created_at DATETIME,
			updated_at DATETIME
		)`)
		_, _ = sqlDB.Exec(`CREATE TABLE IF NOT EXISTS incident_status_events (
			id TEXT PRIMARY KEY,
			incident_id TEXT NOT NULL,
			from_status TEXT,
			to_status TEXT NOT NULL,
			changed_by TEXT NOT NULL,
			changed_at DATETIME,
			note TEXT
		)`)
	}
	
	return &database.DB{DB: db}
}

func TestRepositoryCreate(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	encounterID := "encounter-1"
	severity := "high"
	now := time.Now()

	incident := &Incident{
		ID:          "incident-1",
		EncounterID: &encounterID,
		Type:        "fall",
		Severity:    &severity,
		EventTime:   now,
		ReportedBy:  "user-1",
		ReportedAt:  now,
		Status:      IncidentStatusSubmitted,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := repo.Create(ctx, incident)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, "incident-1")
	require.NoError(t, err)
	assert.Equal(t, "incident-1", retrieved.ID)
	assert.Equal(t, "fall", retrieved.Type)
	assert.Equal(t, IncidentStatusSubmitted, retrieved.Status)
}

func TestRepositoryGetByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	incident := &Incident{
		ID:         "incident-2",
		Type:       "medication-error",
		EventTime:  now,
		ReportedBy: "user-2",
		ReportedAt: now,
		Status:     IncidentStatusSubmitted,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	require.NoError(t, repo.Create(ctx, incident))

	t.Run("found", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, "incident-2")
		require.NoError(t, err)
		assert.Equal(t, "incident-2", retrieved.ID)
		assert.Equal(t, "medication-error", retrieved.Type)
	})

	t.Run("not found", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, "non-existent")
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, retrieved)
	})
}

func TestRepositoryList(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	incidents := []*Incident{
		{
			ID:         "incident-list-1",
			Type:       "fall",
			EventTime:  now,
			ReportedBy: "user-1",
			ReportedAt: now,
			Status:     IncidentStatusSubmitted,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "incident-list-2",
			Type:       "pressure-injury",
			EventTime:  now.Add(1 * time.Hour),
			ReportedBy: "user-2",
			ReportedAt: now.Add(1 * time.Hour),
			Status:     IncidentStatusUnderReview,
			CreatedAt:  now.Add(1 * time.Hour),
			UpdatedAt:  now.Add(1 * time.Hour),
		},
		{
			ID:         "incident-list-3",
			Type:       "fall",
			EventTime:  now.Add(2 * time.Hour),
			ReportedBy: "user-3",
			ReportedAt: now.Add(2 * time.Hour),
			Status:     IncidentStatusClosed,
			CreatedAt:  now.Add(2 * time.Hour),
			UpdatedAt:  now.Add(2 * time.Hour),
		},
	}

	for _, i := range incidents {
		require.NoError(t, repo.Create(ctx, i))
	}

	t.Run("no filter - all results", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListIncidentsFilter{Limit: -1})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
		// Ordered by event_time DESC
		assert.Equal(t, "incident-list-3", results[0].ID)
	})

	t.Run("filter by status", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListIncidentsFilter{
			Status: IncidentStatusSubmitted,
			Limit:  -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "incident-list-1", results[0].ID)
	})

	t.Run("filter by type", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListIncidentsFilter{
			Type:  "fall",
			Limit: -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListIncidentsFilter{
			Limit:  1,
			Offset: 1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "incident-list-2", results[0].ID)
	})
}

func TestRepositoryUpdate(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	incident := &Incident{
		ID:         "incident-update-1",
		Type:       "fall",
		EventTime:  now,
		ReportedBy: "user-1",
		ReportedAt: now,
		Status:     IncidentStatusSubmitted,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	require.NoError(t, repo.Create(ctx, incident))

	// Update status
	incident.Status = IncidentStatusUnderReview
	incident.UpdatedAt = now.Add(10 * time.Minute)

	err := repo.Update(ctx, incident)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, "incident-update-1")
	require.NoError(t, err)
	assert.Equal(t, IncidentStatusUnderReview, updated.Status)
}

func TestRepositoryCreateStatusEvent(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create parent incident
	incident := &Incident{
		ID:         "incident-event-1",
		Type:       "fall",
		EventTime:  now,
		ReportedBy: "user-1",
		ReportedAt: now,
		Status:     IncidentStatusSubmitted,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, repo.Create(ctx, incident))

	// Create status event
	fromStatus := IncidentStatusSubmitted
	note := "Starting investigation"
	event := &IncidentStatusEvent{
		ID:         "event-1",
		IncidentID: "incident-event-1",
		FromStatus: &fromStatus,
		ToStatus:   IncidentStatusUnderReview,
		ChangedBy:  "user-2",
		ChangedAt:  now.Add(1 * time.Hour),
		Note:       &note,
	}

	err := repo.CreateStatusEvent(ctx, event)
	require.NoError(t, err)

	// Verify by getting history
	history, err := repo.GetStatusHistory(ctx, "incident-event-1")
	require.NoError(t, err)
	assert.Len(t, history, 1)
	assert.Equal(t, "event-1", history[0].ID)
	assert.Equal(t, IncidentStatusUnderReview, history[0].ToStatus)
	assert.NotNil(t, history[0].Note)
	assert.Equal(t, "Starting investigation", *history[0].Note)
}

func TestRepositoryGetStatusHistory(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create incident
	incident := &Incident{
		ID:         "incident-history-1",
		Type:       "medication-error",
		EventTime:  now,
		ReportedBy: "user-1",
		ReportedAt: now,
		Status:     IncidentStatusSubmitted,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	require.NoError(t, repo.Create(ctx, incident))

	// Create multiple status events
	events := []*IncidentStatusEvent{
		{
			ID:         "event-h-1",
			IncidentID: "incident-history-1",
			ToStatus:   IncidentStatusUnderReview,
			ChangedBy:  "user-2",
			ChangedAt:  now.Add(1 * time.Hour),
		},
		{
			ID:         "event-h-2",
			IncidentID: "incident-history-1",
			ToStatus:   IncidentStatusClosed,
			ChangedBy:  "user-3",
			ChangedAt:  now.Add(2 * time.Hour),
		},
	}

	for _, e := range events {
		require.NoError(t, repo.CreateStatusEvent(ctx, e))
	}

	// Get history
	history, err := repo.GetStatusHistory(ctx, "incident-history-1")
	require.NoError(t, err)
	assert.Len(t, history, 2)
	// Should be ordered by changed_at ASC
	assert.Equal(t, "event-h-1", history[0].ID)
	assert.Equal(t, "event-h-2", history[1].ID)

	t.Run("no history for non-existent incident", func(t *testing.T) {
		history, err := repo.GetStatusHistory(ctx, "non-existent")
		require.NoError(t, err)
		assert.Empty(t, history)
	})
}
