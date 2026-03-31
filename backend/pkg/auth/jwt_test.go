package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wardflow/backend/internal/models"
)

func TestJWTService_GenerateToken(t *testing.T) {
	secret := "test-secret-key-for-jwt-testing"
	jwtService := NewJWTService(secret, 24)

	t.Run("generates valid token", func(t *testing.T) {
		user := &models.User{
			ID:            "user-123",
			Email:         "test@example.com",
			Role:          models.RoleNurse,
			UnitIDs:       models.StringArray{"unit-1", "unit-2"},
			DepartmentIDs: models.StringArray{"dept-1"},
		}

		token, expiresAt, err := jwtService.GenerateToken(user)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Greater(t, expiresAt, time.Now().Unix())
	})

	t.Run("generates token with correct claims", func(t *testing.T) {
		user := &models.User{
			ID:    "user-456",
			Email: "admin@example.com",
			Role:  models.RoleAdmin,
		}

		token, _, err := jwtService.GenerateToken(user)
		assert.NoError(t, err)

		// Validate and check claims
		claims, err := jwtService.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Role, claims.Role)
	})
}

func TestJWTService_ValidateToken(t *testing.T) {
	secret := "test-secret-key-for-jwt-testing"
	jwtService := NewJWTService(secret, 24)

	t.Run("validates valid token", func(t *testing.T) {
		user := &models.User{
			ID:    "user-123",
			Email: "test@example.com",
			Role:  models.RoleNurse,
		}

		token, _, err := jwtService.GenerateToken(user)
		assert.NoError(t, err)

		claims, err := jwtService.ValidateToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Email, claims.Email)
		assert.Equal(t, user.Role, claims.Role)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		// Create service with very short expiration
		shortService := NewJWTService(secret, 0)

		user := &models.User{
			ID:    "user-123",
			Email: "test@example.com",
			Role:  models.RoleNurse,
		}

		token, _, err := shortService.GenerateToken(user)
		assert.NoError(t, err)

		// Wait for token to expire
		time.Sleep(2 * time.Second)

		claims, err := shortService.ValidateToken(token)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrExpiredToken)
		assert.Nil(t, claims)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		claims, err := jwtService.ValidateToken("invalid.token.string")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("rejects token with wrong secret", func(t *testing.T) {
		wrongSecretService := NewJWTService("wrong-secret", 24)

		user := &models.User{
			ID:    "user-123",
			Email: "test@example.com",
			Role:  models.RoleNurse,
		}

		token, _, err := wrongSecretService.GenerateToken(user)
		assert.NoError(t, err)

		// Try to validate with different secret
		claims, err := jwtService.ValidateToken(token)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, claims)
	})

	t.Run("rejects tampered token", func(t *testing.T) {
		user := &models.User{
			ID:    "user-123",
			Email: "test@example.com",
			Role:  models.RoleNurse,
		}

		token, _, err := jwtService.GenerateToken(user)
		assert.NoError(t, err)

		// Tamper with token by modifying a character
		tamperedToken := token[:len(token)-5] + "XXXXX"

		claims, err := jwtService.ValidateToken(tamperedToken)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, claims)
	})
}

func TestJWTService_RefreshToken(t *testing.T) {
	secret := "test-secret-key-for-jwt-testing"
	jwtService := NewJWTService(secret, 24)

	t.Run("refreshes valid token", func(t *testing.T) {
		user := &models.User{
			ID:    "user-123",
			Email: "test@example.com",
			Role:  models.RoleNurse,
		}

		token, _, err := jwtService.GenerateToken(user)
		assert.NoError(t, err)

		// Refresh token
		newToken, newExpiresAt, err := jwtService.RefreshToken(token)
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		assert.Greater(t, newExpiresAt, time.Now().Unix())

		// Validate new token
		claims, err := jwtService.ValidateToken(newToken)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
	})

	t.Run("rejects invalid token for refresh", func(t *testing.T) {
		newToken, expiresAt, err := jwtService.RefreshToken("invalid.token.string")
		assert.Error(t, err)
		assert.Empty(t, newToken)
		assert.Equal(t, int64(0), expiresAt)
	})
}

func TestUserContext(t *testing.T) {
	t.Run("SetUserContext and GetUserContext", func(t *testing.T) {
		claims := &Claims{
			UserID:  "user-123",
			Email:   "test@example.com",
			Role:    models.RoleNurse,
			UnitIDs: models.StringArray{"unit-1"},
			DeptIDs: models.StringArray{"dept-1"},
		}

		ctx := context.Background()
		ctx = SetUserContext(ctx, claims)

		userCtx, ok := GetUserContext(ctx)
		assert.True(t, ok)
		assert.NotNil(t, userCtx)
		assert.Equal(t, claims.UserID, userCtx.UserID)
		assert.Equal(t, claims.Email, userCtx.Email)
		assert.Equal(t, claims.Role, userCtx.Role)
		assert.Equal(t, claims.UnitIDs, userCtx.UnitIDs)
		assert.Equal(t, claims.DeptIDs, userCtx.DeptIDs)
	})

	t.Run("GetUserContext returns false when not set", func(t *testing.T) {
		ctx := context.Background()

		userCtx, ok := GetUserContext(ctx)
		assert.False(t, ok)
		assert.Nil(t, userCtx)
	})

	t.Run("MustGetUserContext returns context", func(t *testing.T) {
		claims := &Claims{
			UserID: "user-123",
			Email:  "test@example.com",
			Role:   models.RoleAdmin,
		}

		ctx := context.Background()
		ctx = SetUserContext(ctx, claims)

		userCtx := MustGetUserContext(ctx)
		assert.NotNil(t, userCtx)
		assert.Equal(t, claims.UserID, userCtx.UserID)
	})

	t.Run("MustGetUserContext panics when not set", func(t *testing.T) {
		ctx := context.Background()

		assert.Panics(t, func() {
			MustGetUserContext(ctx)
		})
	})
}
