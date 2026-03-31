package bed

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) ListBeds(ctx context.Context, filter ListBedsFilter) ([]Bed, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]Bed), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) CreateBed(ctx context.Context, bed *Bed) error {
	args := m.Called(ctx, bed)
	return args.Error(0)
}

func (m *MockRepository) GetBedByID(ctx context.Context, id string) (*Bed, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Bed), args.Error(1)
}

func (m *MockRepository) CreateBedStatusEvent(ctx context.Context, event *BedStatusEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockRepository) UpdateBedFields(ctx context.Context, bedID string, updates map[string]any) error {
	args := m.Called(ctx, bedID, updates)
	return args.Error(0)
}

func (m *MockRepository) CreateBedRequest(ctx context.Context, req *BedRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRepository) GetBedRequestByID(ctx context.Context, id string) (*BedRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*BedRequest), args.Error(1)
}

func (m *MockRepository) UpdateBedRequestFields(ctx context.Context, requestID string, updates map[string]any) error {
	args := m.Called(ctx, requestID, updates)
	return args.Error(0)
}

func (m *MockRepository) AssignBed(ctx context.Context, requestID, bedID, encounterID, userID string, fromStatus BedStatus) error {
	args := m.Called(ctx, requestID, bedID, encounterID, userID, fromStatus)
	return args.Error(0)
}

