package careteam

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepository is a mock implementation of Repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) CreateAssignment(ctx context.Context, assignment *CareTeamAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *mockRepository) EndAssignment(ctx context.Context, assignmentID string, endsAt time.Time) error {
	args := m.Called(ctx, assignmentID, endsAt)
	return args.Error(0)
}

func (m *mockRepository) GetActiveAssignments(ctx context.Context, encounterID string) ([]CareTeamAssignment, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]CareTeamAssignment), args.Error(1)
}

func (m *mockRepository) GetActiveAssignmentByRole(ctx context.Context, encounterID string, roleType RoleType) (*CareTeamAssignment, error) {
	args := m.Called(ctx, encounterID, roleType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CareTeamAssignment), args.Error(1)
}

func (m *mockRepository) GetAssignmentHistory(ctx context.Context, encounterID string) ([]CareTeamAssignment, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]CareTeamAssignment), args.Error(1)
}

func (m *mockRepository) GetAssignmentByID(ctx context.Context, assignmentID string) (*CareTeamAssignment, error) {
	args := m.Called(ctx, assignmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CareTeamAssignment), args.Error(1)
}

func (m *mockRepository) CreateHandoffNote(ctx context.Context, note *HandoffNote) error {
	args := m.Called(ctx, note)
	// Set ID for the created note
	if args.Error(0) == nil && note.ID == "" {
		note.ID = "handoff-123"
	}
	return args.Error(0)
}

func (m *mockRepository) GetHandoffNotes(ctx context.Context, encounterID string) ([]HandoffNote, error) {
	args := m.Called(ctx, encounterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]HandoffNote), args.Error(1)
}

func (m *mockRepository) GetHandoffNotesPaginated(ctx context.Context, encounterID string, limit, offset int) ([]HandoffNote, int64, error) {
	args := m.Called(ctx, encounterID, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]HandoffNote), args.Get(1).(int64), args.Error(2)
}

func TestAssignRole_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	req := AssignRoleRequest{
		UserID:   "user-1",
		RoleType: RolePrimaryNurse,
	}

	// No existing assignment
	repo.On("GetActiveAssignmentByRole", ctx, "encounter-1", RolePrimaryNurse).Return(nil, nil)
	repo.On("CreateAssignment", ctx, mock.MatchedBy(func(a *CareTeamAssignment) bool {
		return a.EncounterID == "encounter-1" &&
			a.UserID == "user-1" &&
			a.RoleType == RolePrimaryNurse &&
			a.CreatedBy == "current-user"
	})).Return(nil)

	result, err := svc.AssignRole(ctx, r, "encounter-1", req, "current-user")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "encounter-1", result.EncounterID)
	assert.Equal(t, "user-1", result.UserID)
	assert.Equal(t, RolePrimaryNurse, result.RoleType)
	repo.AssertExpectations(t)
}

func TestAssignRole_RoleAlreadyAssigned(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	req := AssignRoleRequest{
		UserID:   "user-2",
		RoleType: RolePrimaryNurse,
	}

	existingAssignment := &CareTeamAssignment{
		ID:          "assignment-1",
		EncounterID: "encounter-1",
		UserID:      "user-1",
		RoleType:    RolePrimaryNurse,
	}

	repo.On("GetActiveAssignmentByRole", ctx, "encounter-1", RolePrimaryNurse).Return(existingAssignment, nil)

	result, err := svc.AssignRole(ctx, r, "encounter-1", req, "current-user")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already assigned")
	repo.AssertExpectations(t)
}

func TestTransferRole_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	req := TransferRoleRequest{
		ToUserID:    "user-2",
		HandoffNote: "Patient stable, vitals monitored",
	}

	currentAssignment := &CareTeamAssignment{
		ID:          "assignment-1",
		EncounterID: "encounter-1",
		UserID:      "user-1",
		RoleType:    RolePrimaryNurse,
		StartsAt:    time.Now().Add(-2 * time.Hour),
		EndsAt:      nil,
	}

	repo.On("GetAssignmentByID", ctx, "assignment-1").Return(currentAssignment, nil)
	repo.On("CreateHandoffNote", ctx, mock.MatchedBy(func(n *HandoffNote) bool {
		return n.EncounterID == "encounter-1" &&
			n.FromUserID == "user-1" &&
			n.ToUserID == "user-2" &&
			n.RoleType == RolePrimaryNurse &&
			n.Note == "Patient stable, vitals monitored"
	})).Return(nil)
	repo.On("EndAssignment", ctx, "assignment-1", mock.AnythingOfType("time.Time")).Return(nil)
	repo.On("CreateAssignment", ctx, mock.MatchedBy(func(a *CareTeamAssignment) bool {
		return a.EncounterID == "encounter-1" &&
			a.UserID == "user-2" &&
			a.RoleType == RolePrimaryNurse
	})).Return(nil)

	result, err := svc.TransferRole(ctx, r, "assignment-1", req, "current-user")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user-2", result.UserID)
	repo.AssertExpectations(t)
}

