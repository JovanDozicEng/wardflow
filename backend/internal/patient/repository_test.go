package patient

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wardflow/backend/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testIDCounter = 0

func newTestID() string {
	testIDCounter++
	return fmt.Sprintf("test-id-%d", testIDCounter)
}

func newRepositoryTestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)
	
	// Create table manually for SQLite compatibility (no UUID type or gen_random_uuid)
	err = db.Exec(`
		CREATE TABLE patients (
			id TEXT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			date_of_birth DATETIME,
			mrn TEXT NOT NULL UNIQUE,
			created_by TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)
	
	return &database.DB{DB: db}
}

func TestRepository_NewRepository(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	assert.NotNil(t, repo)
}

func TestRepository_Create(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully creates patient", func(t *testing.T) {
		dob := time.Date(1980, 5, 15, 0, 0, 0, 0, time.UTC)
		patient := &Patient{
			ID:          newTestID(),
			FirstName:   "John",
			LastName:    "Doe",
			DateOfBirth: &dob,
			MRN:         "MRN-001",
			CreatedBy:   "user-1",
		}

		err := repo.Create(ctx, patient)
		require.NoError(t, err)
		assert.NotEmpty(t, patient.ID)
	})

	t.Run("creates patient without date of birth", func(t *testing.T) {
		patient := &Patient{
			ID:        newTestID(),
			FirstName: "Jane",
			LastName:  "Smith",
			MRN:       "MRN-002",
			CreatedBy: "user-1",
		}

		err := repo.Create(ctx, patient)
		require.NoError(t, err)
		assert.NotEmpty(t, patient.ID)
		assert.Nil(t, patient.DateOfBirth)
	})
}

func TestRepository_GetByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("returns patient when found", func(t *testing.T) {
		patient := &Patient{
			ID:        newTestID(),
			FirstName: "Test",
			LastName:  "Patient",
			MRN:       "MRN-TEST",
			CreatedBy: "user-1",
		}
		require.NoError(t, repo.Create(ctx, patient))

		result, err := repo.GetByID(ctx, patient.ID)
		require.NoError(t, err)
		assert.Equal(t, patient.ID, result.ID)
		assert.Equal(t, patient.FirstName, result.FirstName)
		assert.Equal(t, patient.LastName, result.LastName)
		assert.Equal(t, patient.MRN, result.MRN)
	})

	t.Run("returns ErrNotFound when not found", func(t *testing.T) {
		nonExistentID := newTestID()
		result, err := repo.GetByID(ctx, nonExistentID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, result)
	})
}

func TestRepository_List(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Create test data
	patients := []*Patient{
		{ID: newTestID(), FirstName: "Alice", LastName: "Anderson", MRN: "MRN-101", CreatedBy: "user-1"},
		{ID: newTestID(), FirstName: "Bob", LastName: "Brown", MRN: "MRN-102", CreatedBy: "user-1"},
		{ID: newTestID(), FirstName: "Charlie", LastName: "Clark", MRN: "MRN-103", CreatedBy: "user-1"},
		{ID: newTestID(), FirstName: "Diana", LastName: "Davis", MRN: "MRN-104", CreatedBy: "user-1"},
	}
	for _, p := range patients {
		require.NoError(t, repo.Create(ctx, p))
	}

	t.Run("returns all patients when no filters", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListPatientsFilter{})
		require.NoError(t, err)
		assert.Len(t, results, 4)
		assert.Equal(t, int64(4), total)
		// Should be ordered by last name, then first name ASC
		assert.Equal(t, "Alice", results[0].FirstName)
	})

	t.Run("filters by first name", func(t *testing.T) {
		t.Skip("ILIKE is PostgreSQL-specific; skipped for SQLite in-memory tests")
	})

	t.Run("filters by last name", func(t *testing.T) {
		t.Skip("ILIKE is PostgreSQL-specific; skipped for SQLite in-memory tests")
	})

	t.Run("filters by MRN", func(t *testing.T) {
		t.Skip("ILIKE is PostgreSQL-specific; skipped for SQLite in-memory tests")
	})

	t.Run("applies limit", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListPatientsFilter{Limit: 2})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(4), total) // Total should still be 4
	})

	t.Run("applies offset", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListPatientsFilter{Offset: 2})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(4), total)
	})

	t.Run("applies limit and offset for pagination", func(t *testing.T) {
		results, total, err := repo.List(ctx, ListPatientsFilter{Limit: 2, Offset: 1})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(4), total)
		assert.Equal(t, "Bob", results[0].FirstName)
	})

	t.Run("returns empty list when no matches", func(t *testing.T) {
		t.Skip("ILIKE is PostgreSQL-specific; skipped for SQLite in-memory tests")
	})
}
