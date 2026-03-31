package careteam

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
	err = db.Exec(`CREATE TABLE care_team_assignments (
		id TEXT PRIMARY KEY,
		encounter_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		role_type TEXT NOT NULL,
		starts_at DATETIME NOT NULL,
		ends_at DATETIME,
		created_by TEXT NOT NULL,
		created_at DATETIME,
		handoff_note_id TEXT
	)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE TABLE handoff_notes (
		id TEXT PRIMARY KEY,
		encounter_id TEXT NOT NULL,
		from_user_id TEXT NOT NULL,
		to_user_id TEXT NOT NULL,
		role_type TEXT NOT NULL,
		note TEXT NOT NULL,
		structured_fields_json TEXT,
		created_at DATETIME
	)`).Error
	require.NoError(t, err)
	return &database.DB{DB: db}
}

func TestRepositoryCreateAssignment(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	assignment := &CareTeamAssignment{
		ID:          "assignment-1",
		EncounterID: "encounter-1",
		UserID:      "user-1",
		RoleType:    RolePrimaryNurse,
		StartsAt:    now,
		CreatedBy:   "admin-1",
		CreatedAt:   now,
	}

	err := repo.CreateAssignment(ctx, assignment)
	require.NoError(t, err)

	// Verify
	retrieved, err := repo.GetAssignmentByID(ctx, "assignment-1")
	require.NoError(t, err)
	assert.Equal(t, "assignment-1", retrieved.ID)
	assert.Equal(t, RolePrimaryNurse, retrieved.RoleType)
	assert.True(t, retrieved.IsActive())
}

func TestRepositoryEndAssignment(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	assignment := &CareTeamAssignment{
		ID:          "assignment-end-1",
		EncounterID: "encounter-end-1",
		UserID:      "user-1",
		RoleType:    RolePrimaryNurse,
		StartsAt:    now,
		CreatedBy:   "admin-1",
		CreatedAt:   now,
	}

	require.NoError(t, repo.CreateAssignment(ctx, assignment))

	t.Run("end active assignment", func(t *testing.T) {
		endsAt := now.Add(2 * time.Hour)
		err := repo.EndAssignment(ctx, "assignment-end-1", endsAt)
		require.NoError(t, err)

		retrieved, err := repo.GetAssignmentByID(ctx, "assignment-end-1")
		require.NoError(t, err)
		assert.NotNil(t, retrieved.EndsAt)
		assert.False(t, retrieved.IsActive())
	})

	t.Run("end non-existent assignment", func(t *testing.T) {
		err := repo.EndAssignment(ctx, "non-existent", time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found or already ended")
	})

	t.Run("end already ended assignment", func(t *testing.T) {
		err := repo.EndAssignment(ctx, "assignment-end-1", time.Now())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found or already ended")
	})
}

func TestRepositoryGetActiveAssignments(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	endsAt := now.Add(1 * time.Hour)

	assignments := []*CareTeamAssignment{
		{
			ID:          "assign-active-1",
			EncounterID: "encounter-active-1",
			UserID:      "user-1",
			RoleType:    RolePrimaryNurse,
			StartsAt:    now,
			CreatedBy:   "admin-1",
			CreatedAt:   now,
		},
		{
			ID:          "assign-active-2",
			EncounterID: "encounter-active-1",
			UserID:      "user-2",
			RoleType:    RoleAttendingProvider,
			StartsAt:    now,
			CreatedBy:   "admin-1",
			CreatedAt:   now,
		},
		{
			ID:          "assign-ended-1",
			EncounterID: "encounter-active-1",
			UserID:      "user-3",
			RoleType:    RolePrimaryNurse,
			StartsAt:    now.Add(-5 * time.Hour),
			EndsAt:      &endsAt,
			CreatedBy:   "admin-1",
			CreatedAt:   now.Add(-5 * time.Hour),
		},
	}

	for _, a := range assignments {
		require.NoError(t, repo.CreateAssignment(ctx, a))
	}

	activeAssignments, err := repo.GetActiveAssignments(ctx, "encounter-active-1")
	require.NoError(t, err)
	assert.Len(t, activeAssignments, 2)

	// Verify only active ones are returned
	for _, a := range activeAssignments {
		assert.Nil(t, a.EndsAt)
	}
}

func TestRepositoryGetActiveAssignmentByRole(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	assignments := []*CareTeamAssignment{
		{
			ID:          "assign-role-1",
			EncounterID: "encounter-role-1",
			UserID:      "user-1",
			RoleType:    RolePrimaryNurse,
			StartsAt:    now,
			CreatedBy:   "admin-1",
			CreatedAt:   now,
		},
		{
			ID:          "assign-role-2",
			EncounterID: "encounter-role-1",
			UserID:      "user-2",
			RoleType:    RoleAttendingProvider,
			StartsAt:    now,
			CreatedBy:   "admin-1",
			CreatedAt:   now,
		},
	}

	for _, a := range assignments {
		require.NoError(t, repo.CreateAssignment(ctx, a))
	}

	t.Run("found active assignment", func(t *testing.T) {
		assignment, err := repo.GetActiveAssignmentByRole(ctx, "encounter-role-1", RolePrimaryNurse)
		require.NoError(t, err)
		assert.NotNil(t, assignment)
		assert.Equal(t, "assign-role-1", assignment.ID)
		assert.Equal(t, RolePrimaryNurse, assignment.RoleType)
	})

	t.Run("no active assignment for role", func(t *testing.T) {
		assignment, err := repo.GetActiveAssignmentByRole(ctx, "encounter-role-1", RoleCaseManager)
		require.NoError(t, err)
		assert.Nil(t, assignment)
	})
}

func TestRepositoryGetAssignmentHistory(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	endsAt := now.Add(2 * time.Hour)

	assignments := []*CareTeamAssignment{
		{
			ID:          "assign-hist-1",
			EncounterID: "encounter-hist-1",
			UserID:      "user-1",
			RoleType:    RolePrimaryNurse,
			StartsAt:    now,
			CreatedBy:   "admin-1",
			CreatedAt:   now,
		},
		{
			ID:          "assign-hist-2",
			EncounterID: "encounter-hist-1",
			UserID:      "user-2",
			RoleType:    RoleAttendingProvider,
			StartsAt:    now.Add(1 * time.Hour),
			EndsAt:      &endsAt,
			CreatedBy:   "admin-1",
			CreatedAt:   now.Add(1 * time.Hour),
		},
	}

	for _, a := range assignments {
		require.NoError(t, repo.CreateAssignment(ctx, a))
	}

	history, err := repo.GetAssignmentHistory(ctx, "encounter-hist-1")
	require.NoError(t, err)
	assert.Len(t, history, 2)
	// Ordered by starts_at DESC
	assert.Equal(t, "assign-hist-2", history[0].ID)
	assert.Equal(t, "assign-hist-1", history[1].ID)
}

func TestRepositoryGetAssignmentByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	assignment := &CareTeamAssignment{
		ID:          "assign-by-id-1",
		EncounterID: "encounter-by-id-1",
		UserID:      "user-1",
		RoleType:    RolePrimaryNurse,
		StartsAt:    now,
		CreatedBy:   "admin-1",
		CreatedAt:   now,
	}

	require.NoError(t, repo.CreateAssignment(ctx, assignment))

	t.Run("found", func(t *testing.T) {
		retrieved, err := repo.GetAssignmentByID(ctx, "assign-by-id-1")
		require.NoError(t, err)
		assert.Equal(t, "assign-by-id-1", retrieved.ID)
		assert.Equal(t, RolePrimaryNurse, retrieved.RoleType)
	})

	t.Run("not found", func(t *testing.T) {
		retrieved, err := repo.GetAssignmentByID(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, retrieved)
	})
}

