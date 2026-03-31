package discharge

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
)

// Mock service
type mockService struct {
	mock.Mock
}

func (m *mockService) InitChecklist(ctx context.Context, encounterID, userID string, req InitChecklistRequest) (*DischargeChecklist, error) {
	args := m.Called(ctx, encounterID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DischargeChecklist), args.Error(1)
}

func (m *mockService) GetChecklist(ctx context.Context, encounterID string) (*DischargeChecklist, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DischargeChecklist), args.Error(1)
}

func (m *mockService) CompleteItem(ctx context.Context, itemID, userID string) (*DischargeChecklistItem, error) {
	args := m.Called(ctx, itemID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DischargeChecklistItem), args.Error(1)
}

func (m *mockService) CompleteDischarge(ctx context.Context, encounterID string, req CompleteDischargeRequest, userID string, userRole models.Role) (*DischargeChecklist, error) {
	args := m.Called(ctx, encounterID, req, userID, userRole)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*DischargeChecklist), args.Error(1)
}

func TestHandler_InitChecklist(t *testing.T) {
	t.Run("creates checklist successfully", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		userID := "user-001"

		reqBody := InitChecklistRequest{DischargeType: "standard"}
		checklist := &DischargeChecklist{
			ID:            "checklist-123",
			EncounterID:   encounterID,
			DischargeType: "standard",
			Status:        ChecklistStatusInProgress,
			CreatedBy:     userID,
		}

		svc.On("InitChecklist", mock.Anything, encounterID, userID, reqBody).Return(checklist, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/"+encounterID+"/discharge-checklist/init", reqBody, userID, models.RoleNurse)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.InitChecklist(rr, r)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var result DischargeChecklist
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, "checklist-123", result.ID)
		assert.Equal(t, encounterID, result.EncounterID)

		svc.AssertExpectations(t)
	})

	t.Run("returns conflict when checklist already exists", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		userID := "user-001"

		reqBody := InitChecklistRequest{DischargeType: "standard"}
		svc.On("InitChecklist", mock.Anything, encounterID, userID, reqBody).Return(nil, ErrChecklistAlreadyExists)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/"+encounterID+"/discharge-checklist/init", reqBody, userID, models.RoleNurse)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.InitChecklist(rr, r)

		assert.Equal(t, http.StatusConflict, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		handler := NewHandler(nil, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/enc-001/discharge-checklist/init", nil, "user-001", models.RoleNurse)
		r.SetPathValue("encounterId", "enc-001")
		r.Body = http.NoBody // Force decode error by not providing valid JSON
		rr := httptest.NewRecorder()

		handler.InitChecklist(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestHandler_GetChecklist(t *testing.T) {
	t.Run("returns checklist with items", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		checklist := &DischargeChecklist{
			ID:          "checklist-123",
			EncounterID: encounterID,
			Status:      ChecklistStatusInProgress,
			Items: []DischargeChecklistItem{
				{ID: "item-1", Status: ItemStatusOpen},
				{ID: "item-2", Status: ItemStatusDone},
			},
		}

		svc.On("GetChecklist", mock.Anything, encounterID).Return(checklist, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/"+encounterID+"/discharge-checklist", nil, "user-001", models.RoleNurse)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.GetChecklist(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result DischargeChecklist
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, "checklist-123", result.ID)
		assert.Len(t, result.Items, 2)

		svc.AssertExpectations(t)
	})

	t.Run("returns not found when checklist doesn't exist", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		svc.On("GetChecklist", mock.Anything, encounterID).Return(nil, ErrNotFound)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/"+encounterID+"/discharge-checklist", nil, "user-001", models.RoleNurse)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.GetChecklist(rr, r)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns internal error on service failure", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		svc.On("GetChecklist", mock.Anything, encounterID).Return(nil, errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/"+encounterID+"/discharge-checklist", nil, "user-001", models.RoleNurse)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.GetChecklist(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestHandler_CompleteItem(t *testing.T) {
	t.Run("completes item successfully", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		itemID := "item-001"
		userID := "user-001"

		completedItem := &DischargeChecklistItem{
			ID:          itemID,
			Status:      ItemStatusDone,
			CompletedBy: &userID,
		}

		svc.On("CompleteItem", mock.Anything, itemID, userID).Return(completedItem, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/discharge-checklist/items/"+itemID+"/complete", nil, userID, models.RoleNurse)
		r.SetPathValue("itemId", itemID)
		rr := httptest.NewRecorder()

		handler.CompleteItem(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result DischargeChecklistItem
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, ItemStatusDone, result.Status)

		svc.AssertExpectations(t)
	})

	t.Run("returns not found when item doesn't exist", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		itemID := "item-001"
		userID := "user-001"

		svc.On("CompleteItem", mock.Anything, itemID, userID).Return(nil, ErrNotFound)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/discharge-checklist/items/"+itemID+"/complete", nil, userID, models.RoleNurse)
		r.SetPathValue("itemId", itemID)
		rr := httptest.NewRecorder()

		handler.CompleteItem(rr, r)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request when item already completed", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		itemID := "item-001"
		userID := "user-001"

		svc.On("CompleteItem", mock.Anything, itemID, userID).Return(nil, ErrItemAlreadyCompleted)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/discharge-checklist/items/"+itemID+"/complete", nil, userID, models.RoleNurse)
		r.SetPathValue("itemId", itemID)
		rr := httptest.NewRecorder()

		handler.CompleteItem(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestHandler_CompleteDischarge(t *testing.T) {
	t.Run("completes discharge successfully", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		userID := "user-001"

		reqBody := CompleteDischargeRequest{Override: false}
		checklist := &DischargeChecklist{
			ID:     "checklist-123",
			Status: ChecklistStatusComplete,
		}

		svc.On("CompleteDischarge", mock.Anything, encounterID, reqBody, userID, models.RoleProvider).Return(checklist, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/"+encounterID+"/discharge/complete", reqBody, userID, models.RoleProvider)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.CompleteDischarge(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result DischargeChecklist
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, ChecklistStatusComplete, result.Status)

		svc.AssertExpectations(t)
	})

	t.Run("returns bad request for incomplete checklist", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		userID := "user-001"

		reqBody := CompleteDischargeRequest{Override: false}
		svc.On("CompleteDischarge", mock.Anything, encounterID, reqBody, userID, models.RoleProvider).Return(nil, ErrIncompleteChecklist)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/"+encounterID+"/discharge/complete", reqBody, userID, models.RoleProvider)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.CompleteDischarge(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns forbidden when non-admin tries to override", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		userID := "user-001"
		reason := "some reason"

		reqBody := CompleteDischargeRequest{Override: true, Reason: &reason}
		svc.On("CompleteDischarge", mock.Anything, encounterID, reqBody, userID, models.RoleNurse).Return(nil, ErrOverrideForbidden)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/"+encounterID+"/discharge/complete", reqBody, userID, models.RoleNurse)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.CompleteDischarge(rr, r)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request when discharge already done", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		userID := "user-001"

		reqBody := CompleteDischargeRequest{Override: false}
		svc.On("CompleteDischarge", mock.Anything, encounterID, reqBody, userID, models.RoleProvider).Return(nil, ErrDischargeAlreadyDone)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/"+encounterID+"/discharge/complete", reqBody, userID, models.RoleProvider)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.CompleteDischarge(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request when override reason is missing", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		userID := "user-001"

		reqBody := CompleteDischargeRequest{Override: true, Reason: nil}
		svc.On("CompleteDischarge", mock.Anything, encounterID, reqBody, userID, models.RoleAdmin).Return(nil, ErrOverrideReasonRequired)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/"+encounterID+"/discharge/complete", reqBody, userID, models.RoleAdmin)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.CompleteDischarge(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("allows admin to override with reason", func(t *testing.T) {
		svc := new(mockService)
		handler := NewHandler(svc, nil)

		encounterID := "enc-001"
		userID := "user-001"
		reason := "Emergency override"

		reqBody := CompleteDischargeRequest{Override: true, Reason: &reason}
		checklist := &DischargeChecklist{
			ID:             "checklist-123",
			Status:         ChecklistStatusOverrideComplete,
			OverrideReason: &reason,
		}

		svc.On("CompleteDischarge", mock.Anything, encounterID, reqBody, userID, models.RoleAdmin).Return(checklist, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/encounters/"+encounterID+"/discharge/complete", reqBody, userID, models.RoleAdmin)
		r.SetPathValue("encounterId", encounterID)
		rr := httptest.NewRecorder()

		handler.CompleteDischarge(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result DischargeChecklist
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, ChecklistStatusOverrideComplete, result.Status)

		svc.AssertExpectations(t)
	})
}