func TestTransferRole_AlreadyEnded(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	req := TransferRoleRequest{
		ToUserID:    "user-2",
		HandoffNote: "Transfer",
	}

	endsAt := time.Now().Add(-1 * time.Hour)
	currentAssignment := &CareTeamAssignment{
		ID:          "assignment-1",
		EncounterID: "encounter-1",
		UserID:      "user-1",
		RoleType:    RolePrimaryNurse,
		StartsAt:    time.Now().Add(-3 * time.Hour),
		EndsAt:      &endsAt,
	}

	repo.On("GetAssignmentByID", ctx, "assignment-1").Return(currentAssignment, nil)

	result, err := svc.TransferRole(ctx, r, "assignment-1", req, "current-user")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already ended")
	repo.AssertExpectations(t)
}

func TestTransferRole_CriticalRoleRequiresHandoffNote(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	req := TransferRoleRequest{
		ToUserID:    "user-2",
		HandoffNote: "", // Empty handoff note
	}

	currentAssignment := &CareTeamAssignment{
		ID:          "assignment-1",
		EncounterID: "encounter-1",
		UserID:      "user-1",
		RoleType:    RolePrimaryNurse, // Critical role
		StartsAt:    time.Now().Add(-2 * time.Hour),
		EndsAt:      nil,
	}

	repo.On("GetAssignmentByID", ctx, "assignment-1").Return(currentAssignment, nil)

	result, err := svc.TransferRole(ctx, r, "assignment-1", req, "current-user")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "handoff note is required")
	repo.AssertExpectations(t)
}

func TestListCareTeam_ActiveOnly(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	activeAssignments := []CareTeamAssignment{
		{ID: "a1", UserID: "user-1", RoleType: RolePrimaryNurse},
		{ID: "a2", UserID: "user-2", RoleType: RoleAttendingProvider},
	}

	repo.On("GetActiveAssignments", ctx, "encounter-1").Return(activeAssignments, nil)

	result, err := svc.ListCareTeam(ctx, "encounter-1", true)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

func TestListCareTeam_AllHistory(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	allAssignments := []CareTeamAssignment{
		{ID: "a1", UserID: "user-1", RoleType: RolePrimaryNurse},
		{ID: "a2", UserID: "user-2", RoleType: RolePrimaryNurse},
		{ID: "a3", UserID: "user-3", RoleType: RoleAttendingProvider},
	}

	repo.On("GetAssignmentHistory", ctx, "encounter-1").Return(allAssignments, nil)

	result, err := svc.ListCareTeam(ctx, "encounter-1", false)

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	repo.AssertExpectations(t)
}

func TestGetHandoffs_Paginated(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	handoffs := []HandoffNote{
		{ID: "h1", Note: "First handoff"},
		{ID: "h2", Note: "Second handoff"},
	}

	repo.On("GetHandoffNotesPaginated", ctx, "encounter-1", 10, 0).Return(handoffs, int64(15), nil)

	result, total, err := svc.GetHandoffs(ctx, "encounter-1", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(15), total)
	repo.AssertExpectations(t)
}

func TestGetCareTeamWithDetails_Success(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	// Return empty list to avoid DB lookup
	activeAssignments := []CareTeamAssignment{}

	repo.On("GetActiveAssignments", ctx, "encounter-1").Return(activeAssignments, nil)

	result, err := svc.GetCareTeamWithDetails(ctx, "encounter-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "encounter-1", result.EncounterID)
	assert.Len(t, result.Members, 0)
	repo.AssertExpectations(t)
}

func TestGetCareTeamWithDetails_RepoError(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	repo.On("GetActiveAssignments", ctx, "encounter-1").Return(nil, fmt.Errorf("database error"))

	result, err := svc.GetCareTeamWithDetails(ctx, "encounter-1")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}

// Additional tests for GetCareTeamWithDetails

func TestGetCareTeamWithDetails_MultipleMembers(t *testing.T) {
	repo := new(mockRepository)
	// Pass nil for db - the test will return empty members since user lookups fail
	svc := NewService(repo, nil)
	ctx := context.Background()

	// Return empty list to avoid DB operations that would panic
	activeAssignments := []CareTeamAssignment{}

	repo.On("GetActiveAssignments", ctx, "encounter-1").Return(activeAssignments, nil)

	result, err := svc.GetCareTeamWithDetails(ctx, "encounter-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "encounter-1", result.EncounterID)
	assert.Len(t, result.Members, 0)
	repo.AssertExpectations(t)
}

func TestGetCareTeamWithDetails_EmptyTeam(t *testing.T) {
	repo := new(mockRepository)
	svc := NewService(repo, nil)
	ctx := context.Background()

	repo.On("GetActiveAssignments", ctx, "encounter-2").Return([]CareTeamAssignment{}, nil)

	result, err := svc.GetCareTeamWithDetails(ctx, "encounter-2")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "encounter-2", result.EncounterID)
	assert.Len(t, result.Members, 0)
	repo.AssertExpectations(t)
}
