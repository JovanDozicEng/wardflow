package transport

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
	
	// Create tables manually for SQLite compatibility (no UUID type or gen_random_uuid)
	err = db.Exec(`
		CREATE TABLE transport_requests (
			id TEXT PRIMARY KEY,
			encounter_id TEXT NOT NULL,
			origin TEXT NOT NULL,
			destination TEXT NOT NULL,
			priority VARCHAR(20) NOT NULL DEFAULT 'routine',
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			assigned_to TEXT,
			assigned_at DATETIME,
			created_by TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)
	
	err = db.Exec(`
		CREATE TABLE transport_change_events (
			id TEXT PRIMARY KEY,
			request_id TEXT NOT NULL,
			changed_fields TEXT NOT NULL,
			changed_by TEXT NOT NULL,
			reason TEXT,
			changed_at DATETIME NOT NULL,
			created_at DATETIME
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

func TestRepository_CreateRequest(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully creates transport request", func(t *testing.T) {
		req := &TransportRequest{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "ICU Room 101",
			Destination: "Radiology",
			Priority:    "urgent",
			Status:      TransportStatusPending,
			CreatedBy:   "user-1",
		}

		err := repo.CreateRequest(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, req.ID)
	})

	t.Run("creates transport request with assigned user", func(t *testing.T) {
		assignedTo := newTestID()
		assignedAt := time.Now().UTC()
		req := &TransportRequest{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "ED Bay 5",
			Destination: "CT Scan",
			Priority:    "emergent",
			Status:      TransportStatusAssigned,
			AssignedTo:  &assignedTo,
			AssignedAt:  &assignedAt,
			CreatedBy:   "user-1",
		}

		err := repo.CreateRequest(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, req.ID)
		assert.NotNil(t, req.AssignedTo)
	})
}

func TestRepository_GetRequestByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("returns transport request when found", func(t *testing.T) {
		req := &TransportRequest{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "Ward 3A",
			Destination: "Operating Room 2",
			Priority:    "urgent",
			Status:      TransportStatusPending,
			CreatedBy:   "user-1",
		}
		require.NoError(t, repo.CreateRequest(ctx, req))

		result, err := repo.GetRequestByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, req.ID, result.ID)
		assert.Equal(t, req.Origin, result.Origin)
		assert.Equal(t, req.Destination, result.Destination)
		assert.Equal(t, req.Priority, result.Priority)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		nonExistentID := newTestID()
		result, err := repo.GetRequestByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRepository_ListRequests(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Create test data
	requests := []*TransportRequest{
		{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "ICU 101",
			Destination: "Radiology",
			Priority:    "urgent",
			Status:      TransportStatusPending,
			CreatedBy:   "user-1",
		},
		{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "ED Bay 2",
			Destination: "MRI",
			Priority:    "routine",
			Status:      TransportStatusAssigned,
			CreatedBy:   "user-1",
		},
		{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "Ward 4B",
			Destination: "CT Scan",
			Priority:    "emergent",
			Status:      TransportStatusInTransit,
			CreatedBy:   "user-1",
		},
		{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "ICU 202",
			Destination: "X-Ray",
			Priority:    "urgent",
			Status:      TransportStatusCompleted,
			CreatedBy:   "user-1",
		},
	}
	for _, r := range requests {
		require.NoError(t, repo.CreateRequest(ctx, r))
		// Small sleep to ensure different created_at times
		time.Sleep(5 * time.Millisecond)
	}

	t.Run("returns all requests when no filters", func(t *testing.T) {
		results, total, err := repo.ListRequests(ctx, ListTransportFilter{Limit: -1})
		require.NoError(t, err)
		assert.Len(t, results, 4)
		assert.Equal(t, int64(4), total)
		// Should be ordered by created_at DESC
		assert.Equal(t, "ICU 202", results[0].Origin) // Most recent
	})

	t.Run("filters by status", func(t *testing.T) {
		results, total, err := repo.ListRequests(ctx, ListTransportFilter{
			Status: string(TransportStatusPending),
			Limit:  -1,
		})
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, TransportStatusPending, results[0].Status)
	})

	t.Run("applies limit", func(t *testing.T) {
		results, total, err := repo.ListRequests(ctx, ListTransportFilter{Limit: 2})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(4), total)
	})

	t.Run("applies offset", func(t *testing.T) {
		results, total, err := repo.ListRequests(ctx, ListTransportFilter{Offset: 2, Limit: -1})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(4), total)
	})

	t.Run("applies limit and offset for pagination", func(t *testing.T) {
		results, total, err := repo.ListRequests(ctx, ListTransportFilter{
			Limit:  2,
			Offset: 1,
		})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(4), total)
	})
}

