package flow

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
	// Manually create table for SQLite (UUID constraints don't work)
	err = db.Exec(`CREATE TABLE flow_state_transitions (
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
	)`).Error
	require.NoError(t, err)
	return &database.DB{DB: db}
}

func TestRepositoryCreateTransition(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	fromState := StateArrived
	actorUserID := "user-1"

	transition := &FlowStateTransition{
		ID:            "transition-1",
		EncounterID:   "encounter-1",
		FromState:     &fromState,
		ToState:       StateTriage,
		TransitionedAt: now,
		ActorType:     ActorTypeUser,
		ActorUserID:   &actorUserID,
		CreatedAt:     now,
	}

	err := repo.CreateTransition(ctx, transition)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetTransitionByID(ctx, "transition-1")
	require.NoError(t, err)
	assert.Equal(t, "transition-1", retrieved.ID)
	assert.Equal(t, StateTriage, retrieved.ToState)
	assert.NotNil(t, retrieved.FromState)
	assert.Equal(t, StateArrived, *retrieved.FromState)
}

func TestRepositoryGetTimeline(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	transitions := []*FlowStateTransition{
		{
			ID:            "trans-1",
			EncounterID:   "encounter-timeline-1",
			ToState:       StateArrived,
			TransitionedAt: now,
			ActorType:     ActorTypeSystem,
			CreatedAt:     now,
		},
		{
			ID:            "trans-2",
			EncounterID:   "encounter-timeline-1",
			ToState:       StateTriage,
			TransitionedAt: now.Add(1 * time.Hour),
			ActorType:     ActorTypeUser,
			CreatedAt:     now.Add(1 * time.Hour),
		},
		{
			ID:            "trans-3",
			EncounterID:   "encounter-timeline-1",
			ToState:       StateProviderEval,
			TransitionedAt: now.Add(2 * time.Hour),
			ActorType:     ActorTypeUser,
			CreatedAt:     now.Add(2 * time.Hour),
		},
	}

	for _, tr := range transitions {
		require.NoError(t, repo.CreateTransition(ctx, tr))
	}

	// Get timeline
	timeline, err := repo.GetTimeline(ctx, "encounter-timeline-1")
	require.NoError(t, err)
	assert.Len(t, timeline, 3)
	// Should be ordered by transitioned_at ASC
	assert.Equal(t, "trans-1", timeline[0].ID)
	assert.Equal(t, "trans-2", timeline[1].ID)
	assert.Equal(t, "trans-3", timeline[2].ID)

	t.Run("empty timeline for non-existent encounter", func(t *testing.T) {
		timeline, err := repo.GetTimeline(ctx, "non-existent")
		require.NoError(t, err)
		assert.Empty(t, timeline)
	})
}

