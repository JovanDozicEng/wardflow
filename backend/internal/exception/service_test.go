package exception

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

func (m *MockRepository) Create(ctx context.Context, e *ExceptionEvent) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*ExceptionEvent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExceptionEvent), args.Error(1)
}

func (m *MockRepository) List(ctx context.Context, f ListExceptionsFilter) ([]*ExceptionEvent, int64, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*ExceptionEvent), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) Update(ctx context.Context, e *ExceptionEvent) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name      string
		req       *CreateExceptionRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *ExceptionEvent)
	}{
		{
			name: "successful creation",
			req: &CreateExceptionRequest{
				EncounterID: "enc-1",
				Type:        "medication-delay",
				Data: map[string]interface{}{
					"medication": "aspirin",
					"delayMinutes": 30,
				},
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(e *ExceptionEvent) bool {
					return e.EncounterID == "enc-1" &&
						e.Type == "medication-delay" &&
						e.Status == ExceptionStatusDraft &&
						e.InitiatedBy == "user-1"
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, e *ExceptionEvent) {
				assert.Equal(t, "enc-1", e.EncounterID)
				assert.Equal(t, "medication-delay", e.Type)
				assert.Equal(t, ExceptionStatusDraft, e.Status)
				assert.Equal(t, "user-1", e.InitiatedBy)
			},
		},
		{
			name: "missing encounterId",
			req: &CreateExceptionRequest{
				Type: "medication-delay",
				Data: map[string]interface{}{},
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "encounterId is required",
		},
		{
			name: "missing type",
			req: &CreateExceptionRequest{
				EncounterID: "enc-1",
				Data:        map[string]interface{}{},
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "type is required",
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

func TestService_Update(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		id        string
		req       *UpdateExceptionRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *ExceptionEvent)
	}{
		{
			name: "successful update",
			id:   "exc-1",
			req: &UpdateExceptionRequest{
				Data: map[string]interface{}{
					"medication": "ibuprofen",
					"delayMinutes": 45,
				},
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:          "exc-1",
					EncounterID: "enc-1",
					Type:        "medication-delay",
					Status:      ExceptionStatusDraft,
					Data:        `{"medication":"aspirin","delayMinutes":30}`,
					InitiatedBy: "user-1",
					InitiatedAt: now,
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(e *ExceptionEvent) bool {
					return e.ID == "exc-1" && e.Status == ExceptionStatusDraft
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, e *ExceptionEvent) {
				assert.Equal(t, "exc-1", e.ID)
				assert.Equal(t, ExceptionStatusDraft, e.Status)
			},
		},
		{
			name: "cannot update finalized",
			id:   "exc-1",
			req: &UpdateExceptionRequest{
				Data: map[string]interface{}{"test": "data"},
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusFinalized,
				}, nil)
			},
			wantErr: true,
			errMsg:  "only draft exceptions can be updated",
		},
		{
			name: "not found",
			id:   "exc-999",
			req: &UpdateExceptionRequest{
				Data: map[string]interface{}{},
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-999").Return(nil, ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.Update(context.Background(), tt.id, tt.req, tt.userID)

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

func TestService_Finalize(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name      string
		id        string
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *ExceptionEvent)
	}{
		{
			name:   "successful finalize",
			id:     "exc-1",
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:          "exc-1",
					EncounterID: "enc-1",
					Type:        "medication-delay",
					Status:      ExceptionStatusDraft,
					InitiatedBy: "user-2",
					InitiatedAt: now,
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(e *ExceptionEvent) bool {
					return e.ID == "exc-1" &&
						e.Status == ExceptionStatusFinalized &&
						e.FinalizedBy != nil &&
						*e.FinalizedBy == "user-1" &&
						e.FinalizedAt != nil
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, e *ExceptionEvent) {
				assert.Equal(t, ExceptionStatusFinalized, e.Status)
				assert.NotNil(t, e.FinalizedBy)
				assert.Equal(t, "user-1", *e.FinalizedBy)
				assert.NotNil(t, e.FinalizedAt)
			},
		},
		{
			name:   "already finalized",
			id:     "exc-1",
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				finalizedBy := "user-2"
				finalizedAt := now
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:          "exc-1",
					Status:      ExceptionStatusFinalized,
					FinalizedBy: &finalizedBy,
					FinalizedAt: &finalizedAt,
				}, nil)
			},
			wantErr: true,
			errMsg:  "only draft exceptions can be finalized",
		},
		{
			name:   "not found",
			id:     "exc-999",
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-999").Return(nil, ErrNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.Finalize(context.Background(), tt.id, tt.userID)

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

func TestService_Correct(t *testing.T) {
	now := time.Now().UTC()
	finalizedBy := "user-1"

	tests := []struct {
		name      string
		id        string
		req       *CorrectExceptionRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *ExceptionEvent)
	}{
		{
			name: "successful correction",
			id:   "exc-1",
			req: &CorrectExceptionRequest{
				Reason: "corrected dosage",
				Data: map[string]interface{}{
					"medication": "aspirin",
					"dosage": "100mg",
				},
			},
			userID: "user-2",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:             "exc-1",
					EncounterID:    "enc-1",
					Type:           "medication-delay",
					Status:         ExceptionStatusFinalized,
					RequiredFields: "{}",
					Data:           `{"medication":"aspirin","dosage":"50mg"}`,
					InitiatedBy:    "user-1",
					InitiatedAt:    now,
					FinalizedBy:    &finalizedBy,
					FinalizedAt:    &now,
				}, nil)
				// New exception is created
				m.On("Create", mock.Anything, mock.MatchedBy(func(e *ExceptionEvent) bool {
					return e.EncounterID == "enc-1" &&
						e.Type == "medication-delay" &&
						e.Status == ExceptionStatusFinalized &&
						e.InitiatedBy == "user-1" &&
						e.FinalizedBy != nil &&
						*e.FinalizedBy == "user-2"
				})).Return(nil)
				// Original exception is marked as corrected
				m.On("Update", mock.Anything, mock.MatchedBy(func(e *ExceptionEvent) bool {
					return e.ID == "exc-1" &&
						e.Status == ExceptionStatusCorrected &&
						e.CorrectedByEventID != nil &&
						e.CorrectionReason != nil &&
						*e.CorrectionReason == "corrected dosage"
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, e *ExceptionEvent) {
				assert.Equal(t, ExceptionStatusFinalized, e.Status)
				assert.Equal(t, "user-1", e.InitiatedBy)
				assert.NotNil(t, e.FinalizedBy)
				assert.Equal(t, "user-2", *e.FinalizedBy)
			},
		},
		{
			name: "cannot correct draft",
			id:   "exc-1",
			req: &CorrectExceptionRequest{
				Reason: "test",
				Data:   map[string]interface{}{},
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
			},
			wantErr: true,
			errMsg:  "only finalized exceptions can be corrected",
		},
		{
			name: "missing reason",
			id:   "exc-1",
			req: &CorrectExceptionRequest{
				Reason: "",
				Data:   map[string]interface{}{},
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:          "exc-1",
					Status:      ExceptionStatusFinalized,
					FinalizedBy: &finalizedBy,
					FinalizedAt: &now,
				}, nil)
			},
			wantErr: true,
			errMsg:  "reason is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo)
			result, err := svc.Correct(context.Background(), tt.id, tt.req, tt.userID)

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
		filter    ListExceptionsFilter
		setupMock func(*MockRepository)
		wantTotal int64
		wantCount int
		wantErr   bool
	}{
		{
			name: "successful list",
			filter: ListExceptionsFilter{
				Status: ExceptionStatusDraft,
				Limit:  20,
				Offset: 0,
			},
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, mock.MatchedBy(func(f ListExceptionsFilter) bool {
					return f.Status == ExceptionStatusDraft
				})).Return([]*ExceptionEvent{
					{ID: "exc-1", Status: ExceptionStatusDraft},
					{ID: "exc-2", Status: ExceptionStatusDraft},
				}, int64(2), nil)
			},
			wantTotal: 2,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "empty list",
			filter: ListExceptionsFilter{},
			setupMock: func(m *MockRepository) {
				m.On("List", mock.Anything, mock.Anything).Return([]*ExceptionEvent{}, int64(0), nil)
			},
			wantTotal: 0,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "repository error",
			filter: ListExceptionsFilter{},
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
			id:   "exc-1",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-1").Return(&ExceptionEvent{
					ID:     "exc-1",
					Status: ExceptionStatusDraft,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   "exc-999",
			setupMock: func(m *MockRepository) {
				m.On("GetByID", mock.Anything, "exc-999").Return(nil, ErrNotFound)
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