func TestRepository_UpdateRequestFields(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully updates transport request fields", func(t *testing.T) {
		req := &TransportRequest{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "Ward 3A",
			Destination: "Radiology",
			Priority:    "routine",
			Status:      TransportStatusPending,
			CreatedBy:   "user-1",
		}
		require.NoError(t, repo.CreateRequest(ctx, req))

		// Update status
		updates := map[string]any{
			"status": TransportStatusAssigned,
		}
		err := repo.UpdateRequestFields(ctx, req.ID, updates)
		require.NoError(t, err)

		// Verify update
		result, err := repo.GetRequestByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, TransportStatusAssigned, result.Status)
	})

	t.Run("updates multiple fields", func(t *testing.T) {
		req := &TransportRequest{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "ICU 101",
			Destination: "CT Scan",
			Priority:    "routine",
			Status:      TransportStatusPending,
			CreatedBy:   "user-1",
		}
		require.NoError(t, repo.CreateRequest(ctx, req))

		assignedTo := newTestID()
		assignedAt := time.Now().UTC()
		updates := map[string]any{
			"status":      TransportStatusAssigned,
			"assigned_to": assignedTo,
			"assigned_at": assignedAt,
		}
		err := repo.UpdateRequestFields(ctx, req.ID, updates)
		require.NoError(t, err)

		result, err := repo.GetRequestByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, TransportStatusAssigned, result.Status)
		assert.NotNil(t, result.AssignedTo)
		assert.Equal(t, assignedTo, *result.AssignedTo)
	})

	t.Run("updates origin and destination", func(t *testing.T) {
		req := &TransportRequest{
			ID:          newTestID(),
			EncounterID: newTestID(),
			Origin:      "Ward 2B",
			Destination: "X-Ray",
			Priority:    "urgent",
			Status:      TransportStatusPending,
			CreatedBy:   "user-1",
		}
		require.NoError(t, repo.CreateRequest(ctx, req))

		updates := map[string]any{
			"origin":      "Ward 3C",
			"destination": "MRI",
		}
		err := repo.UpdateRequestFields(ctx, req.ID, updates)
		require.NoError(t, err)

		result, err := repo.GetRequestByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, "Ward 3C", result.Origin)
		assert.Equal(t, "MRI", result.Destination)
	})
}

func TestRepository_CreateChangeEvent(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully creates change event", func(t *testing.T) {
		event := &TransportChangeEvent{
			ID:            newTestID(),
			RequestID:     newTestID(),
			ChangedFields: `{"status":"assigned"}`,
			ChangedBy:     "user-1",
			ChangedAt:     time.Now().UTC(),
		}

		err := repo.CreateChangeEvent(ctx, event)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("creates change event with reason", func(t *testing.T) {
		reason := "Patient condition changed"
		event := &TransportChangeEvent{
			ID:            newTestID(),
			RequestID:     newTestID(),
			ChangedFields: `{"priority":"emergent"}`,
			ChangedBy:     "user-1",
			Reason:        &reason,
			ChangedAt:     time.Now().UTC(),
		}

		err := repo.CreateChangeEvent(ctx, event)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("creates change event with multiple field changes", func(t *testing.T) {
		event := &TransportChangeEvent{
			ID:            newTestID(),
			RequestID:     newTestID(),
			ChangedFields: `{"status":"in_transit","destination":"MRI Suite 2"}`,
			ChangedBy:     "user-2",
			ChangedAt:     time.Now().UTC(),
		}

		err := repo.CreateChangeEvent(ctx, event)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
	})
}