func TestRepositoryGetTimelinePaginated(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	transitions := []*FlowStateTransition{
		{
			ID:            "trans-page-1",
			EncounterID:   "encounter-page-1",
			ToState:       StateArrived,
			TransitionedAt: now,
			ActorType:     ActorTypeSystem,
			CreatedAt:     now,
		},
		{
			ID:            "trans-page-2",
			EncounterID:   "encounter-page-1",
			ToState:       StateTriage,
			TransitionedAt: now.Add(1 * time.Hour),
			ActorType:     ActorTypeUser,
			CreatedAt:     now.Add(1 * time.Hour),
		},
		{
			ID:            "trans-page-3",
			EncounterID:   "encounter-page-1",
			ToState:       StateProviderEval,
			TransitionedAt: now.Add(2 * time.Hour),
			ActorType:     ActorTypeUser,
			CreatedAt:     now.Add(2 * time.Hour),
		},
	}

	for _, tr := range transitions {
		require.NoError(t, repo.CreateTransition(ctx, tr))
	}

	t.Run("all results", func(t *testing.T) {
		results, total, err := repo.GetTimelinePaginated(ctx, "encounter-page-1", -1, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
		// Ordered by transitioned_at DESC
		assert.Equal(t, "trans-page-3", results[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.GetTimelinePaginated(ctx, "encounter-page-1", 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "trans-page-2", results[0].ID)
	})
}

func TestRepositoryGetCurrentState(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	transitions := []*FlowStateTransition{
		{
			ID:            "trans-current-1",
			EncounterID:   "encounter-current-1",
			ToState:       StateArrived,
			TransitionedAt: now,
			ActorType:     ActorTypeSystem,
			CreatedAt:     now,
		},
		{
			ID:            "trans-current-2",
			EncounterID:   "encounter-current-1",
			ToState:       StateTriage,
			TransitionedAt: now.Add(1 * time.Hour),
			ActorType:     ActorTypeUser,
			CreatedAt:     now.Add(1 * time.Hour),
		},
	}

	for _, tr := range transitions {
		require.NoError(t, repo.CreateTransition(ctx, tr))
	}

	t.Run("returns most recent state", func(t *testing.T) {
		current, err := repo.GetCurrentState(ctx, "encounter-current-1")
		require.NoError(t, err)
		assert.NotNil(t, current)
		assert.Equal(t, "trans-current-2", current.ID)
		assert.Equal(t, StateTriage, current.ToState)
	})

	t.Run("no transitions yet", func(t *testing.T) {
		current, err := repo.GetCurrentState(ctx, "encounter-no-transitions")
		require.NoError(t, err)
		assert.Nil(t, current)
	})
}

func TestRepositoryGetTransitionByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	transition := &FlowStateTransition{
		ID:            "trans-by-id-1",
		EncounterID:   "encounter-by-id-1",
		ToState:       StateArrived,
		TransitionedAt: now,
		ActorType:     ActorTypeSystem,
		CreatedAt:     now,
	}

	require.NoError(t, repo.CreateTransition(ctx, transition))

	t.Run("found", func(t *testing.T) {
		retrieved, err := repo.GetTransitionByID(ctx, "trans-by-id-1")
		require.NoError(t, err)
		assert.Equal(t, "trans-by-id-1", retrieved.ID)
		assert.Equal(t, StateArrived, retrieved.ToState)
	})

	t.Run("not found", func(t *testing.T) {
		retrieved, err := repo.GetTransitionByID(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, retrieved)
	})
}

func TestRepositoryGetTransitionsSince(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	baseTime := time.Now()

	transitions := []*FlowStateTransition{
		{
			ID:            "trans-since-1",
			EncounterID:   "encounter-since-1",
			ToState:       StateArrived,
			TransitionedAt: baseTime,
			ActorType:     ActorTypeSystem,
			CreatedAt:     baseTime,
		},
		{
			ID:            "trans-since-2",
			EncounterID:   "encounter-since-1",
			ToState:       StateTriage,
			TransitionedAt: baseTime.Add(2 * time.Hour),
			ActorType:     ActorTypeUser,
			CreatedAt:     baseTime.Add(2 * time.Hour),
		},
		{
			ID:            "trans-since-3",
			EncounterID:   "encounter-since-1",
			ToState:       StateProviderEval,
			TransitionedAt: baseTime.Add(4 * time.Hour),
			ActorType:     ActorTypeUser,
			CreatedAt:     baseTime.Add(4 * time.Hour),
		},
	}

	for _, tr := range transitions {
		require.NoError(t, repo.CreateTransition(ctx, tr))
	}

	// Get transitions since 1 hour after base time
	since := baseTime.Add(1 * time.Hour)
	results, err := repo.GetTransitionsSince(ctx, "encounter-since-1", since)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "trans-since-2", results[0].ID)
	assert.Equal(t, "trans-since-3", results[1].ID)
}

func TestRepositoryGetTransitionsByState(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	transitions := []*FlowStateTransition{
		{
			ID:            "trans-state-1",
			EncounterID:   "encounter-state-1",
			ToState:       StateArrived,
			TransitionedAt: now,
			ActorType:     ActorTypeSystem,
			CreatedAt:     now,
		},
		{
			ID:            "trans-state-2",
			EncounterID:   "encounter-state-1",
			ToState:       StateTriage,
			TransitionedAt: now.Add(1 * time.Hour),
			ActorType:     ActorTypeUser,
			CreatedAt:     now.Add(1 * time.Hour),
		},
		{
			ID:            "trans-state-3",
			EncounterID:   "encounter-state-1",
			ToState:       StateArrived,
			TransitionedAt: now.Add(2 * time.Hour),
			ActorType:     ActorTypeSystem,
			CreatedAt:     now.Add(2 * time.Hour),
		},
	}

	for _, tr := range transitions {
		require.NoError(t, repo.CreateTransition(ctx, tr))
	}

	results, err := repo.GetTransitionsByState(ctx, "encounter-state-1", StateArrived)
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "trans-state-1", results[0].ID)
	assert.Equal(t, "trans-state-3", results[1].ID)
}
