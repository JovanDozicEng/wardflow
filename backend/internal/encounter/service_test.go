package encounter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepository is a mock implementation of Repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, e *Encounter) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*Encounter, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Encounter), args.Error(1)
}

func (m *mockRepository) List(ctx context.Context, f ListEncountersFilter) ([]*Encounter, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Encounter), args.Get(1).(int64), args.Error(2)
}

func (m *mockRepository) Update(ctx context.Context, e *Encounter) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func TestService_Create_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &CreateEncounterRequest{
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
	}

	repo.On("Create", mock.Anything, mock.AnythingOfType("*encounter.Encounter")).
		Run(func(args mock.Arguments) {
			// Simulate database auto-generating ID
			enc := args.Get(1).(*Encounter)
			enc.ID = "enc-123"
		}).
		Return(nil)

	enc, err := svc.Create(context.Background(), req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, enc)
	assert.Equal(t, "enc-123", enc.ID)
	assert.Equal(t, "patient-1", enc.PatientID)
	assert.Equal(t, "unit-1", enc.UnitID)
	assert.Equal(t, "dept-1", enc.DepartmentID)
	assert.Equal(t, EncounterStatusActive, enc.Status)
	assert.Equal(t, "user-1", enc.CreatedBy)
	assert.Equal(t, "user-1", enc.UpdatedBy)
	assert.False(t, enc.StartedAt.IsZero())
	repo.AssertExpectations(t)
}

func TestService_Create_WithCustomStartedAt(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	customTime := time.Now().UTC().Add(-24 * time.Hour)
	req := &CreateEncounterRequest{
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
		StartedAt:    &customTime,
	}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(e *Encounter) bool {
		return e.StartedAt.Equal(customTime)
	})).Return(nil)

	enc, err := svc.Create(context.Background(), req, "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, enc)
	assert.True(t, enc.StartedAt.Equal(customTime))
	repo.AssertExpectations(t)
}

func TestService_Create_MissingPatientID(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &CreateEncounterRequest{
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
	}

	enc, err := svc.Create(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, "patientId is required", err.Error())
}

func TestService_Create_MissingUnitID(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &CreateEncounterRequest{
		PatientID:    "patient-1",
		DepartmentID: "dept-1",
	}

	enc, err := svc.Create(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, "unitId is required", err.Error())
}

func TestService_Create_MissingDepartmentID(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &CreateEncounterRequest{
		PatientID: "patient-1",
		UnitID:    "unit-1",
	}

	enc, err := svc.Create(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, "departmentId is required", err.Error())
}

func TestService_Create_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &CreateEncounterRequest{
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
	}

	dbErr := errors.New("database error")
	repo.On("Create", mock.Anything, mock.AnythingOfType("*encounter.Encounter")).Return(dbErr)

	enc, err := svc.Create(context.Background(), req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, dbErr, err)
	repo.AssertExpectations(t)
}

func TestService_GetByID_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	expectedEnc := &Encounter{
		ID:           "enc-1",
		PatientID:    "patient-1",
		UnitID:       "unit-1",
		DepartmentID: "dept-1",
		Status:       EncounterStatusActive,
	}

	repo.On("GetByID", mock.Anything, "enc-1").Return(expectedEnc, nil)

	enc, err := svc.GetByID(context.Background(), "enc-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedEnc, enc)
	repo.AssertExpectations(t)
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	repo.On("GetByID", mock.Anything, "enc-999").Return(nil, ErrNotFound)

	enc, err := svc.GetByID(context.Background(), "enc-999")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, ErrNotFound, err)
	repo.AssertExpectations(t)
}

func TestService_List_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	filter := ListEncountersFilter{
		UnitID: "unit-1",
		Limit:  10,
		Offset: 0,
	}

	expectedEncs := []*Encounter{
		{ID: "enc-1", UnitID: "unit-1"},
		{ID: "enc-2", UnitID: "unit-1"},
	}

	repo.On("List", mock.Anything, filter).Return(expectedEncs, int64(2), nil)

	encs, total, err := svc.List(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedEncs, encs)
	assert.Equal(t, int64(2), total)
	repo.AssertExpectations(t)
}

