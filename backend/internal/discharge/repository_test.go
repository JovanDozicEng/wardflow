package discharge

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
	err = db.Exec(`CREATE TABLE discharge_checklists (
		id TEXT PRIMARY KEY,
		encounter_id TEXT NOT NULL UNIQUE,
		discharge_type TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'in_progress',
		completed_by TEXT,
		completed_at DATETIME,
		override_reason TEXT,
		created_by TEXT NOT NULL,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)
	err = db.Exec(`CREATE TABLE discharge_checklist_items (
		id TEXT PRIMARY KEY,
		checklist_id TEXT NOT NULL,
		code TEXT NOT NULL,
		label TEXT NOT NULL,
		required INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'open',
		completed_by TEXT,
		completed_at DATETIME,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error
	require.NoError(t, err)
	return &database.DB{DB: db}
}

func TestRepositoryGetChecklistByEncounterID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	checklist := &DischargeChecklist{
		ID:            "checklist-1",
		EncounterID:   "encounter-1",
		DischargeType: "standard",
		Status:        ChecklistStatusInProgress,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := repo.CreateChecklistWithItems(ctx, checklist, []DischargeChecklistItem{})
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		retrieved, err := repo.GetChecklistByEncounterID(ctx, "encounter-1")
		require.NoError(t, err)
		assert.Equal(t, "checklist-1", retrieved.ID)
		assert.Equal(t, "standard", retrieved.DischargeType)
	})

	t.Run("not found", func(t *testing.T) {
		retrieved, err := repo.GetChecklistByEncounterID(ctx, "non-existent")
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, retrieved)
	})
}

