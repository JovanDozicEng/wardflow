package exception

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
	err = db.Migrator().AutoMigrate(&ExceptionEvent{})
	if err != nil {
		// Fallback to manual table creation for SQLite
		sqlDB, _ := db.DB()
		_, _ = sqlDB.Exec(`CREATE TABLE IF NOT EXISTS exception_events (
			id TEXT PRIMARY KEY,
			encounter_id TEXT NOT NULL,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'draft',
			required_fields TEXT NOT NULL,
			data TEXT NOT NULL,
			initiated_by TEXT NOT NULL,
			initiated_at DATETIME,
			finalized_by TEXT,
			finalized_at DATETIME,
			corrected_by_event_id TEXT,
			correction_reason TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`)
	}
	
	return &database.DB{DB: db}
}

func TestRepositoryCreate(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	exception := &ExceptionEvent{
		ID:             "exception-1",
		EncounterID:    "encounter-1",
		Type:           "late_admission",
		Status:         ExceptionStatusDraft,
		RequiredFields: `["reason", "approval"]`,
		Data:           `{"reason": "bed shortage"}`,
		InitiatedBy:    "user-1",
		InitiatedAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := repo.Create(ctx, exception)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetByID(ctx, "exception-1")
	require.NoError(t, err)
	assert.Equal(t, "exception-1", retrieved.ID)
	assert.Equal(t, "late_admission", retrieved.Type)
	assert.Equal(t, ExceptionStatusDraft, retrieved.Status)
}

func TestRepositoryGetByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	exception := &ExceptionEvent{
		ID:             "exception-2",
		EncounterID:    "encounter-2",
		Type:           "early_discharge",
		Status:         ExceptionStatusDraft,
		RequiredFields: `["reason"]`,
		Data:           `{"reason": "patient request"}`,
		InitiatedBy:    "user-2",
		InitiatedAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, repo.Create(ctx, exception))

	t.Run("found", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, "exception-2")
		require.NoError(t, err)
		assert.Equal(t, "exception-2", retrieved.ID)
		assert.Equal(t, "early_discharge", retrieved.Type)
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

	exceptions := []*ExceptionEvent{
		{
			ID:             "exception-list-1",
			EncounterID:    "encounter-1",
			Type:           "late_admission",
			Status:         ExceptionStatusDraft,
			RequiredFields: `[]`,
			Data:           `{}`,
			InitiatedBy:    "user-1",
			InitiatedAt:    now,
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             "exception-list-2",
			EncounterID:    "encounter-1",
			Type:           "early_discharge",
			Status:         ExceptionStatusFinalized,
			RequiredFields: `[]`,
			Data:           `{}`,
			InitiatedBy:    "user-2",
			InitiatedAt:    now.Add(1 * time.Hour),
			CreatedAt:      now.Add(1 * time.Hour),
			UpdatedAt:      now.Add(1 * time.Hour),
		},
		{
			ID:             "exception-list-3",
			EncounterID:    "encounter-2",
			Type:           "late_admission",
			Status:         ExceptionStatusCorrected,
			RequiredFields: `[]`,
			Data:           `{}`,
			InitiatedBy:    "user-3",
			InitiatedAt:    now.Add(2 * time.Hour),
			CreatedAt:      now.Add(2 * time.Hour),
			UpdatedAt:      now.Add(2 * time.Hour),
		},
	}

	for _, e := range exceptions {
		require.NoError(t, repo.Create(ctx, e))
	}

	t.Run("no filter - all results", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListExceptionsFilter{Limit: -1})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
		// Ordered by created_at DESC
		assert.Equal(t, "exception-list-3", results[0].ID)
	})

	t.Run("filter by encounter", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListExceptionsFilter{
			EncounterID: "encounter-1",
			Limit:       -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
	})

	t.Run("filter by type", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListExceptionsFilter{
			Type:  "late_admission",
			Limit: -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
	})

	t.Run("filter by status", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListExceptionsFilter{
			Status: ExceptionStatusFinalized,
			Limit:  -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "exception-list-2", results[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListExceptionsFilter{
			Limit:  2,
			Offset: 1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 2)
		assert.Equal(t, "exception-list-2", results[0].ID)
		assert.Equal(t, "exception-list-1", results[1].ID)
	})
}

func TestRepositoryUpdate(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	exception := &ExceptionEvent{
		ID:             "exception-update-1",
		EncounterID:    "encounter-1",
		Type:           "late_admission",
		Status:         ExceptionStatusDraft,
		RequiredFields: `["reason"]`,
		Data:           `{"reason": "initial"}`,
		InitiatedBy:    "user-1",
		InitiatedAt:    now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, repo.Create(ctx, exception))

	// Update fields
	finalizedBy := "user-2"
	finalizedAt := now.Add(10 * time.Minute)
	exception.Status = ExceptionStatusFinalized
	exception.FinalizedBy = &finalizedBy
	exception.FinalizedAt = &finalizedAt
	exception.Data = `{"reason": "updated reason", "approval": "approved"}`
	exception.UpdatedAt = now.Add(10 * time.Minute)

	err := repo.Update(ctx, exception)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, "exception-update-1")
	require.NoError(t, err)
	assert.Equal(t, ExceptionStatusFinalized, updated.Status)
	assert.NotNil(t, updated.FinalizedBy)
	assert.Equal(t, "user-2", *updated.FinalizedBy)
	assert.Contains(t, updated.Data, "updated reason")
}
