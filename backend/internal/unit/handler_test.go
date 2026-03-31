package unit

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
	"gorm.io/gorm"
)

// mockService is a mock implementation of Service for testing
type mockService struct {
	mock.Mock
}

func (m *mockService) List(ctx context.Context, q, departmentID string) ([]Unit, error) {
	args := m.Called(ctx, q, departmentID)
	return args.Get(0).([]Unit), args.Error(1)
}

func (m *mockService) Create(ctx context.Context, req CreateUnitRequest) (*Unit, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Unit), args.Error(1)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*Unit, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Unit), args.Error(1)
}

func TestHandler_List(t *testing.T) {
	t.Run("success - returns list of units", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		expectedUnits := []Unit{
			{ID: "1", Name: "ICU", Code: "ICU", DepartmentID: "dept-1"},
			{ID: "2", Name: "Emergency", Code: "ED", DepartmentID: "dept-1"},
		}

		svc.On("List", mock.Anything, "", "").Return(expectedUnits, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/units", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var units []Unit
		testutil.DecodeJSON(t, rr, &units)
		assert.Len(t, units, 2)
		assert.Equal(t, "ICU", units[0].Name)

		svc.AssertExpectations(t)
	})

	t.Run("success - filters by query parameter", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		expectedUnits := []Unit{
			{ID: "1", Name: "ICU", Code: "ICU", DepartmentID: "dept-1"},
		}

		svc.On("List", mock.Anything, "ICU", "").Return(expectedUnits, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/units?q=ICU", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var units []Unit
		testutil.DecodeJSON(t, rr, &units)
		assert.Len(t, units, 1)

		svc.AssertExpectations(t)
	})

	t.Run("success - filters by departmentId", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		expectedUnits := []Unit{
			{ID: "1", Name: "ICU", Code: "ICU", DepartmentID: "dept-123"},
		}

		svc.On("List", mock.Anything, "", "dept-123").Return(expectedUnits, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/units?departmentId=dept-123", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - service returns error", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		svc.On("List", mock.Anything, "", "").Return([]Unit{}, errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/units", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		r := testutil.NewRequestNoAuth(http.MethodGet, "/api/v1/units", nil)
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.List(rr, r)
		})
	})
}

func TestHandler_Create(t *testing.T) {
	t.Run("success - creates unit with admin role", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		req := CreateUnitRequest{
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		expectedUnit := &Unit{
			ID:           "unit-123",
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		svc.On("Create", mock.Anything, req).Return(expectedUnit, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/units", req, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var unit Unit
		testutil.DecodeJSON(t, rr, &unit)
		assert.Equal(t, "unit-123", unit.ID)
		assert.Equal(t, "ICU", unit.Name)

		svc.AssertExpectations(t)
	})

	t.Run("error - forbidden for non-admin role", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		req := CreateUnitRequest{
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		r := testutil.NewRequest(http.MethodPost, "/api/v1/units", req, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		svc.AssertNotCalled(t, "Create")
	})

	t.Run("error - invalid JSON body", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		r := httptest.NewRequest(http.MethodPost, "/api/v1/units", bytes.NewBufferString("invalid-json"))
		r.Header.Set("Content-Type", "application/json")
		ctx := testutil.WithUser(r.Context(), "admin-1", models.RoleAdmin)
		r = r.WithContext(ctx)

		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertNotCalled(t, "Create")
	})

	t.Run("error - service validation error", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		req := CreateUnitRequest{
			Name:         "",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		svc.On("Create", mock.Anything, req).Return(nil, errors.New("name is required"))

		r := testutil.NewRequest(http.MethodPost, "/api/v1/units", req, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		req := CreateUnitRequest{
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/units", req)
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.Create(rr, r)
		})
	})
}

func TestHandler_Get(t *testing.T) {
	t.Run("success - returns unit by ID", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		expectedUnit := &Unit{
			ID:           "unit-123",
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		svc.On("GetByID", mock.Anything, "unit-123").Return(expectedUnit, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/units/unit-123", nil, "user-1", models.RoleNurse)
		r.SetPathValue("unitId", "unit-123")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var unit Unit
		testutil.DecodeJSON(t, rr, &unit)
		assert.Equal(t, "unit-123", unit.ID)
		assert.Equal(t, "ICU", unit.Name)

		svc.AssertExpectations(t)
	})

	t.Run("error - unit not found", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		svc.On("GetByID", mock.Anything, "unit-999").Return(nil, gorm.ErrRecordNotFound)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/units/unit-999", nil, "user-1", models.RoleNurse)
		r.SetPathValue("unitId", "unit-999")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - empty unit ID", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/units/", nil, "user-1", models.RoleNurse)
		// Don't set path value - simulate empty ID
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertNotCalled(t, "GetByID")
	})

	t.Run("error - service returns internal error", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		svc.On("GetByID", mock.Anything, "unit-123").Return(nil, errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/units/unit-123", nil, "user-1", models.RoleNurse)
		r.SetPathValue("unitId", "unit-123")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		r := testutil.NewRequestNoAuth(http.MethodGet, "/api/v1/units/unit-123", nil)
		r.SetPathValue("unitId", "unit-123")
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.Get(rr, r)
		})
	})
}
