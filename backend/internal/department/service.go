package department

import (
	"context"
)

// Service defines department business logic operations
type Service interface {
	List(ctx context.Context, q string) ([]Department, error)
	Create(ctx context.Context, req CreateDepartmentRequest) (*Department, error)
	GetByID(ctx context.Context, id string) (*Department, error)
}

type service struct {
	repo Repository
}

// NewService creates a new department service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, q string) ([]Department, error) {
	return s.repo.List(ctx, q)
}

func (s *service) Create(ctx context.Context, req CreateDepartmentRequest) (*Department, error) {
	if req.Name == "" {
		return nil, ErrValidation("name is required")
	}
	if req.Code == "" {
		return nil, ErrValidation("code is required")
	}

	dept := &Department{
		Name: req.Name,
		Code: req.Code,
	}

	if err := s.repo.Create(ctx, dept); err != nil {
		return nil, err
	}

	return dept, nil
}

func (s *service) GetByID(ctx context.Context, id string) (*Department, error) {
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
