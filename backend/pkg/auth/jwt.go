package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wardflow/backend/internal/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents JWT claims
type Claims struct {
	UserID    string      `json:"userId"`
	Email     string      `json:"email"`
	Role      models.Role `json:"role"`
	UnitIDs   models.StringArray `json:"unitIds,omitempty"`
	DeptIDs   models.StringArray `json:"deptIds,omitempty"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey []byte
	expiration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, expirationHours int) *JWTService {
	return &JWTService{
		secretKey:  []byte(secretKey),
		expiration: time.Duration(expirationHours) * time.Hour,
	}
}

// GenerateToken creates a new JWT token for a user
func (s *JWTService) GenerateToken(user *models.User) (string, int64, error) {
	expiresAt := time.Now().Add(s.expiration)
	
	claims := &Claims{
		UserID:  user.ID,
		Email:   user.Email,
		Role:    user.Role,
		UnitIDs: user.UnitIDs,
		DeptIDs: user.DepartmentIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "wardflow-api",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiration
func (s *JWTService) RefreshToken(tokenString string) (string, int64, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		// Allow refresh of expired tokens within a grace period (24 hours)
		if !errors.Is(err, ErrExpiredToken) {
			return "", 0, err
		}
	}

	// Create new token with same claims but new expiration
	expiresAt := time.Now().Add(s.expiration)
	claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	claims.IssuedAt = jwt.NewNumericDate(time.Now())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(s.secretKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt.Unix(), nil
}

// UserContext represents authenticated user in context
type UserContext struct {
	UserID    string
	Email     string
	Role      models.Role
	UnitIDs   models.StringArray
	DeptIDs   models.StringArray
}

type contextKey string

const userContextKey contextKey = "user"

// SetUserContext adds user information to context
func SetUserContext(ctx context.Context, claims *Claims) context.Context {
	userCtx := &UserContext{
		UserID:  claims.UserID,
		Email:   claims.Email,
		Role:    claims.Role,
		UnitIDs: claims.UnitIDs,
		DeptIDs: claims.DeptIDs,
	}
	return context.WithValue(ctx, userContextKey, userCtx)
}

// GetUserContext retrieves user information from context
func GetUserContext(ctx context.Context) (*UserContext, bool) {
	userCtx, ok := ctx.Value(userContextKey).(*UserContext)
	return userCtx, ok
}

// MustGetUserContext retrieves user context or panics (use after auth middleware)
func MustGetUserContext(ctx context.Context) *UserContext {
	userCtx, ok := GetUserContext(ctx)
	if !ok {
		panic("user context not found - ensure auth middleware is applied")
	}
	return userCtx
}
