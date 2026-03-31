package discharge

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
)

// Mock repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetChecklistByEncounterID(ctx context.Context, encounterID string) (*DischargeChecklist, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DischargeChecklist), args.Error(1)
}

func (m *mockRepository) CreateChecklistWithItems(ctx context.Context, checklist *DischargeChecklist, items []DischargeChecklistItem) error {
	args := m.Called(ctx, checklist, items)
	// Simulate ID generation
	if checklist.ID == "" {
		checklist.ID = "checklist-123"
	}
	return args.Error(0)
}

func (m *mockRepository) GetItemByID(ctx context.Context, id string) (*DischargeChecklistItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DischargeChecklistItem), args.Error(1)
}

func (m *mockRepository) UpdateItemFields(ctx context.Context, itemID string, updates map[string]any) error {
	args := m.Called(ctx, itemID, updates)
	return args.Error(0)
}

func (m *mockRepository) GetItemsByChecklistID(ctx context.Context, checklistID string) ([]DischargeChecklistItem, error) {
	args := m.Called(ctx, checklistID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]DischargeChecklistItem), args.Error(1)
}

func (m *mockRepository) GetIncompleteRequiredItems(ctx context.Context, checklistID string) ([]DischargeChecklistItem, error) {
	args := m.Called(ctx, checklistID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]DischargeChecklistItem), args.Error(1)
}

func (m *mockRepository) UpdateChecklistFields(ctx context.Context, checklistID string, updates map[string]any) error {
	args := m.Called(ctx, checklistID, updates)
	return args.Error(0)
}

// Tests
func TestService_InitChecklist(t *testing.T) {
	ctx := context.Background()
	encounterID := "enc-001"
	userID := "user-001"

	t.Run("creates checklist with standard items", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(nil, ErrNotFound)
		repo.On("CreateChecklistWithItems", ctx, mock.AnythingOfType("*discharge.DischargeChecklist"), mock.AnythingOfType("[]discharge.DischargeChecklistItem")).Return(nil)
		repo.On("GetItemsByChecklistID", ctx, "checklist-123").Return([]DischargeChecklistItem{
			{ID: "item-1", Code: "patient_education", Label: "Patient education completed", Required: true, Status: ItemStatusOpen},
		}, nil)

		req := InitChecklistRequest{DischargeType: "standard"}
		checklist, err := svc.InitChecklist(ctx, encounterID, userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, checklist)
		assert.Equal(t, encounterID, checklist.EncounterID)
		assert.Equal(t, "standard", checklist.DischargeType)
		assert.Equal(t, ChecklistStatusInProgress, checklist.Status)
		assert.Equal(t, userID, checklist.CreatedBy)
		assert.Len(t, checklist.Items, 1)

		repo.AssertExpectations(t)
	})

	t.Run("defaults to standard discharge type", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(nil, ErrNotFound)
		repo.On("CreateChecklistWithItems", ctx, mock.MatchedBy(func(cl *DischargeChecklist) bool {
			return cl.DischargeType == "standard"
		}), mock.AnythingOfType("[]discharge.DischargeChecklistItem")).Return(nil)
		repo.On("GetItemsByChecklistID", ctx, "checklist-123").Return([]DischargeChecklistItem{}, nil)

		req := InitChecklistRequest{} // empty discharge type
		_, err := svc.InitChecklist(ctx, encounterID, userID, req)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("returns error if checklist already exists", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		existing := &DischargeChecklist{ID: "existing-123", EncounterID: encounterID}
		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(existing, nil)

		req := InitChecklistRequest{DischargeType: "standard"}
		checklist, err := svc.InitChecklist(ctx, encounterID, userID, req)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrChecklistAlreadyExists)
		assert.Nil(t, checklist)
		repo.AssertExpectations(t)
	})

	t.Run("creates AMA checklist with AMA-specific items", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(nil, ErrNotFound)
		repo.On("CreateChecklistWithItems", ctx, mock.MatchedBy(func(cl *DischargeChecklist) bool {
			return cl.DischargeType == "ama"
		}), mock.MatchedBy(func(items []DischargeChecklistItem) bool {
			// AMA has 3 items
			return len(items) == 3
		})).Return(nil)
		repo.On("GetItemsByChecklistID", ctx, "checklist-123").Return([]DischargeChecklistItem{}, nil)

		req := InitChecklistRequest{DischargeType: "ama"}
		checklist, err := svc.InitChecklist(ctx, encounterID, userID, req)

		assert.NoError(t, err)
		assert.Equal(t, "ama", checklist.DischargeType)
		repo.AssertExpectations(t)
	})
}

