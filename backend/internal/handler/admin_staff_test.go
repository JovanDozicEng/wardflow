package handler

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
	"gorm.io/gorm"
)

// Mock StaffService
type mockStaffService struct {
	mock.Mock
}

func (m *mockStaffService) ListStaff(ctx context.Context, q, role string, limit, offset int) ([]StaffProfile, int64, error) {
	args := m.Called(ctx, q, role, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]StaffProfile), args.Get(1).(int64), args.Error(2)
}

func (m *mockStaffService) UpdateStaff(ctx context.Context, userID string, req UpdateStaffRequest) (*StaffProfile, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*StaffProfile), args.Error(1)
}

func TestAdminStaffHandler_ListStaff(t *testing.T) {
	t.Run("returns paginated staff list for admin", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		profiles := []StaffProfile{
			{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: models.RoleNurse, IsActive: true},
			{ID: "user-2", Name: "Bob", Email: "bob@example.com", Role: models.RoleProvider, IsActive: true},
		}

		svc.On("ListStaff", mock.Anything, "", "", 20, 0).Return(profiles, int64(2), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/admin/staff", nil, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListStaff(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result map[string]interface{}
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, float64(2), result["total"])
		assert.Equal(t, float64(20), result["limit"])
		assert.Equal(t, float64(0), result["offset"])

		svc.AssertExpectations(t)
	})

	t.Run("returns forbidden for non-admin", func(t *testing.T) {
		handler := NewAdminStaffHandler(nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/admin/staff", nil, "user-1", models.RoleNurse)
		rr := httptest.NewRecorder()

		handler.ListStaff(rr, r)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("filters by search query", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		profiles := []StaffProfile{
			{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: models.RoleNurse, IsActive: true},
		}

		svc.On("ListStaff", mock.Anything, "alice", "", 20, 0).Return(profiles, int64(1), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/admin/staff?q=alice", nil, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListStaff(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result map[string]interface{}
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, float64(1), result["total"])

		svc.AssertExpectations(t)
	})

	t.Run("filters by role", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		profiles := []StaffProfile{
			{ID: "user-1", Name: "Alice", Email: "alice@example.com", Role: models.RoleNurse, IsActive: true},
		}

		svc.On("ListStaff", mock.Anything, "", "nurse", 20, 0).Return(profiles, int64(1), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/admin/staff?role=nurse", nil, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListStaff(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("respects pagination parameters", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		profiles := []StaffProfile{
			{ID: "user-11", Name: "User 11", Email: "user11@example.com", Role: models.RoleNurse, IsActive: true},
		}

		svc.On("ListStaff", mock.Anything, "", "", 10, 10).Return(profiles, int64(25), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/admin/staff?limit=10&offset=10", nil, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListStaff(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result map[string]interface{}
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, float64(25), result["total"])
		assert.Equal(t, float64(10), result["limit"])
		assert.Equal(t, float64(10), result["offset"])

		svc.AssertExpectations(t)
	})

	t.Run("enforces maximum limit", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		profiles := []StaffProfile{}

		// Should default to limit of 20, not cap at 100, based on the actual implementation
		// The handler checks limit <= 100 but defaults to 20
		svc.On("ListStaff", mock.Anything, "", "", 20, 0).Return(profiles, int64(0), nil)

		r := testutil.NewRequest(http.MethodGet, "/api/v1/admin/staff?limit=200", nil, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListStaff(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns internal error on service failure", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		svc.On("ListStaff", mock.Anything, "", "", 20, 0).Return(nil, int64(0), errors.New("database error"))

		r := testutil.NewRequest(http.MethodGet, "/api/v1/admin/staff", nil, "admin-1", models.RoleAdmin)
		rr := httptest.NewRecorder()

		handler.ListStaff(rr, r)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		svc.AssertExpectations(t)
	})
}

func TestAdminStaffHandler_UpdateStaff(t *testing.T) {
	t.Run("successfully updates staff role", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		userID := "user-123"
		newRole := models.RoleChargeNurse

		reqBody := UpdateStaffRequest{
			Role: &newRole,
		}

		profile := &StaffProfile{
			ID:       userID,
			Name:     "Alice",
			Email:    "alice@example.com",
			Role:     newRole,
			IsActive: true,
		}

		svc.On("UpdateStaff", mock.Anything, userID, reqBody).Return(profile, nil)

		r := testutil.NewRequest(http.MethodPatch, "/api/v1/admin/staff/"+userID, reqBody, "admin-1", models.RoleAdmin)
		r.SetPathValue("userId", userID)
		rr := httptest.NewRecorder()

		handler.UpdateStaff(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result StaffProfile
		testutil.DecodeJSON(t, rr, &result)
		assert.Equal(t, newRole, result.Role)

		svc.AssertExpectations(t)
	})

	t.Run("successfully updates staff active status", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		userID := "user-123"
		isActive := false

		reqBody := UpdateStaffRequest{
			IsActive: &isActive,
		}

		profile := &StaffProfile{
			ID:       userID,
			Name:     "Alice",
			Email:    "alice@example.com",
			Role:     models.RoleNurse,
			IsActive: false,
		}

		svc.On("UpdateStaff", mock.Anything, userID, reqBody).Return(profile, nil)

		r := testutil.NewRequest(http.MethodPatch, "/api/v1/admin/staff/"+userID, reqBody, "admin-1", models.RoleAdmin)
		r.SetPathValue("userId", userID)
		rr := httptest.NewRecorder()

		handler.UpdateStaff(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result StaffProfile
		testutil.DecodeJSON(t, rr, &result)
		assert.False(t, result.IsActive)

		svc.AssertExpectations(t)
	})

	t.Run("successfully updates unit assignments", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		userID := "user-123"
		unitIDs := models.StringArray{"unit-1", "unit-2"}

		reqBody := UpdateStaffRequest{
			UnitIDs: &unitIDs,
		}

		profile := &StaffProfile{
			ID:      userID,
			Name:    "Alice",
			Email:   "alice@example.com",
			Role:    models.RoleNurse,
			UnitIDs: unitIDs,
		}

		svc.On("UpdateStaff", mock.Anything, userID, reqBody).Return(profile, nil)

		r := testutil.NewRequest(http.MethodPatch, "/api/v1/admin/staff/"+userID, reqBody, "admin-1", models.RoleAdmin)
		r.SetPathValue("userId", userID)
		rr := httptest.NewRecorder()

		handler.UpdateStaff(rr, r)

		assert.Equal(t, http.StatusOK, rr.Code)

		var result StaffProfile
		testutil.DecodeJSON(t, rr, &result)
		assert.Len(t, result.UnitIDs, 2)

		svc.AssertExpectations(t)
	})

	t.Run("returns forbidden for non-admin", func(t *testing.T) {
		handler := NewAdminStaffHandler(nil)

		reqBody := UpdateStaffRequest{}

		r := testutil.NewRequest(http.MethodPatch, "/api/v1/admin/staff/user-123", reqBody, "user-1", models.RoleNurse)
		r.SetPathValue("userId", "user-123")
		rr := httptest.NewRecorder()

		handler.UpdateStaff(rr, r)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("returns not found for non-existent user", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		userID := "nonexistent"
		reqBody := UpdateStaffRequest{}

		svc.On("UpdateStaff", mock.Anything, userID, reqBody).Return(nil, gorm.ErrRecordNotFound)

		r := testutil.NewRequest(http.MethodPatch, "/api/v1/admin/staff/"+userID, reqBody, "admin-1", models.RoleAdmin)
		r.SetPathValue("userId", userID)
		rr := httptest.NewRecorder()

		handler.UpdateStaff(rr, r)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request for invalid role", func(t *testing.T) {
		svc := new(mockStaffService)
		handler := NewAdminStaffHandler(svc)

		userID := "user-123"
		reqBody := UpdateStaffRequest{}

		svc.On("UpdateStaff", mock.Anything, userID, reqBody).Return(nil, errors.New("invalid role"))

		r := testutil.NewRequest(http.MethodPatch, "/api/v1/admin/staff/"+userID, reqBody, "admin-1", models.RoleAdmin)
		r.SetPathValue("userId", userID)
		rr := httptest.NewRecorder()

		handler.UpdateStaff(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		svc.AssertExpectations(t)
	})

	t.Run("returns bad request for missing userId", func(t *testing.T) {
		handler := NewAdminStaffHandler(nil)

		reqBody := UpdateStaffRequest{}

		r := testutil.NewRequest(http.MethodPatch, "/api/v1/admin/staff/", reqBody, "admin-1", models.RoleAdmin)
		// Don't set path value
		rr := httptest.NewRecorder()

		handler.UpdateStaff(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns bad request for invalid JSON", func(t *testing.T) {
		handler := NewAdminStaffHandler(nil)

		r := testutil.NewRequestNoAuth(http.MethodPatch, "/api/v1/admin/staff/user-123", nil)
		r = r.WithContext(testutil.WithUser(r.Context(), "admin-1", models.RoleAdmin))
		r.SetPathValue("userId", "user-123")
		r.Body = http.NoBody
		rr := httptest.NewRecorder()

		handler.UpdateStaff(rr, r)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
