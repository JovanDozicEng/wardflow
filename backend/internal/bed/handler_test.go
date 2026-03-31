package bed

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
	"github.com/wardflow/backend/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTestDB creates a simple in-memory SQLite database for testing
func newTestDB(t *testing.T) *database.DB {
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	// Auto-migrate audit log table (minimal schema) - table name is audit_log (singular)
	gormDB.Exec("CREATE TABLE IF NOT EXISTS audit_log (id TEXT PRIMARY KEY, entity_type TEXT, entity_id TEXT, action TEXT, at DATETIME, by_user_id TEXT, created_at DATETIME)")
	return &database.DB{DB: gormDB}
}

// MockService is a mock implementation of Service
type MockService struct {
	mock.Mock
}

func (m *MockService) ListBeds(ctx context.Context, filter ListBedsFilter) ([]Bed, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]Bed), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) CreateBed(ctx context.Context, req CreateBedRequest, userID string) (*Bed, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Bed), args.Error(1)
}

func (m *MockService) GetBed(ctx context.Context, id string) (*Bed, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Bed), args.Error(1)
}

func (m *MockService) UpdateBedStatus(ctx context.Context, bedID string, req UpdateBedStatusRequest, userID string) (*BedStatusEvent, error) {
	args := m.Called(ctx, bedID, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BedStatusEvent), args.Error(1)
}

func (m *MockService) CreateBedRequest(ctx context.Context, encounterID, userID string, req CreateBedRequestRequest) (*BedRequest, error) {
	args := m.Called(ctx, encounterID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BedRequest), args.Error(1)
}

func (m *MockService) AssignBed(ctx context.Context, requestID string, req AssignBedRequest, userID string) (*BedRequest, error) {
	args := m.Called(ctx, requestID, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BedRequest), args.Error(1)
}