func TestService_GetChecklist(t *testing.T) {
	ctx := context.Background()
	encounterID := "enc-001"

	t.Run("returns checklist with items", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		checklist := &DischargeChecklist{
			ID:          "checklist-123",
			EncounterID: encounterID,
			Status:      ChecklistStatusInProgress,
		}
		items := []DischargeChecklistItem{
			{ID: "item-1", ChecklistID: "checklist-123", Status: ItemStatusOpen},
			{ID: "item-2", ChecklistID: "checklist-123", Status: ItemStatusDone},
		}

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(checklist, nil)
		repo.On("GetItemsByChecklistID", ctx, "checklist-123").Return(items, nil)

		result, err := svc.GetChecklist(ctx, encounterID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "checklist-123", result.ID)
		assert.Len(t, result.Items, 2)
		repo.AssertExpectations(t)
	})

	t.Run("returns error if not found", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(nil, ErrNotFound)

		result, err := svc.GetChecklist(ctx, encounterID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})
}

func TestService_CompleteItem(t *testing.T) {
	ctx := context.Background()
	itemID := "item-001"
	userID := "user-001"

	t.Run("marks item as complete", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		item := &DischargeChecklistItem{
			ID:          itemID,
			ChecklistID: "checklist-123",
			Status:      ItemStatusOpen,
		}

		completedItem := &DischargeChecklistItem{
			ID:          itemID,
			ChecklistID: "checklist-123",
			Status:      ItemStatusDone,
			CompletedBy: &userID,
		}

		repo.On("GetItemByID", ctx, itemID).Return(item, nil).Once()
		repo.On("UpdateItemFields", ctx, itemID, mock.MatchedBy(func(updates map[string]any) bool {
			return updates["status"] == ItemStatusDone &&
				updates["completed_by"] == userID &&
				updates["completed_at"] != nil
		})).Return(nil)
		repo.On("GetItemByID", ctx, itemID).Return(completedItem, nil).Once()

		result, err := svc.CompleteItem(ctx, itemID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, ItemStatusDone, result.Status)
		repo.AssertExpectations(t)
	})

	t.Run("returns error if item already completed", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		completedItem := &DischargeChecklistItem{
			ID:     itemID,
			Status: ItemStatusDone,
		}

		repo.On("GetItemByID", ctx, itemID).Return(completedItem, nil)

		result, err := svc.CompleteItem(ctx, itemID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrItemAlreadyCompleted)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("returns error if item not found", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		repo.On("GetItemByID", ctx, itemID).Return(nil, ErrNotFound)

		result, err := svc.CompleteItem(ctx, itemID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})
}

