package patient

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/internal/testutil"
	"github.com/wardflow/backend/pkg/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// mockService is a mock implementation of Service for testing
type mockService struct {
	mock.Mock
}

// mockDB creates a mock database.DB for testing (for audit logging)
func mockDB(t *testing.T) *database.DB {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	// Expect any audit log insert and return success
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"audit_logs\"").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: false,
	})
	if err != nil {
		t.Fatalf("failed to create gorm db: %v", err)
	}

	return &database.DB{DB: gormDB}
}

func (m *mockService) Create(ctx context.Context, req *CreatePatientRequest, byUserID string) (*Patient, error) {
	args := m.Called(ctx, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Patient), args.Error(1)
}

func (m *mockService) GetByID(ctx context.Context, id string) (*Patient, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Patient), args.Error(1)
}

func (m *mockService) List(ctx context.Context, f ListPatientsFilter) ([]*Patient, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Patient), args.Get(1).(int64), args.Error(2)
}

func TestHandler_List(t *testing.T) {
	t.Run("success - returns paginated list of patients", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil) // db is nil for tests

		expectedPatients := []*Patient{
			{ID: "1", FirstName: "John", LastName: "Doe", MRN: "MRN001"},
			{ID: "2", FirstName: "Jane", LastName: "Smith", MRN: "MRN002"},
		}

		filter := ListPatientsFilter{
			Q:      "",
			Limit:  20,
			Offset: 0,
		}

		svc.On("List", mock.Anything, filter).Return(expectedPatients, int64(2), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response models.PaginatedResponse
		testutil.DecodeJSON(t, rr, &response)
		assert.Equal(t, int64(2), response.Total)
		assert.Equal(t, 20, response.Limit)
		assert.Equal(t, 0, response.Offset)

		svc.AssertExpectations(t)
	})

	t.Run("success - filters by query parameter", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		expectedPatients := []*Patient{
			{ID: "1", FirstName: "John", LastName: "Doe", MRN: "MRN001"},
		}

		filter := ListPatientsFilter{
			Q:      "John",
			Limit:  20,
			Offset: 0,
		}

		svc.On("List", mock.Anything, filter).Return(expectedPatients, int64(1), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients?q=John", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success - custom limit and offset", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		filter := ListPatientsFilter{
			Q:      "",
			Limit:  10,
			Offset: 5,
		}

		svc.On("List", mock.Anything, filter).Return([]*Patient{}, int64(0), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients?limit=10&offset=5", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("success - invalid limit uses default", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		filter := ListPatientsFilter{
			Q:      "",
			Limit:  20, // default
			Offset: 0,
		}

		svc.On("List", mock.Anything, filter).Return([]*Patient{}, int64(0), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients?limit=invalid", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - service returns error", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		filter := ListPatientsFilter{
			Q:      "",
			Limit:  20,
			Offset: 0,
		}

		svc.On("List", mock.Anything, filter).Return(nil, int64(0), errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.List(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		r := testutil.NewRequestNoAuth(http.MethodGet, "/api/v1/patients", nil)
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.List(rr, r)
		})
	})
}

func TestHandler_Create(t *testing.T) {
	t.Run("success - creates patient", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, mockDB(t))

		req := &CreatePatientRequest{
			FirstName: "John",
			LastName:  "Doe",
			MRN:       "MRN123456",
		}

		expectedPatient := &Patient{
			ID:        "patient-123",
			FirstName: "John",
			LastName:  "Doe",
			MRN:       "MRN123456",
			CreatedBy: "user-1",
		}

		svc.On("Create", mock.Anything, req, "user-1").Return(expectedPatient, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/patients", req, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var patient Patient
		testutil.DecodeJSON(t, rr, &patient)
		assert.Equal(t, "patient-123", patient.ID)
		assert.Equal(t, "John", patient.FirstName)

		svc.AssertExpectations(t)
	})

	t.Run("success - creates patient with date of birth", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, mockDB(t))

		dobStr := "1990-05-15"
		req := &CreatePatientRequest{
			FirstName:   "Jane",
			LastName:    "Smith",
			MRN:         "MRN789012",
			DateOfBirth: &dobStr,
		}

		dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		expectedPatient := &Patient{
			ID:          "patient-456",
			FirstName:   "Jane",
			LastName:    "Smith",
			MRN:         "MRN789012",
			DateOfBirth: &dob,
			CreatedBy:   "user-2",
		}

		svc.On("Create", mock.Anything, req, "user-2").Return(expectedPatient, nil)

		r := testutil.NewRequest(http.MethodPost, "/api/v1/patients", req, "user-2", models.RoleProvider)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusCreated, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - invalid JSON body", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		r := httptest.NewRequest(http.MethodPost, "/api/v1/patients", bytes.NewBufferString("invalid-json"))
		r.Header.Set("Content-Type", "application/json")
		ctx := testutil.WithUser(r.Context(), "user-1", models.RoleNurse)
		r = r.WithContext(ctx)

		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertNotCalled(t, "Create")
	})

	t.Run("error - service validation error", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		req := &CreatePatientRequest{
			FirstName: "",
			LastName:  "Doe",
			MRN:       "MRN123456",
		}

		svc.On("Create", mock.Anything, req, "user-1").Return(nil, errors.New("firstName is required"))

		r := testutil.NewRequest(http.MethodPost, "/api/v1/patients", req, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		h.Create(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		req := &CreatePatientRequest{
			FirstName: "John",
			LastName:  "Doe",
			MRN:       "MRN123456",
		}

		r := testutil.NewRequestNoAuth(http.MethodPost, "/api/v1/patients", req)
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.Create(rr, r)
		})
	})
}

func TestHandler_Get(t *testing.T) {
	t.Run("success - returns patient by ID", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		expectedPatient := &Patient{
			ID:          "patient-123",
			FirstName:   "John",
			LastName:    "Doe",
			MRN:         "MRN123456",
			DateOfBirth: &dob,
			CreatedBy:   "user-1",
		}

		svc.On("GetByID", mock.Anything, "patient-123").Return(expectedPatient, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients/patient-123", nil, "user-1", models.RoleNurse)
		r.SetPathValue("patientId", "patient-123")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var patient Patient
		testutil.DecodeJSON(t, rr, &patient)
		assert.Equal(t, "patient-123", patient.ID)
		assert.Equal(t, "John", patient.FirstName)
		assert.Equal(t, "Doe", patient.LastName)

		svc.AssertExpectations(t)
	})

	t.Run("error - patient not found", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		svc.On("GetByID", mock.Anything, "patient-999").Return(nil, ErrNotFound)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients/patient-999", nil, "user-1", models.RoleNurse)
		r.SetPathValue("patientId", "patient-999")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - empty patient ID", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients/", nil, "user-1", models.RoleNurse)
		// Don't set path value - simulate empty ID
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertNotCalled(t, "GetByID")
	})

	t.Run("error - service returns internal error", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		svc.On("GetByID", mock.Anything, "patient-123").Return(nil, errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/patients/patient-123", nil, "user-1", models.RoleNurse)
		r.SetPathValue("patientId", "patient-123")
		rr := httptest.NewRecorder()

		h.Get(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("error - no auth context panics", func(t *testing.T) {
		svc := new(mockService)
		h := NewHandler(svc, nil)

		r := testutil.NewRequestNoAuth(http.MethodGet, "/api/v1/patients/patient-123", nil)
		r.SetPathValue("patientId", "patient-123")
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			h.Get(rr, r)
		})
	})
}
