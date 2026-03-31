package department

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

func (m *mockRepo) List(ctx context.Context, q string) ([]Department, error) {
	args := m.Called(ctx, q)
	return args.Get(0).([]Department), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, dept *Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (*Department, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Department), args.Error(1)
}

func TestService_List(t *testing.T) {
	t.Run("success - returns departments from repository", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedDepts := []Department{
			{ID: "1", Name: "Emergency", Code: "EMERGENCY"},
			{ID: "2", Name: "Cardiology", Code: "CARDIOLOGY"},
		}

		repo.On("List", mock.Anything, "test-query").Return(expectedDepts, nil)

		depts, err := svc.List(context.Background(), "test-query")

		assert.NoError(t, err)
		assert.Len(t, depts, 2)
		assert.Equal(t, "Emergency", depts[0].Name)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedErr := errors.New("database error")
		repo.On("List", mock.Anything, "").Return([]Department{}, expectedErr)

		depts, err := svc.List(context.Background(), "")

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, depts)
		repo.AssertExpectations(t)
	})
}

func TestService_Create(t *testing.T) {
	t.Run("success - creates department with valid data", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateDepartmentRequest{
			Name: "Emergency",
			Code: "EMERGENCY",
		}

		repo.On("Create", mock.Anything, mock.MatchedBy(func(d *Department) bool {
			return d.Name == "Emergency" && d.Code == "EMERGENCY"
		})).Return(nil)

		dept, err := svc.Create(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, dept)
		assert.Equal(t, "Emergency", dept.Name)
		assert.Equal(t, "EMERGENCY", dept.Code)
		repo.AssertExpectations(t)
	})

	t.Run("error - name is required", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateDepartmentRequest{
			Name: "",
			Code: "EMERGENCY",
		}

		dept, err := svc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, dept)
		assert.Contains(t, err.Error(), "name is required")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - code is required", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateDepartmentRequest{
			Name: "Emergency",
			Code: "",
		}

		dept, err := svc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, dept)
		assert.Contains(t, err.Error(), "code is required")
		repo.AssertNotCalled(t, "Create")
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		req := CreateDepartmentRequest{
			Name: "Emergency",
			Code: "EMERGENCY",
		}

		expectedErr := errors.New("database error")
		repo.On("Create", mock.Anything, mock.Anything).Return(expectedErr)

		dept, err := svc.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, dept)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}

func TestService_GetByID(t *testing.T) {
	t.Run("success - returns department by ID", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedDept := &Department{
			ID:   "dept-1",
			Name: "Emergency",
			Code: "EMERGENCY",
		}

		repo.On("GetByID", mock.Anything, "dept-1").Return(expectedDept, nil)

		dept, err := svc.GetByID(context.Background(), "dept-1")

		assert.NoError(t, err)
		assert.NotNil(t, dept)
		assert.Equal(t, "dept-1", dept.ID)
		assert.Equal(t, "Emergency", dept.Name)
		repo.AssertExpectations(t)
	})

	t.Run("error - repository returns error", func(t *testing.T) {
		repo := new(mockRepo)
		svc := NewService(repo)

		expectedErr := errors.New("not found")
		repo.On("GetByID", mock.Anything, "dept-999").Return(nil, expectedErr)

		dept, err := svc.GetByID(context.Background(), "dept-999")

		assert.Error(t, err)
		assert.Nil(t, dept)
		assert.Equal(t, expectedErr, err)
		repo.AssertExpectations(t)
	})
}
