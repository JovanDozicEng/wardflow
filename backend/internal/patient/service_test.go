package patient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepo is a mock implementation of Repository for testing
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Create(ctx context.Context, p *Patient) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Patient, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Patient), args.Error(1)
}

func (m *mockRepo) List(ctx context.Context, f ListPatientsFilter) ([]*Patient, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Patient), args.Get(1).(int64), args.Error(2)
}

func TestService_Create(t *testing.T) {
	t.Run("success - creates patient with valid data", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := &CreatePatientRequest{
			FirstName: "John",
			LastName:  "Doe",
			MRN:       "MRN123456",
		}

		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *Patient) bool {
			return p.FirstName == "John" && p.LastName == "Doe" && p.MRN == "MRN123456" && p.CreatedBy == "user-1"
		})).Return(nil)

		patient, err := svc.Create(context.Background(), req, "user-1")

		assert.NoError(t, err)
		assert.NotNil(t, patient)
		assert.Equal(t, "John", patient.FirstName)
		assert.Equal(t, "Doe", patient.LastName)
		assert.Equal(t, "MRN123456", patient.MRN)
		assert.Equal(t, "user-1", patient.CreatedBy)
		repo.AssertExpectations(t)
	})

	t.Run("success - creates patient with date of birth", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		dobStr := "1990-05-15"
		req := &CreatePatientRequest{
			FirstName:   "Jane",
			LastName:    "Smith",
			MRN:         "MRN789012",
			DateOfBirth: &dobStr,
		}

		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *Patient) bool {
			return p.DateOfBirth != nil && p.DateOfBirth.Format("2006-01-02") == dobStr
		})).Return(nil)

		patient, err := svc.Create(context.Background(), req, "user-2")

		assert.NoError(t, err)
		assert.NotNil(t, patient)
		assert.NotNil(t, patient.DateOfBirth)
		assert.Equal(t, "1990-05-15", patient.DateOfBirth.Format("2006-01-02"))
		repo.AssertExpectations(t)
	})

	t.Run("error - firstName is required", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := &CreatePatientRequest{
			FirstName: "",
			LastName:  "Doe",
			MRN:       "MRN123456",
		}

		patient, err := svc.Create(context.Background(), req, "user-1")

		assert.Error(t, err)
		assert.Nil(t, patient)
		assert.Contains(t, err.Error(), "firstName is required")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - lastName is required", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := &CreatePatientRequest{
			FirstName: "John",
			LastName:  "",
			MRN:       "MRN123456",
		}

		patient, err := svc.Create(context.Background(), req, "user-1")

		assert.Error(t, err)
		assert.Nil(t, patient)
		assert.Contains(t, err.Error(), "lastName is required")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - mrn is required", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := &CreatePatientRequest{
			FirstName: "John",
			LastName:  "Doe",
			MRN:       "",
		}

		patient, err := svc.Create(context.Background(), req, "user-1")

		assert.Error(t, err)
		assert.Nil(t, patient)
		assert.Contains(t, err.Error(), "mrn is required")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - invalid date of birth format", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		dobStr := "15/05/1990" // wrong format
		req := &CreatePatientRequest{
			FirstName:   "John",
			LastName:    "Doe",
			MRN:         "MRN123456",
			DateOfBirth: &dobStr,
		}

		patient, err := svc.Create(context.Background(), req, "user-1")

		assert.Error(t, err)
		assert.Nil(t, patient)
		assert.Contains(t, err.Error(), "dateOfBirth must be in ISO format")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := &CreatePatientRequest{
			FirstName: "John",
			LastName:  "Doe",
			MRN:       "MRN123456",
		}

		expectedErr := errors.New("database error")
		repo.On("Create", mock.Anything, mock.Anything).Return(expectedErr)

		patient, err := svc.Create(context.Background(), req, "user-1")

		assert.Error(t, err)
		assert.Nil(t, patient)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("success - returns patient by ID", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		expectedPatient := &Patient{
			ID:          "patient-1",
			FirstName:   "John",
			LastName:    "Doe",
			MRN:         "MRN123456",
			DateOfBirth: &dob,
			CreatedBy:   "user-1",
		}

		repo.On("GetByID", mock.Anything, "patient-1").Return(expectedPatient, nil)

		patient, err := svc.GetByID(context.Background(), "patient-1")

		assert.NoError(t, err)
		assert.NotNil(t, patient)
		assert.Equal(t, "patient-1", patient.ID)
		assert.Equal(t, "John", patient.FirstName)
		assert.Equal(t, "Doe", patient.LastName)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedErr := ErrNotFound
		repo.On("GetByID", mock.Anything, "patient-999").Return(nil, expectedErr)

		patient, err := svc.GetByID(context.Background(), "patient-999")

		assert.Error(t, err)
		assert.Nil(t, patient)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestService_List(t *testing.T) {
	t.Run("success - returns list of patients", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedPatients := []*Patient{
			{ID: "1", FirstName: "John", LastName: "Doe", MRN: "MRN001"},
			{ID: "2", FirstName: "Jane", LastName: "Smith", MRN: "MRN002"},
		}

		filter := ListPatientsFilter{
			Q:      "John",
			Limit:  20,
			Offset: 0,
		}

		repo.On("List", mock.Anything, filter).Return(expectedPatients, int64(2), nil)

		patients, total, err := svc.List(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, patients, 2)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, "John", patients[0].FirstName)
		repo.AssertExpectations(t)
	})

	t.Run("success - returns empty list when no matches", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		filter := ListPatientsFilter{
			Q:      "nonexistent",
			Limit:  20,
			Offset: 0,
		}

		repo.On("List", mock.Anything, filter).Return([]*Patient{}, int64(0), nil)

		patients, total, err := svc.List(context.Background(), filter)

		assert.NoError(t, err)
		assert.Empty(t, patients)
		assert.Equal(t, int64(0), total)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		filter := ListPatientsFilter{
			Limit:  20,
			Offset: 0,
		}

		expectedErr := errors.New("database error")
		repo.On("List", mock.Anything, filter).Return(nil, int64(0), expectedErr)

		patients, total, err := svc.List(context.Background(), filter)

		assert.Error(t, err)
		assert.Nil(t, patients)
		assert.Equal(t, int64(0), total)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}
