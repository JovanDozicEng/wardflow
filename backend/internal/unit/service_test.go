package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepo is a mock implementation of Repository for testing
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) List(ctx context.Context, q, departmentID string) ([]Unit, error) {
	args := m.Called(ctx, q, departmentID)
	return args.Get(0).([]Unit), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, unit *Unit) error {
	args := m.Called(ctx, unit)
	return args.Error(0)
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Unit, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Unit), args.Error(1)
}

func TestService_List(t *testing.T) {
	t.Run("success - returns units from repository", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedUnits := []Unit{
			{ID: "1", Name: "ICU", Code: "ICU", DepartmentID: "dept-1"},
			{ID: "2", Name: "Emergency", Code: "ED", DepartmentID: "dept-1"},
		}

		repo.On("List", mock.Anything, "test-query", "dept-1").Return(expectedUnits, nil)

		units, err := svc.List(context.Background(), "test-query", "dept-1")

		assert.NoError(t, err)
		assert.Len(t, units, 2)
		assert.Equal(t, "ICU", units[0].Name)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedErr := errors.New("database error")
		repo.On("List", mock.Anything, "", "").Return([]Unit{}, expectedErr)

		units, err := svc.List(context.Background(), "", "")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, units)
		repo.AssertExpectations(t)
	})
}

func TestService_Create(t *testing.T) {
	t.Run("success - creates unit with valid data", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateUnitRequest{
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		repo.On("Create", mock.Anything, mock.MatchedBy(func(u *Unit) bool {
			return u.Name == "ICU" && u.Code == "ICU" && u.DepartmentID == "dept-1"
		})).Return(nil)

		unit, err := svc.Create(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, unit)
		assert.Equal(t, "ICU", unit.Name)
		assert.Equal(t, "ICU", unit.Code)
		assert.Equal(t, "dept-1", unit.DepartmentID)
		repo.AssertExpectations(t)
	})

	t.Run("error - name is required", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateUnitRequest{
			Name:         "",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		unit, err := svc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, unit)
		assert.Contains(t, err.Error(), "name is required")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - code is required", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateUnitRequest{
			Name:         "ICU",
			Code:         "",
			DepartmentID: "dept-1",
		}

		unit, err := svc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, unit)
		assert.Contains(t, err.Error(), "code is required")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - departmentId is required", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateUnitRequest{
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "",
		}

		unit, err := svc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, unit)
		assert.Contains(t, err.Error(), "departmentId is required")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateUnitRequest{
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		expectedErr := errors.New("database error")
		repo.On("Create", mock.Anything, mock.Anything).Return(expectedErr)

		unit, err := svc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, unit)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("success - returns unit by ID", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedUnit := &Unit{
			ID:           "unit-1",
			Name:         "ICU",
			Code:         "ICU",
			DepartmentID: "dept-1",
		}

		repo.On("GetByID", mock.Anything, "unit-1").Return(expectedUnit, nil)

		unit, err := svc.GetByID(context.Background(), "unit-1")

		assert.NoError(t, err)
		assert.NotNil(t, unit)
		assert.Equal(t, "unit-1", unit.ID)
		assert.Equal(t, "ICU", unit.Name)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedErr := errors.New("not found")
		repo.On("GetByID", mock.Anything, "unit-999").Return(nil, expectedErr)

		unit, err := svc.GetByID(context.Background(), "unit-999")

		assert.Error(t, err)
		assert.Nil(t, unit)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}