func TestService_CompleteDischarge(t *testing.T) {
	ctx := context.Background()
	encounterID := "enc-001"
	userID := "user-001"

	t.Run("completes discharge when all required items are done", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		checklist := &DischargeChecklist{
			ID:          "checklist-123",
			EncounterID: encounterID,
			Status:      ChecklistStatusInProgress,
		}

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(checklist, nil).Twice()
		repo.On("GetIncompleteRequiredItems", ctx, "checklist-123").Return([]DischargeChecklistItem{}, nil)
		repo.On("UpdateChecklistFields", ctx, "checklist-123", mock.MatchedBy(func(updates map[string]any) bool {
			return updates["status"] == ChecklistStatusComplete &&
				updates["completed_by"] == userID &&
				updates["completed_at"] != nil
		})).Return(nil)
		repo.On("GetItemsByChecklistID", ctx, "checklist-123").Return([]DischargeChecklistItem{}, nil)

		req := CompleteDischargeRequest{Override: false}
		result, err := svc.CompleteDischarge(ctx, encounterID, req, userID, models.RoleProvider)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when required items incomplete without override", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		checklist := &DischargeChecklist{
			ID:     "checklist-123",
			Status: ChecklistStatusInProgress,
		}

		incompleteItems := []DischargeChecklistItem{
			{ID: "item-1", Required: true, Status: ItemStatusOpen},
		}

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(checklist, nil)
		repo.On("GetIncompleteRequiredItems", ctx, "checklist-123").Return(incompleteItems, nil)

		req := CompleteDischargeRequest{Override: false}
		result, err := svc.CompleteDischarge(ctx, encounterID, req, userID, models.RoleProvider)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrIncompleteChecklist)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("allows admin to override incomplete checklist", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		checklist := &DischargeChecklist{
			ID:     "checklist-123",
			Status: ChecklistStatusInProgress,
		}

		incompleteItems := []DischargeChecklistItem{
			{ID: "item-1", Required: true, Status: ItemStatusOpen},
		}

		reason := "Patient requested early discharge"

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(checklist, nil).Twice()
		repo.On("GetIncompleteRequiredItems", ctx, "checklist-123").Return(incompleteItems, nil)
		repo.On("UpdateChecklistFields", ctx, "checklist-123", mock.MatchedBy(func(updates map[string]any) bool {
			return updates["status"] == ChecklistStatusOverrideComplete &&
				updates["override_reason"] == reason
		})).Return(nil)
		repo.On("GetItemsByChecklistID", ctx, "checklist-123").Return([]DischargeChecklistItem{}, nil)

		req := CompleteDischargeRequest{Override: true, Reason: &reason}
		result, err := svc.CompleteDischarge(ctx, encounterID, req, userID, models.RoleAdmin)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("allows charge nurse to override incomplete checklist", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		checklist := &DischargeChecklist{
			ID:     "checklist-123",
			Status: ChecklistStatusInProgress,
		}

		incompleteItems := []DischargeChecklistItem{
			{ID: "item-1", Required: true, Status: ItemStatusOpen},
		}

		reason := "Emergency situation"

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(checklist, nil).Twice()
		repo.On("GetIncompleteRequiredItems", ctx, "checklist-123").Return(incompleteItems, nil)
		repo.On("UpdateChecklistFields", ctx, "checklist-123", mock.AnythingOfType("map[string]interface {}")).Return(nil)
		repo.On("GetItemsByChecklistID", ctx, "checklist-123").Return([]DischargeChecklistItem{}, nil)

		req := CompleteDischargeRequest{Override: true, Reason: &reason}
		result, err := svc.CompleteDischarge(ctx, encounterID, req, userID, models.RoleChargeNurse)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when non-admin tries to override", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		checklist := &DischargeChecklist{
			ID:     "checklist-123",
			Status: ChecklistStatusInProgress,
		}

		incompleteItems := []DischargeChecklistItem{
			{ID: "item-1", Required: true, Status: ItemStatusOpen},
		}

		reason := "some reason"

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(checklist, nil)
		repo.On("GetIncompleteRequiredItems", ctx, "checklist-123").Return(incompleteItems, nil)

		req := CompleteDischargeRequest{Override: true, Reason: &reason}
		result, err := svc.CompleteDischarge(ctx, encounterID, req, userID, models.RoleNurse)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrOverrideForbidden)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("returns error when override without reason", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		checklist := &DischargeChecklist{
			ID:     "checklist-123",
			Status: ChecklistStatusInProgress,
		}

		incompleteItems := []DischargeChecklistItem{
			{ID: "item-1", Required: true, Status: ItemStatusOpen},
		}

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(checklist, nil)
		repo.On("GetIncompleteRequiredItems", ctx, "checklist-123").Return(incompleteItems, nil)

		req := CompleteDischargeRequest{Override: true, Reason: nil}
		result, err := svc.CompleteDischarge(ctx, encounterID, req, userID, models.RoleAdmin)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrOverrideReasonRequired)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})

	t.Run("returns error if discharge already completed", func(t *testing.T) {
		repo := new(mockRepository)
		svc := NewService(repo)

		now := time.Now()
		checklist := &DischargeChecklist{
			ID:          "checklist-123",
			Status:      ChecklistStatusComplete,
			CompletedAt: &now,
		}

		repo.On("GetChecklistByEncounterID", ctx, encounterID).Return(checklist, nil)

		req := CompleteDischargeRequest{Override: false}
		result, err := svc.CompleteDischarge(ctx, encounterID, req, userID, models.RoleProvider)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrDischargeAlreadyDone)
		assert.Nil(t, result)
		repo.AssertExpectations(t)
	})
}
