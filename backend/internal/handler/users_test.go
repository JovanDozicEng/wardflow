package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
	"github.com/wardflow/backend/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock UserService
type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) ListUsers(ctx context.Context, q, role string) ([]UserSummary, error) {
	args := m.Called(ctx, q, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]UserSummary), args.Error(1)
}

func TestUsersHandler_ListUsers(t *testing.T) {
	t.Run("returns list of users", func(t *testing.T) {
		svc := new(mockUserService)
		handler := NewUsersHandler(svc)

		summaries := []UserSummary{
			{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: models.RoleNurse},
			{ID: "user-2", Name: "Bob", Email: "bob@example.com", Role: models.RoleProvider},
		}

		svc.On("ListUsers", mock.Anything, "", "").Return(summaries, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/users", nil, "user-admin", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []UserSummary
		testutil.DecodeJSON(t, rr, &result)
		assert.Len(t, result, 2)
		assert.Equal(t, "Alice", result[0].Name)
		assert.Equal(t, "Bob", result[1].Name)

		svc.AssertExpectations(t)
	})

	t.Run("filters by search query", func(t *testing.T) {
		svc := new(mockUserService)
		handler := NewUsersHandler(svc)

		summaries := []UserSummary{
			{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: models.RoleNurse},
		}

		svc.On("ListUsers", mock.Anything, "alice", "").Return(summaries, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/users?q=alice", nil, "user-admin", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []UserSummary
		testutil.DecodeJSON(t, rr, &result)
		assert.Len(t, result, 1)
		assert.Equal(t, "Alice", result[0].Name)

		svc.AssertExpectations(t)
	})

	t.Run("filters by role", func(t *testing.T) {
		svc := new(mockUserService)
		handler := NewUsersHandler(svc)

		summaries := []UserSummary{
			{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: models.RoleNurse},
			{ID: "user-2", Name: "Charlie", Email: "charlie@example.com", Role: models.RoleNurse},
		}

		svc.On("ListUsers", mock.Anything, "", "nurse").Return(summaries, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/users?role=nurse", nil, "user-admin", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []UserSummary
		testutil.DecodeJSON(t, rr, &result)
		assert.Len(t, result, 2)
		assert.Equal(t, models.RoleNurse, result[0].Role)
		assert.Equal(t, models.RoleNurse, result[1].Role)

		svc.AssertExpectations(t)
	})

	t.Run("filters by both query and role", func(t *testing.T) {
		svc := new(mockUserService)
		handler := NewUsersHandler(svc)

		summaries := []UserSummary{
			{ID: "user-1", Name: "Alice Nurse", Email: "alice@example.com", Role: models.RoleNurse},
		}

		svc.On("ListUsers", mock.Anything, "alice", "nurse").Return(summaries, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/users?q=alice&role=nurse", nil, "user-admin", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []UserSummary
		testutil.DecodeJSON(t, rr, &result)
		assert.Len(t, result, 1)
		assert.Equal(t, "Alice Nurse", result[0].Name)

		svc.AssertExpectations(t)
	})

	t.Run("returns empty list when no users found", func(t *testing.T) {
		svc := new(mockUserService)
		handler := NewUsersHandler(svc)

		svc.On("ListUsers", mock.Anything, "nonexistent", "").Return([]UserSummary{}, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/users?q=nonexistent", nil, "user-admin", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result []UserSummary
		testutil.DecodeJSON(t, rr, &result)
		assert.Len(t, result, 0)

		svc.AssertExpectations(t)
	})

	t.Run("returns internal error on service failure", func(t *testing.T) {
		svc := new(mockUserService)
		handler := NewUsersHandler(svc)

		svc.On("ListUsers", mock.Anything, "", "").Return(nil, errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/users", nil, "user-admin", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListUsers(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})
}

// Tests for concrete userService implementation (not the handler)
// NOTE: Simple integration test to cover the actual userService.ListUsers implementation
func TestUserService_ListUsers_Concrete(t *testing.T) {
	// NOTE: The handler tests already provide good coverage with mocks
	// This test just ensures the concrete implementation runs without error
	db := newUsersTestDB(t)
	svc := NewUserService(db)
	ctx := context.Background()

	// Just test that it doesn't crash with empty DB
	result, err := svc.ListUsers(ctx, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// Helper to create test DB for concrete userService tests
func newUsersTestDB(t *testing.T) *database.DB {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	assert.NoError(t, err)

	// Manually create a simplified users table for SQLite
	// Note: SQLite stores booleans as INTEGER (0 or 1)
	gormDB.Exec(`CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		name TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		role TEXT NOT NULL,
		unit_ids TEXT DEFAULT '[]',
		department_ids TEXT DEFAULT '[]',
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`)

	return &database.DB{DB: gormDB}
}
