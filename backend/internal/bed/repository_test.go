package bed

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
		CREATE TABLE beds (
			id TEXT PRIMARY KEY,
			unit_id TEXT NOT NULL,
			room TEXT NOT NULL,
			label TEXT NOT NULL,
			capabilities TEXT DEFAULT '[]',
			current_status VARCHAR(20) NOT NULL DEFAULT 'available',
			current_encounter_id TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)
	
	err = db.Exec(`
		CREATE TABLE bed_status_events (
			id TEXT PRIMARY KEY,
			bed_id TEXT NOT NULL,
			from_status VARCHAR(20),
			to_status VARCHAR(20) NOT NULL,
			reason TEXT,
			changed_by TEXT NOT NULL,
			changed_at DATETIME NOT NULL,
			created_at DATETIME
		)
	`).Error
	require.NoError(t, err)
	
	err = db.Exec(`
		CREATE TABLE bed_requests (
			id TEXT PRIMARY KEY,
			encounter_id TEXT NOT NULL,
			required_capabilities TEXT DEFAULT '[]',
			priority VARCHAR(20) NOT NULL DEFAULT 'routine',
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			assigned_bed_id TEXT,
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

func TestRepository_CreateBed(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully creates bed", func(t *testing.T) {
		bed := &Bed{
			ID:            newTestID(),
			UnitID:        newTestID(),
			Room:          "101",
			Label:         "Bed A",
			Capabilities:  StringSlice{"ventilator", "monitor"},
			CurrentStatus: BedStatusAvailable,
		}

		err := repo.CreateBed(ctx, bed)
		require.NoError(t, err)
		assert.NotEmpty(t, bed.ID)
	})

	t.Run("creates bed with empty capabilities", func(t *testing.T) {
		bed := &Bed{
			ID:            newTestID(),
			UnitID:        newTestID(),
			Room:          "102",
			Label:         "Bed B",
			Capabilities:  StringSlice{},
			CurrentStatus: BedStatusAvailable,
		}

		err := repo.CreateBed(ctx, bed)
		require.NoError(t, err)
		assert.NotEmpty(t, bed.ID)
	})
}

func TestRepository_GetBedByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("returns bed when found", func(t *testing.T) {
		bed := &Bed{
			ID:            newTestID(),
			UnitID:        newTestID(),
			Room:          "101",
			Label:         "Bed A",
			Capabilities:  StringSlice{"ventilator"},
			CurrentStatus: BedStatusAvailable,
		}
		require.NoError(t, repo.CreateBed(ctx, bed))

		result, err := repo.GetBedByID(ctx, bed.ID)
		require.NoError(t, err)
		assert.Equal(t, bed.ID, result.ID)
		assert.Equal(t, bed.Room, result.Room)
		assert.Equal(t, bed.Label, result.Label)
		assert.Equal(t, BedStatusAvailable, result.CurrentStatus)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		nonExistentID := newTestID()
		result, err := repo.GetBedByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRepository_ListBeds(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	unitID1 := newTestID()
	unitID2 := newTestID()

	// Create test data
	beds := []*Bed{
		{ID: newTestID(), UnitID: unitID1, Room: "101", Label: "Bed A", CurrentStatus: BedStatusAvailable},
		{ID: newTestID(), UnitID: unitID1, Room: "102", Label: "Bed B", CurrentStatus: BedStatusOccupied},
		{ID: newTestID(), UnitID: unitID2, Room: "201", Label: "Bed C", CurrentStatus: BedStatusAvailable},
		{ID: newTestID(), UnitID: unitID2, Room: "202", Label: "Bed D", CurrentStatus: BedStatusCleaning},
	}
	for _, b := range beds {
		require.NoError(t, repo.CreateBed(ctx, b))
	}

	t.Run("returns all beds when no filters", func(t *testing.T) {
		results, total, err := repo.ListBeds(ctx, ListBedsFilter{Limit: -1})
		require.NoError(t, err)
		assert.Len(t, results, 4)
		assert.Equal(t, int64(4), total)
	})

	t.Run("filters by unit ID", func(t *testing.T) {
		results, total, err := repo.ListBeds(ctx, ListBedsFilter{UnitID: unitID1, Limit: -1})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(2), total)
		for _, b := range results {
			assert.Equal(t, unitID1, b.UnitID)
		}
	})

	t.Run("filters by status", func(t *testing.T) {
		results, total, err := repo.ListBeds(ctx, ListBedsFilter{Status: string(BedStatusAvailable), Limit: -1})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(2), total)
		for _, b := range results {
			assert.Equal(t, BedStatusAvailable, b.CurrentStatus)
		}
	})

	t.Run("filters by multiple unit IDs", func(t *testing.T) {
		results, total, err := repo.ListBeds(ctx, ListBedsFilter{UnitIDs: []string{unitID1, unitID2}, Limit: -1})
		require.NoError(t, err)
		assert.Len(t, results, 4)
		assert.Equal(t, int64(4), total)
	})

	t.Run("applies limit", func(t *testing.T) {
		results, total, err := repo.ListBeds(ctx, ListBedsFilter{Limit: 2})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(4), total)
	})

	t.Run("applies offset", func(t *testing.T) {
		results, total, err := repo.ListBeds(ctx, ListBedsFilter{Offset: 2, Limit: -1})
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, int64(4), total)
	})
}

func TestRepository_UpdateBedFields(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully updates bed fields", func(t *testing.T) {
		bed := &Bed{
			ID:            newTestID(),
			UnitID:        newTestID(),
			Room:          "101",
			Label:         "Bed A",
			CurrentStatus: BedStatusAvailable,
		}
		require.NoError(t, repo.CreateBed(ctx, bed))

		// Update bed status
		updates := map[string]any{
			"current_status": BedStatusOccupied,
		}
		err := repo.UpdateBedFields(ctx, bed.ID, updates)
		require.NoError(t, err)

		// Verify update
		result, err := repo.GetBedByID(ctx, bed.ID)
		require.NoError(t, err)
		assert.Equal(t, BedStatusOccupied, result.CurrentStatus)
	})

	t.Run("updates multiple fields", func(t *testing.T) {
		bed := &Bed{
			ID:            newTestID(),
			UnitID:        newTestID(),
			Room:          "102",
			Label:         "Bed B",
			CurrentStatus: BedStatusAvailable,
		}
		require.NoError(t, repo.CreateBed(ctx, bed))

		encounterID := newTestID()
		updates := map[string]any{
			"current_status":       BedStatusOccupied,
			"current_encounter_id": encounterID,
		}
		err := repo.UpdateBedFields(ctx, bed.ID, updates)
		require.NoError(t, err)

		result, err := repo.GetBedByID(ctx, bed.ID)
		require.NoError(t, err)
		assert.Equal(t, BedStatusOccupied, result.CurrentStatus)
		assert.NotNil(t, result.CurrentEncounterID)
		assert.Equal(t, encounterID, *result.CurrentEncounterID)
	})
}

func TestRepository_CreateBedStatusEvent(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully creates bed status event", func(t *testing.T) {
		fromStatus := BedStatusAvailable
		event := &BedStatusEvent{
			ID:         newTestID(),
			BedID:      newTestID(),
			FromStatus: &fromStatus,
			ToStatus:   BedStatusOccupied,
			ChangedBy:  "user-1",
			ChangedAt:  time.Now().UTC(),
		}

		err := repo.CreateBedStatusEvent(ctx, event)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
	})

	t.Run("creates event with reason", func(t *testing.T) {
		reason := "Patient admitted"
		fromStatus := BedStatusAvailable
		event := &BedStatusEvent{
			ID:         newTestID(),
			BedID:      newTestID(),
			FromStatus: &fromStatus,
			ToStatus:   BedStatusOccupied,
			Reason:     &reason,
			ChangedBy:  "user-1",
			ChangedAt:  time.Now().UTC(),
		}

		err := repo.CreateBedStatusEvent(ctx, event)
		require.NoError(t, err)
		assert.NotEmpty(t, event.ID)
	})
}

func TestRepository_CreateBedRequest(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully creates bed request", func(t *testing.T) {
		req := &BedRequest{
			ID:                   newTestID(),
			EncounterID:          newTestID(),
			RequiredCapabilities: StringSlice{"ventilator"},
			Priority:             "urgent",
			Status:               BedRequestStatusPending,
			CreatedBy:            "user-1",
		}

		err := repo.CreateBedRequest(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, req.ID)
	})

	t.Run("creates bed request with empty capabilities", func(t *testing.T) {
		req := &BedRequest{
			ID:                   newTestID(),
			EncounterID:          newTestID(),
			RequiredCapabilities: StringSlice{},
			Priority:             "routine",
			Status:               BedRequestStatusPending,
			CreatedBy:            "user-1",
		}

		err := repo.CreateBedRequest(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, req.ID)
	})
}

func TestRepository_GetBedRequestByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("returns bed request when found", func(t *testing.T) {
		req := &BedRequest{
			ID:                   newTestID(),
			EncounterID:          newTestID(),
			RequiredCapabilities: StringSlice{"monitor"},
			Priority:             "urgent",
			Status:               BedRequestStatusPending,
			CreatedBy:            "user-1",
		}
		require.NoError(t, repo.CreateBedRequest(ctx, req))

		result, err := repo.GetBedRequestByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, req.ID, result.ID)
		assert.Equal(t, req.EncounterID, result.EncounterID)
		assert.Equal(t, req.Priority, result.Priority)
	})

	t.Run("returns error when not found", func(t *testing.T) {
		nonExistentID := newTestID()
		result, err := repo.GetBedRequestByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestRepository_UpdateBedRequestFields(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully updates bed request fields", func(t *testing.T) {
		req := &BedRequest{
			ID:                   newTestID(),
			EncounterID:          newTestID(),
			RequiredCapabilities: StringSlice{},
			Priority:             "routine",
			Status:               BedRequestStatusPending,
			CreatedBy:            "user-1",
		}
		require.NoError(t, repo.CreateBedRequest(ctx, req))

		updates := map[string]any{
			"status": BedRequestStatusAssigned,
		}
		err := repo.UpdateBedRequestFields(ctx, req.ID, updates)
		require.NoError(t, err)

		result, err := repo.GetBedRequestByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, BedRequestStatusAssigned, result.Status)
	})
}

func TestRepository_AssignBed(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	t.Run("successfully assigns bed to request", func(t *testing.T) {
		// Create a bed
		bed := &Bed{
			ID:            newTestID(),
			UnitID:        newTestID(),
			Room:          "101",
			Label:         "Bed A",
			CurrentStatus: BedStatusAvailable,
		}
		require.NoError(t, repo.CreateBed(ctx, bed))

		// Create a bed request
		req := &BedRequest{
			ID:                   newTestID(),
			EncounterID:          newTestID(),
			RequiredCapabilities: StringSlice{},
			Priority:             "urgent",
			Status:               BedRequestStatusPending,
			CreatedBy:            "user-1",
		}
		require.NoError(t, repo.CreateBedRequest(ctx, req))

		// Assign the bed
		err := repo.AssignBed(ctx, req.ID, bed.ID, req.EncounterID, "user-1", BedStatusAvailable)
		require.NoError(t, err)

		// Verify bed status updated
		updatedBed, err := repo.GetBedByID(ctx, bed.ID)
		require.NoError(t, err)
		assert.Equal(t, BedStatusOccupied, updatedBed.CurrentStatus)
		assert.NotNil(t, updatedBed.CurrentEncounterID)
		assert.Equal(t, req.EncounterID, *updatedBed.CurrentEncounterID)

		// Verify bed request updated
		updatedReq, err := repo.GetBedRequestByID(ctx, req.ID)
		require.NoError(t, err)
		assert.Equal(t, BedRequestStatusAssigned, updatedReq.Status)
		assert.NotNil(t, updatedReq.AssignedBedID)
		assert.Equal(t, bed.ID, *updatedReq.AssignedBedID)
	})

	t.Run("fails when bed is not available", func(t *testing.T) {
		// Create an occupied bed
		bed := &Bed{
			ID:            newTestID(),
			UnitID:        newTestID(),
			Room:          "102",
			Label:         "Bed B",
			CurrentStatus: BedStatusOccupied,
		}
		require.NoError(t, repo.CreateBed(ctx, bed))

		// Create a bed request
		req := &BedRequest{
			ID:                   newTestID(),
			EncounterID:          newTestID(),
			RequiredCapabilities: StringSlice{},
			Priority:             "urgent",
			Status:               BedRequestStatusPending,
			CreatedBy:            "user-1",
		}
		require.NoError(t, repo.CreateBedRequest(ctx, req))

		// Try to assign the occupied bed
		err := repo.AssignBed(ctx, req.ID, bed.ID, req.EncounterID, "user-1", BedStatusAvailable)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no longer available")
	})

	t.Run("fails when bed does not exist", func(t *testing.T) {
		// Create a bed request
		req := &BedRequest{
			ID:                   newTestID(),
			EncounterID:          newTestID(),
			RequiredCapabilities: StringSlice{},
			Priority:             "urgent",
			Status:               BedRequestStatusPending,
			CreatedBy:            "user-1",
		}
		require.NoError(t, repo.CreateBedRequest(ctx, req))

		// Try to assign a non-existent bed
		nonExistentBedID := newTestID()
		err := repo.AssignBed(ctx, req.ID, nonExistentBedID, req.EncounterID, "user-1", BedStatusAvailable)
		assert.Error(t, err)
	})
}
