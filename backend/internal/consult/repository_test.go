package consult

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
		// Disable default transaction to improve test performance
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	
	// SQLite doesn't support uuid type with default function
	// We'll manually set IDs in tests
	err = db.Migrator().AutoMigrate(&ConsultRequest{})
	if err != nil {
		// Try with exec to handle SQLite compatibility
		sqlDB, _ := db.DB()
		_, err = sqlDB.Exec(`CREATE TABLE IF NOT EXISTS consult_requests (
			id TEXT PRIMARY KEY,
			encounter_id TEXT NOT NULL,
			target_service TEXT NOT NULL,
			reason TEXT NOT NULL,
			urgency TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			created_by TEXT NOT NULL,
			created_at DATETIME,
			accepted_by TEXT,
			accepted_at DATETIME,
			closed_at DATETIME,
			close_reason TEXT,
			redirected_to TEXT,
			updated_at DATETIME
		)`)
		require.NoError(t, err)
	}
	
	return &database.DB{DB: db}
}

func TestRepositoryCreate(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	consult := &ConsultRequest{
		ID:            "consult-1",
		EncounterID:   "encounter-1",
		TargetService: "cardiology",
		Reason:        "chest pain evaluation",
		Urgency:       ConsultUrgencyUrgent,
		Status:        ConsultStatusPending,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := repo.Create(ctx, consult)
	require.NoError(t, err)

	// Verify it was saved
	retrieved, err := repo.GetByID(ctx, "consult-1")
	require.NoError(t, err)
	assert.Equal(t, "consult-1", retrieved.ID)
	assert.Equal(t, "encounter-1", retrieved.EncounterID)
	assert.Equal(t, "cardiology", retrieved.TargetService)
	assert.Equal(t, ConsultUrgencyUrgent, retrieved.Urgency)
}

func TestRepositoryGetByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	consult := &ConsultRequest{
		ID:            "consult-2",
		EncounterID:   "encounter-2",
		TargetService: "neurology",
		Reason:        "headache assessment",
		Urgency:       ConsultUrgencyRoutine,
		Status:        ConsultStatusPending,
		CreatedBy:     "user-2",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, repo.Create(ctx, consult))

	t.Run("found", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, "consult-2")
		require.NoError(t, err)
		assert.Equal(t, "consult-2", retrieved.ID)
		assert.Equal(t, "neurology", retrieved.TargetService)
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

	// Create multiple consults
	consults := []*ConsultRequest{
		{
			ID:            "consult-list-1",
			EncounterID:   "encounter-1",
			TargetService: "cardiology",
			Reason:        "test",
			Urgency:       ConsultUrgencyUrgent,
			Status:        ConsultStatusPending,
			CreatedBy:     "user-1",
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            "consult-list-2",
			EncounterID:   "encounter-1",
			TargetService: "neurology",
			Reason:        "test",
			Urgency:       ConsultUrgencyRoutine,
			Status:        ConsultStatusAccepted,
			CreatedBy:     "user-1",
			CreatedAt:     now.Add(1 * time.Minute),
			UpdatedAt:     now.Add(1 * time.Minute),
		},
		{
			ID:            "consult-list-3",
			EncounterID:   "encounter-2",
			TargetService: "cardiology",
			Reason:        "test",
			Urgency:       ConsultUrgencyRoutine,
			Status:        ConsultStatusCompleted,
			CreatedBy:     "user-2",
			CreatedAt:     now.Add(2 * time.Minute),
			UpdatedAt:     now.Add(2 * time.Minute),
		},
	}

	for _, c := range consults {
		require.NoError(t, repo.Create(ctx, c))
	}

	t.Run("no filter - all results", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListConsultsFilter{Limit: -1})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
		// Should be ordered by created_at DESC
		assert.Equal(t, "consult-list-3", results[0].ID)
		assert.Equal(t, "consult-list-2", results[1].ID)
		assert.Equal(t, "consult-list-1", results[2].ID)
	})

	t.Run("filter by status", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListConsultsFilter{
			Status: ConsultStatusPending,
			Limit:  -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "consult-list-1", results[0].ID)
	})

	t.Run("filter by target service", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListConsultsFilter{
			TargetService: "cardiology",
			Limit:         -1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, results, 2)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListConsultsFilter{
			Limit:  2,
			Offset: 1,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 2)
		assert.Equal(t, "consult-list-2", results[0].ID)
		assert.Equal(t, "consult-list-1", results[1].ID)
	})
}

func TestRepositoryUpdate(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	consult := &ConsultRequest{
		ID:            "consult-update-1",
		EncounterID:   "encounter-1",
		TargetService: "cardiology",
		Reason:        "original reason",
		Urgency:       ConsultUrgencyRoutine,
		Status:        ConsultStatusPending,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, repo.Create(ctx, consult))

	// Update fields
	acceptedBy := "user-2"
	acceptedAt := now.Add(10 * time.Minute)
	consult.Status = ConsultStatusAccepted
	consult.AcceptedBy = &acceptedBy
	consult.AcceptedAt = &acceptedAt
	consult.UpdatedAt = now.Add(10 * time.Minute)

	err := repo.Update(ctx, consult)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, "consult-update-1")
	require.NoError(t, err)
	assert.Equal(t, ConsultStatusAccepted, updated.Status)
	assert.NotNil(t, updated.AcceptedBy)
	assert.Equal(t, "user-2", *updated.AcceptedBy)
	assert.NotNil(t, updated.AcceptedAt)
}