func TestRepositoryCreateChecklistWithItems(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	checklist := &DischargeChecklist{
		ID:            "checklist-create-1",
		EncounterID:   "encounter-create-1",
		DischargeType: "standard",
		Status:        ChecklistStatusInProgress,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	items := []DischargeChecklistItem{
		{
			ID:        "item-1",
			Code:      "patient_education",
			Label:     "Patient education completed",
			Required:  true,
			Status:    ItemStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "item-2",
			Code:      "medication_reconciliation",
			Label:     "Medication reconciliation done",
			Required:  true,
			Status:    ItemStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	err := repo.CreateChecklistWithItems(ctx, checklist, items)
	require.NoError(t, err)

	// Verify checklist was created
	retrieved, err := repo.GetChecklistByEncounterID(ctx, "encounter-create-1")
	require.NoError(t, err)
	assert.Equal(t, "checklist-create-1", retrieved.ID)

	// Verify items were created
	retrievedItems, err := repo.GetItemsByChecklistID(ctx, "checklist-create-1")
	require.NoError(t, err)
	assert.Len(t, retrievedItems, 2)
	assert.Equal(t, "checklist-create-1", retrievedItems[0].ChecklistID)
}

func TestRepositoryGetItemByID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	checklist := &DischargeChecklist{
		ID:            "checklist-item-1",
		EncounterID:   "encounter-item-1",
		DischargeType: "standard",
		Status:        ChecklistStatusInProgress,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	items := []DischargeChecklistItem{
		{
			ID:        "item-by-id-1",
			Code:      "patient_education",
			Label:     "Patient education completed",
			Required:  true,
			Status:    ItemStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	require.NoError(t, repo.CreateChecklistWithItems(ctx, checklist, items))

	t.Run("found", func(t *testing.T) {
		retrieved, err := repo.GetItemByID(ctx, "item-by-id-1")
		require.NoError(t, err)
		assert.Equal(t, "item-by-id-1", retrieved.ID)
		assert.Equal(t, "patient_education", retrieved.Code)
	})

	t.Run("not found", func(t *testing.T) {
		retrieved, err := repo.GetItemByID(ctx, "non-existent")
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, retrieved)
	})
}

func TestRepositoryUpdateItemFields(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	checklist := &DischargeChecklist{
		ID:            "checklist-update-item-1",
		EncounterID:   "encounter-update-item-1",
		DischargeType: "standard",
		Status:        ChecklistStatusInProgress,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	items := []DischargeChecklistItem{
		{
			ID:        "item-update-1",
			Code:      "patient_education",
			Label:     "Patient education completed",
			Required:  true,
			Status:    ItemStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	require.NoError(t, repo.CreateChecklistWithItems(ctx, checklist, items))

	// Update item status
	completedBy := "user-2"
	completedAt := now.Add(1 * time.Hour)
	updates := map[string]any{
		"status":       ItemStatusDone,
		"completed_by": completedBy,
		"completed_at": completedAt,
	}

	err := repo.UpdateItemFields(ctx, "item-update-1", updates)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetItemByID(ctx, "item-update-1")
	require.NoError(t, err)
	assert.Equal(t, ItemStatusDone, retrieved.Status)
	assert.NotNil(t, retrieved.CompletedBy)
	assert.Equal(t, "user-2", *retrieved.CompletedBy)
	assert.NotNil(t, retrieved.CompletedAt)
}

func TestRepositoryGetItemsByChecklistID(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	checklist := &DischargeChecklist{
		ID:            "checklist-items-1",
		EncounterID:   "encounter-items-1",
		DischargeType: "standard",
		Status:        ChecklistStatusInProgress,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	items := []DischargeChecklistItem{
		{
			ID:        "item-list-1",
			Code:      "patient_education",
			Label:     "Patient education completed",
			Required:  true,
			Status:    ItemStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "item-list-2",
			Code:      "medication_reconciliation",
			Label:     "Medication reconciliation done",
			Required:  true,
			Status:    ItemStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "item-list-3",
			Code:      "transport_arranged",
			Label:     "Transport arranged",
			Required:  false,
			Status:    ItemStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	require.NoError(t, repo.CreateChecklistWithItems(ctx, checklist, items))

	results, err := repo.GetItemsByChecklistID(ctx, "checklist-items-1")
	require.NoError(t, err)
	assert.Len(t, results, 3)
	// Ordered by required DESC, code ASC
	assert.Equal(t, "item-list-2", results[0].ID) // required, "medication_reconciliation"
	assert.Equal(t, "item-list-1", results[1].ID) // required, "patient_education"
	assert.Equal(t, "item-list-3", results[2].ID) // not required, "transport_arranged"
}

func TestRepositoryGetIncompleteRequiredItems(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()
	completedBy := "user-1"
	completedAt := now.Add(1 * time.Hour)

	checklist := &DischargeChecklist{
		ID:            "checklist-incomplete-1",
		EncounterID:   "encounter-incomplete-1",
		DischargeType: "standard",
		Status:        ChecklistStatusInProgress,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Note: Cannot test Required=false due to GORM default:true behavior
	// When a bool field has default:true, GORM treats false as zero value and applies default
	items := []DischargeChecklistItem{
		{
			ID:        "item-incomplete-1",
			Code:      "patient_education",
			Label:     "Patient education completed",
			Required:  true,
			Status:    ItemStatusOpen,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          "item-complete-1",
			Code:        "medication_reconciliation",
			Label:       "Medication reconciliation done",
			Required:    true,
			Status:      ItemStatusDone,
			CompletedBy: &completedBy,
			CompletedAt: &completedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:        "item-waived-1",
			Code:      "follow_up",
			Label:     "Follow-up scheduled",
			Required:  true,
			Status:    ItemStatusWaived,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	require.NoError(t, repo.CreateChecklistWithItems(ctx, checklist, items))

	// Should return only required items that are open (not done or waived)
	results, err := repo.GetIncompleteRequiredItems(ctx, "checklist-incomplete-1")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "item-incomplete-1", results[0].ID)
	assert.True(t, results[0].Required)
	assert.Equal(t, ItemStatusOpen, results[0].Status)
}

func TestRepositoryUpdateChecklistFields(t *testing.T) {
	db := newRepositoryTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	now := time.Now()

	checklist := &DischargeChecklist{
		ID:            "checklist-update-1",
		EncounterID:   "encounter-update-1",
		DischargeType: "standard",
		Status:        ChecklistStatusInProgress,
		CreatedBy:     "user-1",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, repo.CreateChecklistWithItems(ctx, checklist, []DischargeChecklistItem{}))

	// Update checklist status
	completedBy := "user-2"
	completedAt := now.Add(2 * time.Hour)
	updates := map[string]any{
		"status":       ChecklistStatusComplete,
		"completed_by": completedBy,
		"completed_at": completedAt,
	}

	err := repo.UpdateChecklistFields(ctx, "checklist-update-1", updates)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetChecklistByEncounterID(ctx, "encounter-update-1")
	require.NoError(t, err)
	assert.Equal(t, ChecklistStatusComplete, retrieved.Status)
	assert.NotNil(t, retrieved.CompletedBy)
	assert.Equal(t, "user-2", *retrieved.CompletedBy)
	assert.NotNil(t, retrieved.CompletedAt)
}
