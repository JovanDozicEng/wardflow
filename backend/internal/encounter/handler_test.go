package encounter

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
	"github.com/wardflow/backend/pkg/auth"
)

// mockService is a mock implementation of Service
type mockService struct {
	mock.Mock
}

func (m *mockService) Create(ctx context.Context, req *CreateEncounterRequest, byUserID string) (*Encounter, error) {
	args := m.Called(ctx, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Encounter), args.Error(1)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*Encounter, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Encounter), args.Error(1)
}

func (m *mockService) List(ctx context.Context, f ListEncountersFilter) ([]*Encounter, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Encounter), args.Get(1).(int64), args.Error(2)
}

func (m *mockService) Update(ctx context.Context, id string, req *UpdateEncounterRequest, byUserID string) (*Encounter, error) {
	args := m.Called(ctx, id, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Encounter), args.Error(1)
}

func TestHandler_Create_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateEncounterRequest{
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
	}

	expectedEnc := &Encounter{
		ID:           "enc-1",
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
		Status:       EncounterStatusActive,
		CreatedBy:    "user-1",
	}

	svc.On("Create", mock.Anything, mock.AnythingOfType("*encounter.CreateEncounterRequest"), "user-1").
		Return(expectedEnc, nil)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/encounters", reqBody, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var result Encounter
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, "enc-1", result.ID)
	assert.Equal(t, "patient-1", result.PatientID)
	svc.AssertExpectations(t)
}

func TestHandler_Create_NonAdminWithAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateEncounterRequest{
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
	}

	expectedEnc := &Encounter{
		ID:        "enc-1",
		PatientID: "patient-1",
		UnitID:    "unit-1",
		CreatedBy: "user-1",
	}

	svc.On("Create", mock.Anything, mock.AnythingOfType("*encounter.CreateEncounterRequest"), "user-1").
		Return(expectedEnc, nil)

	req := testutil.NewRequest(http.MethodPost, "/api/v1/encounters", reqBody, "user-1", models.RoleNurse)
	// Manually set UnitIDs in context
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Create_NonAdminWithoutAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateEncounterRequest{
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
	}

	req := testutil.NewRequest(http.MethodPost, "/api/v1/encounters", reqBody, "user-1", models.RoleNurse)
	// Manually set UnitIDs in context
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-2", "unit-3"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/encounters", "invalid-json")
	req = req.WithContext(auth.SetUserContext(req.Context(), &auth.Claims{
		UserID: "user-1",
		Role:   models.RoleAdmin,
	}))
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Create_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	reqBody := CreateEncounterRequest{
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
	}

	svc.On("Create", mock.Anything, mock.AnythingOfType("*encounter.CreateEncounterRequest"), "user-1").
		Return(nil, errors.New("validation error"))

	req := testutil.NewRequest(http.MethodPost, "/api/v1/encounters", reqBody, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetByID_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	expectedEnc := &Encounter{
		ID:        "enc-1",
		PatientID: "patient-1",
		UnitID:    "unit-1",
		Status:    EncounterStatusActive,
	}

	svc.On("GetByID", mock.Anything, "enc-1").Return(expectedEnc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/enc-1", nil, "user-1", models.RoleAdmin)
	req.SetPathValue("encounterId", "enc-1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result Encounter
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, "enc-1", result.ID)
	svc.AssertExpectations(t)
}

func TestHandler_GetByID_NonAdminWithAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	expectedEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-1",
	}

	svc.On("GetByID", mock.Anything, "enc-1").Return(expectedEnc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/enc-1", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	req.SetPathValue("encounterId", "enc-1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetByID_NonAdminWithoutAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	expectedEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-1",
	}

	svc.On("GetByID", mock.Anything, "enc-1").Return(expectedEnc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/enc-1", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-2", "unit-3"}
	})
	req = req.WithContext(ctx)
	req.SetPathValue("encounterId", "enc-1")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetByID", mock.Anything, "enc-999").Return(nil, ErrNotFound)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/enc-999", nil, "user-1", models.RoleAdmin)
	req.SetPathValue("encounterId", "enc-999")
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_GetByID_MissingID(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters/", nil, "user-1", models.RoleAdmin)
	// Do not set PathValue
	rr := httptest.NewRecorder()

	handler.GetByID(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_List_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	encounters := []*Encounter{
		{ID: "enc-1", UnitID: "unit-1"},
		{ID: "enc-2", UnitID: "unit-1"},
	}

	svc.On("List", mock.Anything, mock.MatchedBy(func(f ListEncountersFilter) bool {
		return f.UnitID == "unit-1" && f.Limit == 20
	})).Return(encounters, int64(2), nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters?unitId=unit-1", nil, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result models.PaginatedResponse
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, int64(2), result.Total)
	svc.AssertExpectations(t)
}

func TestHandler_List_NonAdminWithAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	encounters := []*Encounter{
		{ID: "enc-1", UnitID: "unit-1"},
	}

	svc.On("List", mock.Anything, mock.AnythingOfType("encounter.ListEncountersFilter")).
		Return(encounters, int64(1), nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters?unitId=unit-1", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_List_NonAdminWithoutAccess(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters?unitId=unit-1", nil, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-2", "unit-3"}
	})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestHandler_List_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("List", mock.Anything, mock.AnythingOfType("encounter.ListEncountersFilter")).
		Return(nil, int64(0), errors.New("database error"))

	req := testutil.NewRequest(http.MethodGet, "/api/v1/encounters", nil, "user-1", models.RoleAdmin)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Update_Success(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	existingEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-1",
		Status: EncounterStatusActive,
	}

	newStatus := EncounterStatusDischarged
	reqBody := UpdateEncounterRequest{
		Status: &newStatus,
	}

	updatedEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-1",
		Status: EncounterStatusDischarged,
	}

	svc.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)
	svc.On("Update", mock.Anything, "enc-1", mock.AnythingOfType("*encounter.UpdateEncounterRequest"), "user-1").
		Return(updatedEnc, nil)

	req := testutil.NewRequest(http.MethodPatch, "/api/v1/encounters/enc-1", reqBody, "user-1", models.RoleAdmin)
	req.SetPathValue("encounterId", "enc-1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result Encounter
	testutil.DecodeJSON(t, rr, &result)
	assert.Equal(t, EncounterStatusDischarged, result.Status)
	svc.AssertExpectations(t)
}

