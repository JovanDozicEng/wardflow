package unit

import (
	"context"
)

// Service defines unit business logic operations
type Service interface {
	List(ctx context.Context, q, departmentID string) ([]Unit, error)
	Create(ctx context.Context, req CreateUnitRequest) (*Unit, error)
	GetByID(ctx context.Context, id string) (*Unit, error)
}

type service struct {
	repo Repository
}

// NewService creates a new unit service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, q, departmentID string) ([]Unit, error) {
	return s.repo.List(ctx, q, departmentID)
}

func (s *service) Create(ctx context.Context, req CreateUnitRequest) (*Unit, error) {
	if req.Name == "" {
		return nil, ErrValidation("name is required")
	}
	if req.Code == "" {
		return nil, ErrValidation("code is required")
	}
	if req.DepartmentID == "" {
		return nil, ErrValidation("departmentId is required")
	}

	unit := &Unit{
		Name:         req.Name,
		Code:         req.Code,
		DepartmentID: req.DepartmentID,
	}

	if err := s.repo.Create(ctx, unit); err != nil {
		return nil, err
	}

	return unit, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Unit, error) {
	return s.repo.GetByID(ctx, id)
}

// Error helpers
func ErrValidation(msg string) error {
	return &ValidationError{Message: msg}
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