func TestService_ListBeds(t *testing.T) {
	tests := []struct {
		name      string
		filter    ListBedsFilter
		setupMock func(*MockRepository)
		wantTotal int64
		wantCount int
		wantErr   bool
	}{
		{
			name: "successful list",
			filter: ListBedsFilter{
				UnitID: "unit-1",
				Status: "available",
				Limit:  20,
				Offset: 0,
			},
			setupMock: func(m *MockRepository) {
				m.On("ListBeds", mock.Anything, mock.MatchedBy(func(f ListBedsFilter) bool {
					return f.UnitID == "unit-1" && f.Status == "available"
				})).Return([]Bed{
					{ID: "bed-1", CurrentStatus: BedStatusAvailable},
					{ID: "bed-2", CurrentStatus: BedStatusAvailable},
				}, int64(2), nil)
			},
			wantTotal: 2,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:   "empty list",
			filter: ListBedsFilter{},
			setupMock: func(m *MockRepository) {
				m.On("ListBeds", mock.Anything, mock.Anything).Return([]Bed{}, int64(0), nil)
			},
			wantTotal: 0,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:   "repository error",
			filter: ListBedsFilter{},
			setupMock: func(m *MockRepository) {
				m.On("ListBeds", mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo, nil)
			result, total, err := svc.ListBeds(context.Background(), tt.filter)

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

func TestService_CreateBed(t *testing.T) {
	tests := []struct {
		name      string
		req       CreateBedRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *Bed)
	}{
		{
			name: "successful creation",
			req: CreateBedRequest{
				UnitID:       "unit-1",
				Room:         "101",
				Label:        "Bed A",
				Capabilities: []string{"telemetry", "isolation"},
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("CreateBed", mock.Anything, mock.MatchedBy(func(b *Bed) bool {
					return b.UnitID == "unit-1" &&
						b.Room == "101" &&
						b.Label == "Bed A" &&
						b.CurrentStatus == BedStatusAvailable &&
						len(b.Capabilities) == 2
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, b *Bed) {
				assert.Equal(t, "unit-1", b.UnitID)
				assert.Equal(t, "101", b.Room)
				assert.Equal(t, "Bed A", b.Label)
				assert.Equal(t, BedStatusAvailable, b.CurrentStatus)
				assert.Len(t, b.Capabilities, 2)
			},
		},
		{
			name: "missing unitId",
			req: CreateBedRequest{
				Room:  "101",
				Label: "Bed A",
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "unitId, room, and label are required",
		},
		{
			name: "missing room",
			req: CreateBedRequest{
				UnitID: "unit-1",
				Label:  "Bed A",
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "unitId, room, and label are required",
		},
		{
			name: "missing label",
			req: CreateBedRequest{
				UnitID: "unit-1",
				Room:   "101",
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "unitId, room, and label are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo, nil)
			result, err := svc.CreateBed(context.Background(), tt.req, tt.userID)

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

func TestService_GetBed(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(*MockRepository)
		wantErr   bool
	}{
		{
			name: "successful get",
			id:   "bed-1",
			setupMock: func(m *MockRepository) {
				m.On("GetBedByID", mock.Anything, "bed-1").Return(&Bed{
					ID:            "bed-1",
					UnitID:        "unit-1",
					CurrentStatus: BedStatusAvailable,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   "bed-999",
			setupMock: func(m *MockRepository) {
				m.On("GetBedByID", mock.Anything, "bed-999").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo, nil)
			result, err := svc.GetBed(context.Background(), tt.id)

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

func TestService_UpdateBedStatus(t *testing.T) {
	tests := []struct {
		name      string
		bedID     string
		req       UpdateBedStatusRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *BedStatusEvent)
	}{
		{
			name:  "successful status change",
			bedID: "bed-1",
			req: UpdateBedStatusRequest{
				Status: BedStatusCleaning,
				Reason: strPtr("routine cleaning"),
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetBedByID", mock.Anything, "bed-1").Return(&Bed{
					ID:            "bed-1",
					CurrentStatus: BedStatusAvailable,
				}, nil)
				m.On("CreateBedStatusEvent", mock.Anything, mock.MatchedBy(func(e *BedStatusEvent) bool {
					return e.BedID == "bed-1" &&
						e.FromStatus != nil &&
						*e.FromStatus == BedStatusAvailable &&
						e.ToStatus == BedStatusCleaning &&
						e.ChangedBy == "user-1"
				})).Return(nil)
				m.On("UpdateBedFields", mock.Anything, "bed-1", mock.MatchedBy(func(updates map[string]any) bool {
					return updates["current_status"] == BedStatusCleaning
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, e *BedStatusEvent) {
				assert.Equal(t, "bed-1", e.BedID)
				assert.Equal(t, BedStatusCleaning, e.ToStatus)
				assert.NotNil(t, e.FromStatus)
				assert.Equal(t, BedStatusAvailable, *e.FromStatus)
			},
		},
		{
			name:  "missing status",
			bedID: "bed-1",
			req: UpdateBedStatusRequest{
				Status: "",
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "status is required",
		},
		{
			name:  "bed not found",
			bedID: "bed-999",
			req: UpdateBedStatusRequest{
				Status: BedStatusCleaning,
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetBedByID", mock.Anything, "bed-999").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo, nil)
			result, err := svc.UpdateBedStatus(context.Background(), tt.bedID, tt.req, tt.userID)

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

func TestService_CreateBedRequest(t *testing.T) {
	tests := []struct {
		name        string
		encounterID string
		req         CreateBedRequestRequest
		userID      string
		setupMock   func(*MockRepository)
		wantErr     bool
		validate    func(*testing.T, *BedRequest)
	}{
		{
			name:        "successful creation with priority",
			encounterID: "enc-1",
			req: CreateBedRequestRequest{
				RequiredCapabilities: []string{"telemetry"},
				Priority:             "urgent",
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("CreateBedRequest", mock.Anything, mock.MatchedBy(func(r *BedRequest) bool {
					return r.EncounterID == "enc-1" &&
						r.Priority == "urgent" &&
						r.Status == BedRequestStatusPending &&
						r.CreatedBy == "user-1"
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, r *BedRequest) {
				assert.Equal(t, "enc-1", r.EncounterID)
				assert.Equal(t, "urgent", r.Priority)
				assert.Equal(t, BedRequestStatusPending, r.Status)
			},
		},
		{
			name:        "default priority to routine",
			encounterID: "enc-1",
			req: CreateBedRequestRequest{
				RequiredCapabilities: []string{},
				Priority:             "",
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("CreateBedRequest", mock.Anything, mock.MatchedBy(func(r *BedRequest) bool {
					return r.Priority == "routine"
				})).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, r *BedRequest) {
				assert.Equal(t, "routine", r.Priority)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo, nil)
			result, err := svc.CreateBedRequest(context.Background(), tt.encounterID, tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
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

func TestService_AssignBed(t *testing.T) {
	tests := []struct {
		name      string
		requestID string
		req       AssignBedRequest
		userID    string
		setupMock func(*MockRepository)
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *BedRequest)
	}{
		{
			name:      "successful assignment",
			requestID: "req-1",
			req: AssignBedRequest{
				BedID: "bed-1",
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetBedRequestByID", mock.Anything, "req-1").Return(&BedRequest{
					ID:          "req-1",
					EncounterID: "enc-1",
					Status:      BedRequestStatusPending,
				}, nil).Once()
				m.On("GetBedByID", mock.Anything, "bed-1").Return(&Bed{
					ID:            "bed-1",
					CurrentStatus: BedStatusAvailable,
				}, nil)
				m.On("AssignBed", mock.Anything, "req-1", "bed-1", "enc-1", "user-1", BedStatusAvailable).Return(nil)
				m.On("GetBedRequestByID", mock.Anything, "req-1").Return(&BedRequest{
					ID:            "req-1",
					EncounterID:   "enc-1",
					Status:        BedRequestStatusAssigned,
					AssignedBedID: strPtr("bed-1"),
				}, nil).Once()
			},
			wantErr: false,
			validate: func(t *testing.T, r *BedRequest) {
				assert.Equal(t, BedRequestStatusAssigned, r.Status)
				assert.NotNil(t, r.AssignedBedID)
				assert.Equal(t, "bed-1", *r.AssignedBedID)
			},
		},
		{
			name:      "missing bedId",
			requestID: "req-1",
			req: AssignBedRequest{
				BedID: "",
			},
			userID:    "user-1",
			setupMock: func(m *MockRepository) {},
			wantErr:   true,
			errMsg:    "bedId is required",
		},
		{
			name:      "request not pending",
			requestID: "req-1",
			req: AssignBedRequest{
				BedID: "bed-1",
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetBedRequestByID", mock.Anything, "req-1").Return(&BedRequest{
					ID:     "req-1",
					Status: BedRequestStatusAssigned,
				}, nil)
			},
			wantErr: true,
			errMsg:  "bed request is not pending",
		},
		{
			name:      "bed not available",
			requestID: "req-1",
			req: AssignBedRequest{
				BedID: "bed-1",
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetBedRequestByID", mock.Anything, "req-1").Return(&BedRequest{
					ID:          "req-1",
					EncounterID: "enc-1",
					Status:      BedRequestStatusPending,
				}, nil)
				m.On("GetBedByID", mock.Anything, "bed-1").Return(&Bed{
					ID:            "bed-1",
					CurrentStatus: BedStatusOccupied,
				}, nil)
			},
			wantErr: true,
			errMsg:  "bed is not available for assignment",
		},
		{
			name:      "bed request not found",
			requestID: "req-999",
			req: AssignBedRequest{
				BedID: "bed-1",
			},
			userID: "user-1",
			setupMock: func(m *MockRepository) {
				m.On("GetBedRequestByID", mock.Anything, "req-999").Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			svc := NewService(mockRepo, nil)
			result, err := svc.AssignBed(context.Background(), tt.requestID, tt.req, tt.userID)

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

// Helper function for pointer to string
func strPtr(s string) *string {
	return &s
}