func TestHandler_Update_NonAdminChangingUnit(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	existingEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-1",
		Status: EncounterStatusActive,
	}

	newUnitID := "unit-2"
	reqBody := UpdateEncounterRequest{
		UnitID: &newUnitID,
	}

	updatedEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-2",
		Status: EncounterStatusActive,
	}

	svc.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)
	svc.On("Update", mock.Anything, "enc-1", mock.AnythingOfType("*encounter.UpdateEncounterRequest"), "user-1").
		Return(updatedEnc, nil)

	req := testutil.NewRequest(http.MethodPatch, "/api/v1/encounters/enc-1", reqBody, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	req.SetPathValue("encounterId", "enc-1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Update_NonAdminChangingToInaccessibleUnit(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	existingEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-1",
	}

	newUnitID := "unit-3"
	reqBody := UpdateEncounterRequest{
		UnitID: &newUnitID,
	}

	svc.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)

	req := testutil.NewRequest(http.MethodPatch, "/api/v1/encounters/enc-1", reqBody, "user-1", models.RoleNurse)
	ctx := testutil.WithUser(req.Context(), "user-1", models.RoleNurse, func(claims *auth.Claims) {
		claims.UnitIDs = []string{"unit-1", "unit-2"}
	})
	req = req.WithContext(ctx)
	req.SetPathValue("encounterId", "enc-1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Update_NotFound(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	svc.On("GetByID", mock.Anything, "enc-999").Return(nil, ErrNotFound)

	req := testutil.NewRequest(http.MethodPatch, "/api/v1/encounters/enc-999", UpdateEncounterRequest{}, "user-1", models.RoleAdmin)
	req.SetPathValue("encounterId", "enc-999")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Update_InvalidJSON(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	existingEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-1",
	}

	svc.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)

	req := testutil.NewRequestNoAuth(http.MethodPatch, "/api/v1/encounters/enc-1", "invalid-json")
	req = req.WithContext(auth.SetUserContext(req.Context(), &auth.Claims{
		UserID:  "user-1",
		Role:    models.RoleAdmin,
		UnitIDs: []string{"unit-1"},
	}))
	req.SetPathValue("encounterId", "enc-1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}

func TestHandler_Update_ServiceError(t *testing.T) {
	svc := new(mockService)
	handler := NewHandler(svc, nil)

	existingEnc := &Encounter{
		ID:     "enc-1",
		UnitID: "unit-1",
		Status: EncounterStatusActive,
	}

	newStatus := EncounterStatusActive
	reqBody := UpdateEncounterRequest{
		Status: &newStatus,
	}

	svc.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)
	svc.On("Update", mock.Anything, "enc-1", mock.AnythingOfType("*encounter.UpdateEncounterRequest"), "user-1").
		Return(nil, errors.New("update failed"))

	req := testutil.NewRequest(http.MethodPatch, "/api/v1/encounters/enc-1", reqBody, "user-1", models.RoleAdmin)
	req.SetPathValue("encounterId", "enc-1")
	rr := httptest.NewRecorder()

	handler.Update(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	svc.AssertExpectations(t)
}
