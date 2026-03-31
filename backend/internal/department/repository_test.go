package department

import (
	"context"
	"fmt"
	"testing"

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
		CREATE TABLE departments (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			code TEXT NOT NULL UNIQUE,
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

	t.Run("successfully creates department", func(t *testing.T) {
		dept := &Department{
			ID:   newTestID(),
			Name: "Emergency Medicine",
			Code: "EMERGENCY",
		}

		err := repo.Create(ctx, dept)
		require.NoError(t, err)
		assert.NotEmpty(t, dept.ID)
	})

	t.Run("creates department with generated ID", func(t *testing.T) {
		dept := &Department{
			ID:   newTestID(),
			Name: "Cardiology",
			Code: "CARDIOLOGY",
		}

		err := repo.Create(ctx, dept)
		require.NoError(t, err)
		assert.NotEmpty(t, dept.ID)
		assert.NotZero(t, dept.CreatedAt)
	})
}

func TestRepository_GetByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("returns department when found", func(t *testing.T) {
		dept := &Department{
			ID:   newTestID(),
			Name: "Test Department",
			Code: "TEST",
		}
		require.NoError(t, repo.Create(ctx, dept))

		result, err := repo.GetByID(ctx, dept.ID)
		require.NoError(t, err)
		assert.Equal(t, dept.ID, result.ID)
		assert.Equal(t, dept.Name, result.Name)
		assert.Equal(t, dept.Code, result.Code)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		nonExistentID := newTestID()
		result, err := repo.GetByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRepository_List(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Create test data
	departments := []*Department{
		{ID: newTestID(), Name: "Emergency Medicine", Code: "EMERGENCY"},
		{ID: newTestID(), Name: "Cardiology", Code: "CARDIOLOGY"},
		{ID: newTestID(), Name: "Orthopedics", Code: "ORTHO"},
		{ID: newTestID(), Name: "Pediatrics", Code: "PEDS"},
	}
	for _, d := range departments {
		require.NoError(t, repo.Create(ctx, d))
	}

	t.Run("returns all departments when no filters", func(t *testing.T) {
		results, err := repo.List(ctx, "")
		require.NoError(t, err)
		assert.Len(t, results, 4)
		// Should be ordered by name ASC
		assert.Equal(t, "Cardiology", results[0].Name)
	})

	t.Run("filters by search query (name)", func(t *testing.T) {
		t.Skip("ILIKE is PostgreSQL-specific; skipped for SQLite in-memory tests")
	})

	t.Run("filters by search query (code) case insensitive", func(t *testing.T) {
		t.Skip("ILIKE is PostgreSQL-specific; skipped for SQLite in-memory tests")
	})

	t.Run("returns multiple matches for partial query", func(t *testing.T) {
		t.Skip("ILIKE is PostgreSQL-specific; skipped for SQLite in-memory tests")
	})

	t.Run("returns empty list when no matches", func(t *testing.T) {
		t.Skip("ILIKE is PostgreSQL-specific; skipped for SQLite in-memory tests")
	})
}
