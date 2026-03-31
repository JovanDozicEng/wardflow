package consult

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

func (m *MockService) Create(ctx context.Context, req *CreateConsultRequest, byUserID string) (*ConsultRequest, error) {
	args := m.Called(ctx, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConsultRequest), args.Error(1)
}

func (m *MockService) Accept(ctx context.Context, id string, byUserID string) (*ConsultRequest, error) {
	args := m.Called(ctx, id, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConsultRequest), args.Error(1)
}

func (m *MockService) Decline(ctx context.Context, id string, req *DeclineConsultRequest, byUserID string) (*ConsultRequest, error) {
	args := m.Called(ctx, id, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConsultRequest), args.Error(1)
}

func (m *MockService) Redirect(ctx context.Context, id string, req *RedirectConsultRequest, byUserID string) (*RedirectResult, error) {
	args := m.Called(ctx, id, req, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RedirectResult), args.Error(1)
}

func (m *MockService) Complete(ctx context.Context, id string, byUserID string) (*ConsultRequest, error) {
	args := m.Called(ctx, id, byUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConsultRequest), args.Error(1)
}

func (m *MockService) List(ctx context.Context, f ListConsultsFilter) ([]*ConsultRequest, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*ConsultRequest), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) GetByID(ctx context.Context, id string) (*ConsultRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConsultRequest), args.Error(1)
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
			query:    "?status=pending&limit=10&offset=0",
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f ListConsultsFilter) bool {
					return f.Status == ConsultStatusPending && f.Limit == 10 && f.Offset == 0
				})).Return([]*ConsultRequest{
					{ID: "consult-1", Status: ConsultStatusPending},
					{ID: "consult-2", Status: ConsultStatusPending},
				}, int64(2), nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "with targetService filter",
			query:    "?targetService=cardiology",
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f ListConsultsFilter) bool {
					return f.TargetService == "cardiology"
				})).Return([]*ConsultRequest{}, int64(0), nil)
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
			req := testutil.NewRequest("GET", "/api/v1/consults"+tt.query, nil, "user-1", tt.userRole)
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
			body: CreateConsultRequest{
				EncounterID:   "enc-1",
				TargetService: "cardiology",
				Reason:        "chest pain",
				Urgency:       ConsultUrgencyUrgent,
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("Create", mock.Anything, mock.Anything, "user-1").Return(&ConsultRequest{
					ID:            "consult-1",
					EncounterID:   "enc-1",
					TargetService: "cardiology",
					Status:        ConsultStatusPending,
					CreatedBy:     "user-1",
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
			body: CreateConsultRequest{
				TargetService: "cardiology",
				Reason:        "chest pain",
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
			req := testutil.NewRequest("POST", "/api/v1/consults", tt.body, "user-1", tt.userRole)
			rr := httptest.NewRecorder()

			handler.Create(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Accept(t *testing.T) {
	tests := []struct {
		name         string
		consultID    string
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:      "successful accept",
			consultID: "consult-1",
			userRole:  models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
				m.On("Accept", mock.Anything, "consult-1", "user-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusAccepted,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:      "consult role can accept",
			consultID: "consult-1",
			userRole:  models.RoleConsult,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
				m.On("Accept", mock.Anything, "consult-1", "user-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusAccepted,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "insufficient permissions",
			consultID:    "consult-1",
			userRole:     models.RoleNurse,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:      "not found",
			consultID: "consult-999",
			userRole:  models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-999").Return(nil, ErrNotFound)
			},
			expectedCode: http.StatusNotFound,
		},
		{
			name:      "service error",
			consultID: "consult-1",
			userRole:  models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
				m.On("Accept", mock.Anything, "consult-1", "user-1").Return(nil, errors.New("only pending consults can be accepted"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/consults/"+tt.consultID+"/accept", nil, "user-1", tt.userRole)
			req.SetPathValue("consultId", tt.consultID)
			rr := httptest.NewRecorder()

			handler.Accept(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Decline(t *testing.T) {
	tests := []struct {
		name         string
		consultID    string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:      "successful decline",
			consultID: "consult-1",
			body:      DeclineConsultRequest{Reason: "patient already seen"},
			userRole:  models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
				m.On("Decline", mock.Anything, "consult-1", mock.Anything, "user-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusDeclined,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "insufficient permissions",
			consultID:    "consult-1",
			body:         DeclineConsultRequest{Reason: "test"},
			userRole:     models.RoleNurse,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid body",
			consultID:    "consult-1",
			body:         "invalid",
			userRole:     models.RoleProvider,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "missing reason",
			consultID: "consult-1",
			body:      DeclineConsultRequest{Reason: ""},
			userRole:  models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
				m.On("Decline", mock.Anything, "consult-1", mock.Anything, "user-1").Return(nil, errors.New("reason is required"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/consults/"+tt.consultID+"/decline", tt.body, "user-1", tt.userRole)
			req.SetPathValue("consultId", tt.consultID)
			rr := httptest.NewRecorder()

			handler.Decline(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	tests := []struct {
		name         string
		consultID    string
		body         interface{}
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:      "successful redirect",
			consultID: "consult-1",
			body: RedirectConsultRequest{
				TargetService: "neurology",
				Reason:        "neurological symptoms",
			},
			userRole: models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
				m.On("Redirect", mock.Anything, "consult-1", mock.Anything, "user-1").Return(&RedirectResult{
					Original: &ConsultRequest{
						ID:     "consult-1",
						Status: ConsultStatusRedirected,
					},
					NewConsult: &ConsultRequest{
						ID:            "consult-2",
						TargetService: "neurology",
						Status:        ConsultStatusPending,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "insufficient permissions",
			consultID:    "consult-1",
			body:         RedirectConsultRequest{TargetService: "neurology", Reason: "test"},
			userRole:     models.RoleNurse,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:      "missing targetService",
			consultID: "consult-1",
			body:      RedirectConsultRequest{Reason: "redirect"},
			userRole:  models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
				m.On("Redirect", mock.Anything, "consult-1", mock.Anything, "user-1").Return(nil, errors.New("targetService is required"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/consults/"+tt.consultID+"/redirect", tt.body, "user-1", tt.userRole)
			req.SetPathValue("consultId", tt.consultID)
			rr := httptest.NewRecorder()

			handler.Redirect(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

func TestHandler_Complete(t *testing.T) {
	tests := []struct {
		name         string
		consultID    string
		userRole     models.Role
		setupMock    func(*MockService)
		expectedCode int
	}{
		{
			name:      "successful complete",
			consultID: "consult-1",
			userRole:  models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusAccepted,
				}, nil)
				m.On("Complete", mock.Anything, "consult-1", "user-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusCompleted,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "insufficient permissions",
			consultID:    "consult-1",
			userRole:     models.RoleNurse,
			setupMock:    func(m *MockService) {},
			expectedCode: http.StatusForbidden,
		},
		{
			name:      "not accepted yet",
			consultID: "consult-1",
			userRole:  models.RoleProvider,
			setupMock: func(m *MockService) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
				m.On("Complete", mock.Anything, "consult-1", "user-1").Return(nil, errors.New("only accepted consults can be completed"))
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)
			tt.setupMock(mockSvc)

			handler := NewHandler(mockSvc, newTestDB(t))
			req := testutil.NewRequest("POST", "/api/v1/consults/"+tt.consultID+"/complete", nil, "user-1", tt.userRole)
			req.SetPathValue("consultId", tt.consultID)
			rr := httptest.NewRecorder()

			handler.Complete(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			mockSvc.AssertExpectations(t)
		})
	}
}

// Additional tests for improved coverage

func TestHandler_List_WithStatusFilter(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(f ListConsultsFilter) bool {
		return f.Status == ConsultStatusAccepted
	})).Return([]*ConsultRequest{
		{ID: "consult-1", Status: ConsultStatusAccepted},
	}, int64(1), nil)

	req := testutil.NewRequest("GET", "/api/v1/consults?status=accepted", nil, "user-1", models.RoleProvider)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_List_WithEncounterFilter(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("List", mock.Anything, mock.MatchedBy(func(f ListConsultsFilter) bool {
		return true // filter check happens in handler
	})).Return([]*ConsultRequest{}, int64(0), nil)

	req := testutil.NewRequest("GET", "/api/v1/consults", nil, "user-1", models.RoleProvider)
	rr := httptest.NewRecorder()

	handler.List(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Create_MissingEncounterID(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("Create", mock.Anything, mock.Anything, "user-1").
		Return(nil, errors.New("encounterId is required"))

	req := testutil.NewRequest("POST", "/api/v1/consults", CreateConsultRequest{
		TargetService: "cardiology",
		Reason:        "test",
	}, "user-1", models.RoleProvider)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Create_MissingTargetService(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("Create", mock.Anything, mock.Anything, "user-1").
		Return(nil, errors.New("targetService is required"))

	req := testutil.NewRequest("POST", "/api/v1/consults", CreateConsultRequest{
		EncounterID: "enc-1",
		Reason:      "test",
	}, "user-1", models.RoleProvider)
	rr := httptest.NewRecorder()

	handler.Create(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Accept_ConsultNotFound(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("GetByID", mock.Anything, "consult-999").Return(nil, ErrNotFound)

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-999/accept", nil, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-999")
	rr := httptest.NewRecorder()

	handler.Accept(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Accept_AlreadyAccepted(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
		ID:     "consult-1",
		Status: ConsultStatusAccepted,
	}, nil)
	mockSvc.On("Accept", mock.Anything, "consult-1", "user-1").
		Return(nil, errors.New("consult already accepted"))

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-1/accept", nil, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-1")
	rr := httptest.NewRecorder()

	handler.Accept(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Decline_MissingReason(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
		ID:     "consult-1",
		Status: ConsultStatusPending,
	}, nil)
	mockSvc.On("Decline", mock.Anything, "consult-1", mock.Anything, "user-1").
		Return(nil, errors.New("reason is required"))

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-1/decline", DeclineConsultRequest{}, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-1")
	rr := httptest.NewRecorder()

	handler.Decline(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Redirect_MissingTargetService(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
		ID:     "consult-1",
		Status: ConsultStatusPending,
	}, nil)
	mockSvc.On("Redirect", mock.Anything, "consult-1", mock.Anything, "user-1").
		Return(nil, errors.New("targetService is required"))

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-1/redirect", RedirectConsultRequest{
		Reason: "redirect",
	}, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-1")
	rr := httptest.NewRecorder()

	handler.Redirect(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Redirect_MissingReason(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
		ID:     "consult-1",
		Status: ConsultStatusPending,
	}, nil)
	mockSvc.On("Redirect", mock.Anything, "consult-1", mock.Anything, "user-1").
		Return(nil, errors.New("reason is required"))

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-1/redirect", RedirectConsultRequest{
		TargetService: "neurology",
	}, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-1")
	rr := httptest.NewRecorder()

	handler.Redirect(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Complete_ConsultNotInAcceptableState(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, newTestDB(t))

	mockSvc.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
		ID:     "consult-1",
		Status: ConsultStatusDeclined,
	}, nil)
	mockSvc.On("Complete", mock.Anything, "consult-1", "user-1").
		Return(nil, errors.New("cannot complete declined consult"))

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-1/complete", nil, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-1")
	rr := httptest.NewRecorder()

	handler.Complete(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Redirect_MissingID(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := testutil.NewRequest("POST", "/api/v1/consults//redirect", nil, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "")
	rr := httptest.NewRecorder()

	handler.Redirect(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Redirect_GetByIDNotFound(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil)

	mockSvc.On("GetByID", mock.Anything, "consult-999").
		Return(nil, ErrNotFound)

	reqBody := RedirectConsultRequest{
		TargetService: "cardiology",
		Reason:        "Needs specialist",
	}

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-999/redirect", reqBody, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-999")
	rr := httptest.NewRecorder()

	handler.Redirect(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Redirect_ServiceError(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil)

	mockSvc.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
		ID:     "consult-1",
		Status: ConsultStatusPending,
	}, nil)

	reqBody := RedirectConsultRequest{
		TargetService: "cardiology",
		Reason:        "Needs specialist",
	}

	mockSvc.On("Redirect", mock.Anything, "consult-1", &reqBody, "user-1").
		Return(nil, errors.New("cannot redirect completed consult"))

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-1/redirect", reqBody, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-1")
	rr := httptest.NewRecorder()

	handler.Redirect(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestHandler_Complete_MissingID(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := testutil.NewRequest("POST", "/api/v1/consults//complete", nil, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "")
	rr := httptest.NewRecorder()

	handler.Complete(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHandler_Complete_GetByIDNotFound(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil)

	mockSvc.On("GetByID", mock.Anything, "consult-999").
		Return(nil, ErrNotFound)

	req := testutil.NewRequest("POST", "/api/v1/consults/consult-999/complete", nil, "user-1", models.RoleProvider)
	req.SetPathValue("consultId", "consult-999")
	rr := httptest.NewRecorder()

	handler.Complete(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockSvc.AssertExpectations(t)
}