func TestService_List_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	filter := ListEncountersFilter{Limit: 10}
	dbErr := errors.New("database error")

	repo.On("List", mock.Anything, filter).Return(nil, int64(0), dbErr)

	encs, total, err := svc.List(context.Background(), filter)

	assert.Error(t, err)
	assert.Nil(t, encs)
	assert.Equal(t, int64(0), total)
	assert.Equal(t, dbErr, err)
	repo.AssertExpectations(t)
}

func TestService_Update_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingEnc := &Encounter{
		ID:        "enc-1",
		Status:    EncounterStatusActive,
		UnitID:    "unit-1",
		CreatedBy: "user-1",
		UpdatedBy: "user-1",
	}

	newStatus := EncounterStatusDischarged
	endedAt := time.Now().UTC()
	newUnitID := "unit-2"

	req := &UpdateEncounterRequest{
		Status:  &newStatus,
		EndedAt: &endedAt,
		UnitID:  &newUnitID,
	}

	repo.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)
	repo.On("Update", mock.Anything, mock.MatchedBy(func(e *Encounter) bool {
		return e.ID == "enc-1" &&
			e.Status == EncounterStatusDischarged &&
			e.UnitID == "unit-2" &&
			e.UpdatedBy == "user-2"
	})).Return(nil)

	enc, err := svc.Update(context.Background(), "enc-1", req, "user-2")

	assert.NoError(t, err)
	assert.NotNil(t, enc)
	assert.Equal(t, EncounterStatusDischarged, enc.Status)
	assert.Equal(t, "unit-2", enc.UnitID)
	assert.Equal(t, "user-2", enc.UpdatedBy)
	assert.NotNil(t, enc.EndedAt)
	repo.AssertExpectations(t)
}

func TestService_Update_CannotReactivateDischargedEncounter(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingEnc := &Encounter{
		ID:     "enc-1",
		Status: EncounterStatusDischarged,
	}

	newStatus := EncounterStatusActive
	req := &UpdateEncounterRequest{
		Status: &newStatus,
	}

	repo.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)

	enc, err := svc.Update(context.Background(), "enc-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, "cannot re-activate discharged or cancelled encounter", err.Error())
	repo.AssertExpectations(t)
}

func TestService_Update_CannotReactivateCancelledEncounter(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingEnc := &Encounter{
		ID:     "enc-1",
		Status: EncounterStatusCancelled,
	}

	newStatus := EncounterStatusActive
	req := &UpdateEncounterRequest{
		Status: &newStatus,
	}

	repo.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)

	enc, err := svc.Update(context.Background(), "enc-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, "cannot re-activate discharged or cancelled encounter", err.Error())
	repo.AssertExpectations(t)
}

func TestService_Update_NotFound(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	req := &UpdateEncounterRequest{}

	repo.On("GetByID", mock.Anything, "enc-999").Return(nil, ErrNotFound)

	enc, err := svc.Update(context.Background(), "enc-999", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, ErrNotFound, err)
	repo.AssertExpectations(t)
}

func TestService_Update_RepositoryError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo)

	existingEnc := &Encounter{
		ID:     "enc-1",
		Status: EncounterStatusActive,
	}

	newStatus := EncounterStatusDischarged
	req := &UpdateEncounterRequest{
		Status: &newStatus,
	}

	dbErr := errors.New("database error")
	repo.On("GetByID", mock.Anything, "enc-1").Return(existingEnc, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*encounter.Encounter")).Return(dbErr)

	enc, err := svc.Update(context.Background(), "enc-1", req, "user-1")

	assert.Error(t, err)
	assert.Nil(t, enc)
	assert.Equal(t, dbErr, err)
	repo.AssertExpectations(t)
}
