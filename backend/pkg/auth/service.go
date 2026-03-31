package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrEmailExists       = errors.New("email already exists")
	ErrUserInactive      = errors.New("user account is inactive")
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error
	DeactivateUser(ctx context.Context, userID string) error
}

// Service handles authentication operations
type Service struct {
	db         *database.DB
	jwtService TokenService
}

// NewService creates a new auth service
func NewService(db *database.DB, jwtService TokenService) AuthService {
	return &Service{
		db:         db,
		jwtService: jwtService,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if email already exists
	var existingUser models.User
	err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		return nil, ErrEmailExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		Name:          req.Name,
		Role:          req.Role,
		UnitIDs:       req.UnitIDs,
		DepartmentIDs: req.DepartmentIDs,
		IsActive:      true,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Find user by email
	var user models.User
	err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidPassword
	}

	// Generate JWT token
	token, expiresAt, err := s.jwtService.GenerateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user.ToUserInfo(),
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(ctx context.Context, userID string, req *models.ChangePasswordRequest) error {
	// Get user
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.db.WithContext(ctx).Model(user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeactivateUser deactivates a user account
func (s *Service) DeactivateUser(ctx context.Context, userID string) error {
	result := s.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_active", false)
	
	if result.Error != nil {
		return fmt.Errorf("failed to deactivate user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	
	return nil
}