func TestRepositoryCreateHandoffNote(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	note := &HandoffNote{
		ID:          "handoff-1",
		EncounterID: "encounter-1",
		FromUserID:  "user-1",
		ToUserID:    "user-2",
		RoleType:    RolePrimaryNurse,
		Note:        "Patient stable, vitals normal.",
		CreatedAt:   now,
	}

	err := repo.CreateHandoffNote(ctx, note)
	require.NoError(t, err)

	// Verify by getting handoff notes
	notes, err := repo.GetHandoffNotes(ctx, "encounter-1")
	require.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.Equal(t, "handoff-1", notes[0].ID)
	assert.Equal(t, "Patient stable, vitals normal.", notes[0].Note)
}

func TestRepositoryGetHandoffNotes(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	notes := []*HandoffNote{
		{
			ID:          "handoff-note-1",
			EncounterID: "encounter-note-1",
			FromUserID:  "user-1",
			ToUserID:    "user-2",
			RoleType:    RolePrimaryNurse,
			Note:        "First handoff",
			CreatedAt:   now,
		},
		{
			ID:          "handoff-note-2",
			EncounterID: "encounter-note-1",
			FromUserID:  "user-2",
			ToUserID:    "user-3",
			RoleType:    RolePrimaryNurse,
			Note:        "Second handoff",
			CreatedAt:   now.Add(1 * time.Hour),
		},
	}

	for _, n := range notes {
		require.NoError(t, repo.CreateHandoffNote(ctx, n))
	}

	results, err := repo.GetHandoffNotes(ctx, "encounter-note-1")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	// Ordered by created_at DESC
	assert.Equal(t, "handoff-note-2", results[0].ID)
	assert.Equal(t, "handoff-note-1", results[1].ID)

	t.Run("no handoff notes for non-existent encounter", func(t *testing.T) {
		results, err := repo.GetHandoffNotes(ctx, "non-existent")
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestRepositoryGetHandoffNotesPaginated(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	notes := []*HandoffNote{
		{
			ID:          "handoff-page-1",
			EncounterID: "encounter-page-1",
			FromUserID:  "user-1",
			ToUserID:    "user-2",
			RoleType:    RolePrimaryNurse,
			Note:        "First handoff",
			CreatedAt:   now,
		},
		{
			ID:          "handoff-page-2",
			EncounterID: "encounter-page-1",
			FromUserID:  "user-2",
			ToUserID:    "user-3",
			RoleType:    RolePrimaryNurse,
			Note:        "Second handoff",
			CreatedAt:   now.Add(1 * time.Hour),
		},
		{
			ID:          "handoff-page-3",
			EncounterID: "encounter-page-1",
			FromUserID:  "user-3",
			ToUserID:    "user-4",
			RoleType:    RolePrimaryNurse,
			Note:        "Third handoff",
			CreatedAt:   now.Add(2 * time.Hour),
		},
	}

	for _, n := range notes {
		require.NoError(t, repo.CreateHandoffNote(ctx, n))
	}

	t.Run("all results", func(t *testing.T) {
		results, total, err := repo.GetHandoffNotesPaginated(ctx, "encounter-page-1", -1, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 3)
		// Ordered by created_at DESC
		assert.Equal(t, "handoff-page-3", results[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		results, total, err := repo.GetHandoffNotesPaginated(ctx, "encounter-page-1", 1, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, results, 1)
		assert.Equal(t, "handoff-page-2", results[0].ID)
	})
}
