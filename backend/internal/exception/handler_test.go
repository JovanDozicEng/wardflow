package exception

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

func (m *MockService) Create(ctx context.Context, req *CreateExceptionRequest, byUserID string) (*ExceptionEvent, error) {
	args := m.Called(ctx, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExceptionEvent), args.Error(1)
}

func (m *MockService) Update(ctx context.Context, id string, req *UpdateExceptionRequest, byUserID string) (*ExceptionEvent, error) {
	args := m.Called(ctx, id, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExceptionEvent), args.Error(1)
}

func (m *MockService) Finalize(ctx context.Context, id string, byUserID string) (*ExceptionEvent, error) {
	args := m.Called(ctx, id, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExceptionEvent), args.Error(1)
}

func (m *MockService) Correct(ctx context.Context, id string, req *CorrectExceptionRequest, byUserID string) (*ExceptionEvent, error) {
	args := m.Called(ctx, id, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExceptionEvent), args.Error(1)
}

func (m *MockService) List(ctx context.Context, f ListExceptionsFilter) ([]*ExceptionEvent, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*ExceptionEvent), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetByID(ctx context.Context, id string) (*ExceptionEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExceptionEvent), args.Error(1)
}

func TestHandler_List(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:     "successful list",
			query:    "?status=draft&limit=10&offset=0",
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f ListExceptionsFilter) bool {
					return f.Status == ExceptionStatusDraft && f.Limit == 10 && f.Offset == 0
				})).Return([]*ExceptionEvent{
					{ID: "exc-1", Status: ExceptionStatusDraft},
					{ID: "exc-2", Status: ExceptionStatusDraft},
				}, int64(2), nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "with type filter",
			query:    "?type=medication-delay",
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f ListExceptionsFilter) bool {
					return f.Type == "medication-delay"
				})).Return([]*ExceptionEvent{}, int64(0), nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "service error",
			query:    "",
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("GET", "/api/v1/exceptions"+tt.query, nil, "user-1", tt.userRole)
			rr := httptest.NewRecorder()

			handler.List(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name         string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name: "successful creation",
			body: CreateExceptionRequest{
				EncounterID: "enc-1",
				Type:        "medication-delay",
				Data: map[string]interface{}{
					"medication": "aspirin",
				},
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("Create", mock.Anything, mock.Anything, "user-1").Return(&ExceptionEvent{
					ID:          "exc-1",
					EncounterID: "enc-1",
					Type:        "medication-delay",
					Status:      ExceptionStatusDraft,
					InitiatedBy: "user-1",
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "invalid body",
			body:         "invalid json",
			userRole:     models.RoleProvider,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "service validation error",
			body: CreateExceptionRequest{
				Type: "medication-delay",
				Data: map[string]interface{}{},
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("Create", mock.Anything, mock.Anything, "user-1").Return(nil, errors.New("encounterId is required"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/exceptions", tt.body, "user-1", tt.userRole)
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Update(t *testing.T) {
	tests := []struct {
		name         string
		exceptionID  string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:        "successful update",
			exceptionID: "exc-1",
			body: UpdateExceptionRequest{
				Data: map[string]interface{}{
					"medication": "ibuprofen",
				},
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
				m.On("Update", mock.Anything, "exc-1", mock.Anything, "user-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:        "charge nurse can update",
			exceptionID: "exc-1",
			body: UpdateExceptionRequest{
				Data: map[string]interface{}{},
			},
			userRole: models.RoleChargeNurse,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
				m.On("Update", mock.Anything, "exc-1", mock.Anything, "user-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:        "quality_safety can update",
			exceptionID: "exc-1",
			body: UpdateExceptionRequest{
				Data: map[string]interface{}{},
			},
			userRole: models.RoleQualitySafety,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
				m.On("Update", mock.Anything, "exc-1", mock.Anything, "user-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:        "insufficient permissions",
			exceptionID: "exc-1",
			body:        UpdateExceptionRequest{Data: map[string]interface{}{}},
			userRole:    models.RoleNurse,
			setupMock:   func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid body",
			exceptionID:  "exc-1",
			body:         "invalid",
			userRole:     models.RoleProvider,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "not found",
			exceptionID: "exc-999",
			body:        UpdateExceptionRequest{Data: map[string]interface{}{}},
			userRole:    models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-999").Return(nil, ErrNotFound)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:        "cannot update finalized",
			exceptionID: "exc-1",
			body:        UpdateExceptionRequest{Data: map[string]interface{}{}},
			userRole:    models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusFinalized,
				}, nil)
				m.On("Update", mock.Anything, "exc-1", mock.Anything, "user-1").Return(nil, errors.New("only draft exceptions can be updated"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("PATCH", "/api/v1/exceptions/"+tt.exceptionID, tt.body, "user-1", tt.userRole)
			req.SetPathValue("exceptionId", tt.exceptionID)
			rr := httptest.NewRecorder()

			handler.Update(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Finalize(t *testing.T) {
	tests := []struct {
		name         string
		exceptionID  string
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:        "successful finalize",
			exceptionID: "exc-1",
			userRole:    models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
				m.On("Finalize", mock.Anything, "exc-1", "user-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusFinalized,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:        "charge nurse can finalize",
			exceptionID: "exc-1",
			userRole:    models.RoleChargeNurse,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
				m.On("Finalize", mock.Anything, "exc-1", "user-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusFinalized,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "insufficient permissions",
			exceptionID:  "exc-1",
			userRole:     models.RoleNurse,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:        "not found",
			exceptionID: "exc-999",
			userRole:    models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-999").Return(nil, ErrNotFound)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:        "already finalized",
			exceptionID: "exc-1",
			userRole:    models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusFinalized,
				}, nil)
				m.On("Finalize", mock.Anything, "exc-1", "user-1").Return(nil, errors.New("only draft exceptions can be finalized"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/exceptions/"+tt.exceptionID+"/finalize", nil, "user-1", tt.userRole)
			req.SetPathValue("exceptionId", tt.exceptionID)
			rr := httptest.NewRecorder()

			handler.Finalize(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Correct(t *testing.T) {
	tests := []struct {
		name         string
		exceptionID  string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:        "successful correction",
			exceptionID: "exc-1",
			body: CorrectExceptionRequest{
				Reason: "corrected dosage",
				Data: map[string]interface{}{
					"medication": "aspirin",
					"dosage": "100mg",
				},
			},
			userRole: models.RoleQualitySafety,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusFinalized,
				}, nil)
				m.On("Correct", mock.Anything, "exc-1", mock.Anything, "user-1").Return(&ExceptionEvent{
					ID:     "exc-2",
					Status: ExceptionStatusFinalized,
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:        "admin can correct",
			exceptionID: "exc-1",
			body: CorrectExceptionRequest{
				Reason: "correction",
				Data:   map[string]interface{}{},
			},
			userRole: models.RoleAdmin,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusFinalized,
				}, nil)
				m.On("Correct", mock.Anything, "exc-1", mock.Anything, "user-1").Return(&ExceptionEvent{
					ID:     "exc-2",
					Status: ExceptionStatusFinalized,
				}, nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:        "insufficient permissions",
			exceptionID: "exc-1",
			body:        CorrectExceptionRequest{Reason: "test", Data: map[string]interface{}{}},
			userRole:    models.RoleProvider,
			setupMock:   func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid body",
			exceptionID:  "exc-1",
			body:         "invalid",
			userRole:     models.RoleQualitySafety,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:        "not found",
			exceptionID: "exc-999",
			body:        CorrectExceptionRequest{Reason: "test", Data: map[string]interface{}{}},
			userRole:    models.RoleQualitySafety,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-999").Return(nil, ErrNotFound)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:        "cannot correct draft",
			exceptionID: "exc-1",
			body:        CorrectExceptionRequest{Reason: "test", Data: map[string]interface{}{}},
			userRole:    models.RoleQualitySafety,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
				m.On("Correct", mock.Anything, "exc-1", mock.Anything, "user-1").Return(nil, errors.New("only finalized exceptions can be corrected"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/exceptions/"+tt.exceptionID+"/correct", tt.body, "user-1", tt.userRole)
			req.SetPathValue("exceptionId", tt.exceptionID)
			rr := httptest.NewRecorder()

			handler.Correct(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}
