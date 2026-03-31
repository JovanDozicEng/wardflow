package department

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

func (m *mockService) List(ctx context.Context, q string) ([]Department, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]Department), args.Error(1)
}

func (m *mockService) Create(ctx context.Context, req CreateDepartmentRequest) (*Department, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Department), args.Error(1)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*Department, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Department), args.Error(1)
}

func TestHandler_List(t *testing.T) {
	t.Run("success - returns list of departments", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		expectedDepts := []Department{
			{ID: "1", Name: "Emergency", Code: "EMERGENCY"},
			{ID: "2", Name: "Cardiology", Code: "CARDIOLOGY"},
		}

		svc.On("List", mock.Anything, "").Return(expectedDepts, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/departments", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var depts []Department
		testutil.DecodeJSON(t, rr, &depts)
		assert.Len(t, depts, 2)
		assert.Equal(t, "Emergency", depts[0].Name)

		svc.AssertExpectations(t)
	})

	t.Run("success - filters by query parameter", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		expectedDepts := []Department{
			{ID: "1", Name: "Emergency", Code: "EMERGENCY"},
		}

		svc.On("List", mock.Anything, "Emergency").Return(expectedDepts, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/departments?q=Emergency", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var depts []Department
		testutil.DecodeJSON(t, rr, &depts)
		assert.Len(t, depts, 1)

		svc.AssertExpectations(t)
	})

	t.Run("error - service returns error", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		svc.On("List", mock.Anything, "").Return([]Department{}, errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/departments", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		r := testutil.NewRequestNoAuth(http.MethodGet, "/api/v1/departments", nil)
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.List(rr, r)
		})
	})
}

func TestHandler_Create(t *testing.T) {
	t.Run("success - creates department with admin role", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		req := CreateDepartmentRequest{
			Name: "Emergency",
			Code: "EMERGENCY",
		}

		expectedDept := &Department{
			ID:   "dept-123",
			Name: "Emergency",
			Code: "EMERGENCY",
		}

		svc.On("Create", mock.Anything, req).Return(expectedDept, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/departments", req, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var dept Department
		testutil.DecodeJSON(t, rr, &dept)
		assert.Equal(t, "dept-123", dept.ID)
		assert.Equal(t, "Emergency", dept.Name)

		svc.AssertExpectations(t)
	})

	t.Run("error - forbidden for non-admin role", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		req := CreateDepartmentRequest{
			Name: "Emergency",
			Code: "EMERGENCY",
		}

		r := testutil.NewRequest(http.MethodPost, "/api/v1/departments", req, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusForbidden, rr.Code)
		svc.AssertNotCalled(t, "Create")
	})

	t.Run("error - invalid JSON body", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		r := httptest.NewRequest(http.MethodPost, "/api/v1/departments", bytes.NewBufferString("invalid-json"))
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

		req := CreateDepartmentRequest{
			Name: "",
			Code: "EMERGENCY",
		}

		svc.On("Create", mock.Anything, req).Return(nil, errors.New("name is required"))

		r := testutil.NewRequest(http.MethodPost, "/api/v1/departments", req, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		req := CreateDepartmentRequest{
			Name: "Emergency",
			Code: "EMERGENCY",
		}

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/departments", req)
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.Create(rr, r)
		})
	})
}

func TestHandler_Get(t *testing.T) {
	t.Run("success - returns department by ID", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		expectedDept := &Department{
			ID:   "dept-123",
			Name: "Emergency",
			Code: "EMERGENCY",
		}

		svc.On("GetByID", mock.Anything, "dept-123").Return(expectedDept, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/departments/dept-123", nil, "user-1", models.RoleNurse)
		r.SetPathValue("departmentId", "dept-123")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var dept Department
		testutil.DecodeJSON(t, rr, &dept)
		assert.Equal(t, "dept-123", dept.ID)
		assert.Equal(t, "Emergency", dept.Name)

		svc.AssertExpectations(t)
	})

	t.Run("error - department not found", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		svc.On("GetByID", mock.Anything, "dept-999").Return(nil, gorm.ErrRecordNotFound)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/departments/dept-999", nil, "user-1", models.RoleNurse)
		r.SetPathValue("departmentId", "dept-999")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - empty department ID", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/departments/", nil, "user-1", models.RoleNurse)
		// Don't set path value - simulate empty ID
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertNotCalled(t, "GetByID")
	})

	t.Run("error - service returns internal error", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		svc.On("GetByID", mock.Anything, "dept-123").Return(nil, errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/departments/dept-123", nil, "user-1", models.RoleNurse)
		r.SetPathValue("departmentId", "dept-123")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc)

		r := testutil.NewRequestNoAuth(http.MethodGet, "/api/v1/departments/dept-123", nil)
		r.SetPathValue("departmentId", "dept-123")
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.Get(rr, r)
		})
	})
}