func TestHandler_ListBeds(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:     "successful list",
			query:    "?unitId=unit-1&status=available&limit=10&offset=0",
			userRole: models.RoleAdmin,
			setupMock: func(m *MockService) {
				m.On("ListBeds", mock.Anything, mock.MatchedBy(func(f ListBedsFilter) bool {
					return f.UnitID == "unit-1" && f.Status == "available" && f.Limit == 10 && f.Offset == 0
				})).Return([]Bed{
					{ID: "bed-1", CurrentStatus: BedStatusAvailable},
					{ID: "bed-2", CurrentStatus: BedStatusAvailable},
				}, int64(2), nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "service error",
			query:    "",
			userRole: models.RoleAdmin,
			setupMock: func(m *MockService) {
				m.On("ListBeds", mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("GET", "/api/v1/beds"+tt.query, nil, "user-1", tt.userRole)
			rr := httptest.NewRecorder()

			handler.ListBeds(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_CreateBed(t *testing.T) {
	tests := []struct {
		name         string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name: "successful creation",
			body: CreateBedRequest{
				UnitID:       "unit-1",
				Room:         "101",
				Label:        "Bed A",
				Capabilities: []string{"telemetry"},
			},
			userRole: models.RoleAdmin,
			setupMock: func(m *MockService) {
				m.On("CreateBed", mock.Anything, mock.Anything, "user-1").Return(&Bed{
					ID:            "bed-1",
					UnitID:        "unit-1",
					Room:          "101",
					Label:         "Bed A",
					CurrentStatus: BedStatusAvailable,
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "non-admin forbidden",
			body:         CreateBedRequest{UnitID: "unit-1", Room: "101", Label: "Bed A"},
			userRole:     models.RoleProvider,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid body",
			body:         "invalid json",
			userRole:     models.RoleAdmin,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "service validation error",
			body: CreateBedRequest{
				Room:  "101",
				Label: "Bed A",
			},
			userRole: models.RoleAdmin,
			setupMock: func(m *MockService) {
				m.On("CreateBed", mock.Anything, mock.Anything, "user-1").Return(nil, errors.New("unitId, room, and label are required"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/beds", tt.body, "user-1", tt.userRole)
			rr := httptest.NewRecorder()

			handler.CreateBed(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_GetBed(t *testing.T) {
	tests := []struct {
		name         string
		bedID        string
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:     "successful get",
			bedID:    "bed-1",
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetBed", mock.Anything, "bed-1").Return(&Bed{
					ID:            "bed-1",
					CurrentStatus: BedStatusAvailable,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "not found",
			bedID:    "bed-999",
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetBed", mock.Anything, "bed-999").Return(nil, errors.New("not found"))
			},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("GET", "/api/v1/beds/"+tt.bedID, nil, "user-1", tt.userRole)
			req.SetPathValue("bedId", tt.bedID)
			rr := httptest.NewRecorder()

			handler.GetBed(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_UpdateBedStatus(t *testing.T) {
	tests := []struct {
		name         string
		bedID        string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:  "successful update by admin",
			bedID: "bed-1",
			body: UpdateBedStatusRequest{
				Status: BedStatusCleaning,
			},
			userRole: models.RoleAdmin,
			setupMock: func(m *MockService) {
				fromStatus := BedStatusAvailable
				m.On("UpdateBedStatus", mock.Anything, "bed-1", mock.Anything, "user-1").Return(&BedStatusEvent{
					ID:         "event-1",
					BedID:      "bed-1",
					FromStatus: &fromStatus,
					ToStatus:   BedStatusCleaning,
					ChangedBy:  "user-1",
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:  "operations can update",
			bedID: "bed-1",
			body: UpdateBedStatusRequest{
				Status: BedStatusCleaning,
			},
			userRole: models.RoleOperations,
			setupMock: func(m *MockService) {
				fromStatus := BedStatusAvailable
				m.On("UpdateBedStatus", mock.Anything, "bed-1", mock.Anything, "user-1").Return(&BedStatusEvent{
					ID:         "event-1",
					BedID:      "bed-1",
					FromStatus: &fromStatus,
					ToStatus:   BedStatusCleaning,
					ChangedBy:  "user-1",
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:  "charge nurse can update",
			bedID: "bed-1",
			body: UpdateBedStatusRequest{
				Status: BedStatusCleaning,
			},
			userRole: models.RoleChargeNurse,
			setupMock: func(m *MockService) {
				fromStatus := BedStatusAvailable
				m.On("UpdateBedStatus", mock.Anything, "bed-1", mock.Anything, "user-1").Return(&BedStatusEvent{
					ID:         "event-1",
					BedID:      "bed-1",
					FromStatus: &fromStatus,
					ToStatus:   BedStatusCleaning,
					ChangedBy:  "user-1",
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:  "insufficient permissions",
			bedID: "bed-1",
			body: UpdateBedStatusRequest{
				Status: BedStatusCleaning,
			},
			userRole:     models.RoleNurse,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid body",
			bedID:        "bed-1",
			body:         "invalid",
			userRole:     models.RoleAdmin,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:  "service error",
			bedID: "bed-1",
			body: UpdateBedStatusRequest{
				Status: BedStatusCleaning,
			},
			userRole: models.RoleAdmin,
			setupMock: func(m *MockService) {
				m.On("UpdateBedStatus", mock.Anything, "bed-1", mock.Anything, "user-1").Return(nil, errors.New("bed not found"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/beds/"+tt.bedID+"/status", tt.body, "user-1", tt.userRole)
			req.SetPathValue("bedId", tt.bedID)
			rr := httptest.NewRecorder()

			handler.UpdateBedStatus(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_CreateBedRequest(t *testing.T) {
	tests := []struct {
		name         string
		encounterID  string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:        "successful creation",
			encounterID: "enc-1",
			body: CreateBedRequestRequest{
				RequiredCapabilities: []string{"telemetry"},
				Priority:             "urgent",
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("CreateBedRequest", mock.Anything, "enc-1", "user-1", mock.Anything).Return(&BedRequest{
					ID:          "req-1",
					EncounterID: "enc-1",
					Priority:    "urgent",
					Status:      BedRequestStatusPending,
					CreatedBy:   "user-1",
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "invalid body",
			encounterID:  "enc-1",
			body:         "invalid",
			userRole:     models.RoleProvider,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "service error",
			encounterID: "enc-1",
			body: CreateBedRequestRequest{
				RequiredCapabilities: []string{},
				Priority:             "routine",
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("CreateBedRequest", mock.Anything, "enc-1", "user-1", mock.Anything).Return(nil, errors.New("encounter not found"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/encounters/"+tt.encounterID+"/bed-requests", tt.body, "user-1", tt.userRole)
			req.SetPathValue("encounterId", tt.encounterID)
			rr := httptest.NewRecorder()

			handler.CreateBedRequest(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_AssignBed(t *testing.T) {
	tests := []struct {
		name         string
		requestID    string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:      "successful assignment",
			requestID: "req-1",
			body: AssignBedRequest{
				BedID: "bed-1",
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				bedID := "bed-1"
				m.On("AssignBed", mock.Anything, "req-1", mock.Anything, "user-1").Return(&BedRequest{
					ID:            "req-1",
					Status:        BedRequestStatusAssigned,
					AssignedBedID: &bedID,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid body",
			requestID:    "req-1",
			body:         "invalid",
			userRole:     models.RoleProvider,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "bed not available",
			requestID: "req-1",
			body: AssignBedRequest{
				BedID: "bed-1",
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("AssignBed", mock.Anything, "req-1", mock.Anything, "user-1").Return(nil, errors.New("bed is not available for assignment"))
			},
			expectedCode: http.StatusConflict,
		},
		{
			name:      "request not pending",
			requestID: "req-1",
			body: AssignBedRequest{
				BedID: "bed-1",
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("AssignBed", mock.Anything, "req-1", mock.Anything, "user-1").Return(nil, errors.New("bed request is not pending"))
			},
			expectedCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/bed-requests/"+tt.requestID+"/assign", tt.body, "user-1", tt.userRole)
			req.SetPathValue("requestId", tt.requestID)
			rr := httptest.NewRecorder()

			handler.AssignBed(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
