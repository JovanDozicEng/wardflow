package consult

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, c *ConsultRequest) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*ConsultRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ConsultRequest), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, f ListConsultsFilter) ([]*ConsultRequest, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*ConsultRequest), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) Update(ctx context.Context, c *ConsultRequest) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name      string
		req       *CreateConsultRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *ConsultRequest)
	}{
		{
			name: "successful creation",
			req: &CreateConsultRequest{
				EncounterID:   "enc-1",
				TargetService: "cardiology",
				Reason:        "chest pain evaluation",
				Urgency:       ConsultUrgencyUrgent,
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(c *ConsultRequest) bool {
					return c.EncounterID == "enc-1" &&
						c.TargetService == "cardiology" &&
						c.Reason == "chest pain evaluation" &&
						c.Urgency == ConsultUrgencyUrgent &&
						c.Status == ConsultStatusPending &&
						c.CreatedBy == "user-1"
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, c *ConsultRequest) {
				assert.Equal(t, "enc-1", c.EncounterID)
				assert.Equal(t, ConsultStatusPending, c.Status)
				assert.Equal(t, "user-1", c.CreatedBy)
			},
		},
		{
			name: "missing encounterId",
			req: &CreateConsultRequest{
				TargetService: "cardiology",
				Reason:        "chest pain",
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "encounterId is required",
		},
		{
			name: "missing targetService",
			req: &CreateConsultRequest{
				EncounterID: "enc-1",
				Reason:      "chest pain",
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "targetService is required",
		},
		{
			name: "missing reason",
			req: &CreateConsultRequest{
				EncounterID:   "enc-1",
				TargetService: "cardiology",
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "reason is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.Create(context.Background(), tt.req, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Accept(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		id        string
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *ConsultRequest)
	}{
		{
			name:   "successful accept",
			id:     "consult-1",
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:            "consult-1",
					EncounterID:   "enc-1",
					TargetService: "cardiology",
					Status:        ConsultStatusPending,
					CreatedBy:     "user-1",
					CreatedAt:     now,
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(c *ConsultRequest) bool {
					return c.ID == "consult-1" &&
						c.Status == ConsultStatusAccepted &&
						c.AcceptedBy != nil &&
						*c.AcceptedBy == "provider-1" &&
						c.AcceptedAt != nil
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, c *ConsultRequest) {
				assert.Equal(t, ConsultStatusAccepted, c.Status)
				assert.NotNil(t, c.AcceptedBy)
				assert.Equal(t, "provider-1", *c.AcceptedBy)
				assert.NotNil(t, c.AcceptedAt)
			},
		},
		{
			name:   "already accepted",
			id:     "consult-1",
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				acceptedBy := "provider-2"
				acceptedAt := now
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:         "consult-1",
					Status:     ConsultStatusAccepted,
					AcceptedBy: &acceptedBy,
					AcceptedAt: &acceptedAt,
				}, nil)
			},
			wantErr: true,
			errMsg:  "only pending consults can be accepted",
		},
		{
			name:   "consult not found",
			id:     "consult-999",
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-999").Return(nil, ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.Accept(context.Background(), tt.id, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Decline(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		id        string
		req       *DeclineConsultRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *ConsultRequest)
	}{
		{
			name:   "successful decline",
			id:     "consult-1",
			req:    &DeclineConsultRequest{Reason: "patient already seen"},
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:            "consult-1",
					EncounterID:   "enc-1",
					TargetService: "cardiology",
					Status:        ConsultStatusPending,
					CreatedBy:     "user-1",
					CreatedAt:     now,
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(c *ConsultRequest) bool {
					return c.ID == "consult-1" &&
						c.Status == ConsultStatusDeclined &&
						c.CloseReason != nil &&
						*c.CloseReason == "patient already seen" &&
						c.ClosedAt != nil
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, c *ConsultRequest) {
				assert.Equal(t, ConsultStatusDeclined, c.Status)
				assert.NotNil(t, c.CloseReason)
				assert.Equal(t, "patient already seen", *c.CloseReason)
				assert.NotNil(t, c.ClosedAt)
			},
		},
		{
			name:   "missing reason",
			id:     "consult-1",
			req:    &DeclineConsultRequest{Reason: ""},
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
			},
			wantErr: true,
			errMsg:  "reason is required",
		},
		{
			name:   "wrong status",
			id:     "consult-1",
			req:    &DeclineConsultRequest{Reason: "duplicate request"},
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusAccepted,
				}, nil)
			},
			wantErr: true,
			errMsg:  "only pending consults can be declined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.Decline(context.Background(), tt.id, tt.req, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Redirect(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		id        string
		req       *RedirectConsultRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *RedirectResult)
	}{
		{
			name: "successful redirect",
			id:   "consult-1",
			req: &RedirectConsultRequest{
				TargetService: "neurology",
				Reason:        "neurological symptoms",
			},
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:            "consult-1",
					EncounterID:   "enc-1",
					TargetService: "cardiology",
					Reason:        "chest pain",
					Urgency:       ConsultUrgencyUrgent,
					Status:        ConsultStatusPending,
					CreatedBy:     "user-1",
					CreatedAt:     now,
				}, nil)
				// Update original to redirected
				m.On("Update", mock.Anything, mock.MatchedBy(func(c *ConsultRequest) bool {
					return c.ID == "consult-1" &&
						c.Status == ConsultStatusRedirected &&
						c.RedirectedTo != nil &&
						*c.RedirectedTo == "neurology" &&
						c.CloseReason != nil &&
						*c.CloseReason == "neurological symptoms" &&
						c.ClosedAt != nil
				})).Return(nil)
				// Create new consult
				m.On("Create", mock.Anything, mock.MatchedBy(func(c *ConsultRequest) bool {
					return c.EncounterID == "enc-1" &&
						c.TargetService == "neurology" &&
						c.Reason == "chest pain" &&
						c.Urgency == ConsultUrgencyUrgent &&
						c.Status == ConsultStatusPending &&
						c.CreatedBy == "provider-1"
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, r *RedirectResult) {
				assert.NotNil(t, r.Original)
				assert.Equal(t, ConsultStatusRedirected, r.Original.Status)
				assert.NotNil(t, r.Original.RedirectedTo)
				assert.Equal(t, "neurology", *r.Original.RedirectedTo)
				
				assert.NotNil(t, r.NewConsult)
				assert.Equal(t, ConsultStatusPending, r.NewConsult.Status)
				assert.Equal(t, "neurology", r.NewConsult.TargetService)
				assert.Equal(t, "enc-1", r.NewConsult.EncounterID)
			},
		},
		{
			name: "missing reason",
			id:   "consult-1",
			req: &RedirectConsultRequest{
				TargetService: "neurology",
				Reason:        "",
			},
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
			},
			wantErr: true,
			errMsg:  "reason is required",
		},
		{
			name: "missing targetService",
			id:   "consult-1",
			req: &RedirectConsultRequest{
				TargetService: "",
				Reason:        "redirect",
			},
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
			},
			wantErr: true,
			errMsg:  "targetService is required",
		},
		{
			name: "already accepted",
			id:   "consult-1",
			req: &RedirectConsultRequest{
				TargetService: "neurology",
				Reason:        "redirect",
			},
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusAccepted,
				}, nil)
			},
			wantErr: true,
			errMsg:  "only pending consults can be redirected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.Redirect(context.Background(), tt.id, tt.req, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Complete(t *testing.T) {
	now := time.Now().UTC()
	acceptedBy := "provider-1"

	tests := []struct {
		name      string
		id        string
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *ConsultRequest)
	}{
		{
			name:   "successful complete",
			id:     "consult-1",
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:         "consult-1",
					Status:     ConsultStatusAccepted,
					AcceptedBy: &acceptedBy,
					AcceptedAt: &now,
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(c *ConsultRequest) bool {
					return c.ID == "consult-1" &&
						c.Status == ConsultStatusCompleted &&
						c.ClosedAt != nil
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, c *ConsultRequest) {
				assert.Equal(t, ConsultStatusCompleted, c.Status)
				assert.NotNil(t, c.ClosedAt)
			},
		},
		{
			name:   "not accepted yet",
			id:     "consult-1",
			userID: "provider-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
			},
			wantErr: true,
			errMsg:  "only accepted consults can be completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.Complete(context.Background(), tt.id, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name      string
		filter    ListConsultsFilter
		setupMock func(*MockRepository)
		wantTotal int64
		wantCount int
		wantErr   bool
	}{
		{
			name: "successful list",
			filter: ListConsultsFilter{
				Status: ConsultStatusPending,
				Limit:  20,
				Offset: 0,
			},
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f ListConsultsFilter) bool {
					return f.Status == ConsultStatusPending
				})).Return([]*ConsultRequest{
					{ID: "consult-1", Status: ConsultStatusPending},
					{ID: "consult-2", Status: ConsultStatusPending},
				}, int64(2), nil)
			},
			wantTotal: 2,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "empty list",
			filter: ListConsultsFilter{},
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, mock.Anything).Return([]*ConsultRequest{}, int64(0), nil)
			},
			wantTotal: 0,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "repository error",
			filter: ListConsultsFilter{},
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, total, err := svc.List(context.Background(), tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTotal, total)
				assert.Len(t, result, tt.wantCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(*MockRepository)
		wantErr   bool
	}{
		{
			name: "successful get",
			id:   "consult-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-1").Return(&ConsultRequest{
					ID:     "consult-1",
					Status: ConsultStatusPending,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   "consult-999",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "consult-999").Return(nil, ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
