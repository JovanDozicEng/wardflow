package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/wardflow/backend/internal/models"
	"github.com/wardflow/backend/pkg/database"
)

// Mock TokenService for testing
type mockTokenService struct {
	mock.Mock
}

func (m *mockTokenService) GenerateToken(user *models.User) (string, int64, error) {
	args := m.Called(user)
	return args.String(0), args.Get(1).(int64), args.Error(2)
}

func (m *mockTokenService) ValidateToken(tokenString string) (*Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Claims), args.Error(1)
}

func (m *mockTokenService) RefreshToken(tokenString string) (string, int64, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Get(1).(int64), args.Error(2)
}

// Helper functions for testing

// newAuthTestDB creates an in-memory SQLite database for testing
func newAuthTestDB(t *testing.T) *database.DB {
	t.Helper()
	
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Disable foreign keys for SQLite compatibility
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err, "failed to open test database")
	
	// Register a BeforeCreate callback to auto-generate UUIDs for User models in SQLite
	// This simulates PostgreSQL's gen_random_uuid() default
	err = db.Callback().Create().Before("gorm:create").Register("generate_uuid", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil && tx.Statement.Schema.Table == "users" {
			// Check if we're creating a User
			if user, ok := tx.Statement.Dest.(*models.User); ok && user.ID == "" {
				user.ID = uuid.New().String()
			}
		}
	})
	require.NoError(t, err, "failed to register UUID callback")
	
	// Customize the migrator to use SQLite-compatible types
	// We need to use GORM callbacks to override PostgreSQL-specific types
	migrator := db.Migrator()
	
	// AutoMigrate with a custom User struct that's SQLite-compatible
	err = migrator.AutoMigrate(&testUser{})
	require.NoError(t, err, "failed to migrate test database")
	
	return &database.DB{DB: db}
}

// testUser is a SQLite-compatible version of models.User for testing
type testUser struct {
	ID            string              `gorm:"type:text;primaryKey"` // Use text instead of uuid
	Email         string              `gorm:"uniqueIndex;not null"`
	PasswordHash  string              `gorm:"not null"`
	Name          string              `gorm:"not null"`
	IsActive      bool                `gorm:"default:true;not null"`
	Role          models.Role         `gorm:"type:varchar(50);not null;index"`
	UnitIDs       models.StringArray  `gorm:"type:text;default:'[]'"` // Use text instead of jsonb
	DepartmentIDs models.StringArray  `gorm:"type:text;default:'[]'"` // Use text instead of jsonb
	CreatedAt     time.Time           `gorm:"not null"`
	UpdatedAt     time.Time           `gorm:"not null"`
	DeletedAt     gorm.DeletedAt      `gorm:"index"`
}

// TableName ensures testUser maps to the users table
func (testUser) TableName() string {
	return "users"
}

// newTestService creates a new service with test database and mock JWT service
func newTestService(t *testing.T) (AuthService, *database.DB, *mockTokenService) {
	db := newAuthTestDB(t)
	mockJWT := new(mockTokenService)
	svc := NewService(db, mockJWT)
	return svc, db, mockJWT
}

// createTestUser is a helper to insert a user directly into the database
func createTestUser(t *testing.T, db *database.DB, user *models.User) *models.User {
	t.Helper()
	
	// Generate UUID if not set (SQLite doesn't auto-generate)
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	
	// Set timestamps if not set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now().UTC()
	}
	
	// Initialize empty arrays if nil (prevents NULL in database)
	if user.UnitIDs == nil {
		user.UnitIDs = models.StringArray{}
	}
	if user.DepartmentIDs == nil {
		user.DepartmentIDs = models.StringArray{}
	}
	
	// Remember the IsActive value before create
	isActive := user.IsActive
	
	// Create the user
	err := db.Create(user).Error
	require.NoError(t, err, "failed to create test user")
	
	// GORM skips zero values like false for booleans, so we need to explicitly update
	// IsActive if it's false (use Update with a map to force the update)
	if !isActive {
		err = db.Model(user).Update("is_active", false).Error
		require.NoError(t, err, "failed to set user inactive")
	}
	
	return user
}

func TestService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully registers new user", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		req := &models.RegisterRequest{
			Email:         "newuser@example.com",
			Password:      "securepassword123",
			Name:          "New User",
			Role:          models.RoleNurse,
			UnitIDs:       models.StringArray{"unit-1", "unit-2"},
			DepartmentIDs: models.StringArray{"dept-1"},
		}

		user, err := svc.Register(ctx, req)
		
		require.NoError(t, err)
		require.NotNil(t, user)
		
		// Verify user fields
		assert.NotEmpty(t, user.ID, "user ID should be generated")
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, req.Role, user.Role)
		assert.Equal(t, req.UnitIDs, user.UnitIDs)
		assert.Equal(t, req.DepartmentIDs, user.DepartmentIDs)
		assert.True(t, user.IsActive, "new user should be active by default")
		assert.NotEmpty(t, user.CreatedAt)
		assert.NotEmpty(t, user.UpdatedAt)
		
		// Verify password is hashed, not stored plaintext
		assert.NotEqual(t, req.Password, user.PasswordHash, "password should be hashed")
		assert.NotEmpty(t, user.PasswordHash)
		
		// Verify password hash is valid
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		assert.NoError(t, err, "password hash should match original password")
		
		// Verify user was actually saved to database
		var dbUser models.User
		err = db.First(&dbUser, "id = ?", user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, user.Email, dbUser.Email)
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create existing user
		existingUser := &models.User{
			Email:        "existing@example.com",
			PasswordHash: "hashedpassword",
			Name:         "Existing User",
			Role:         models.RoleNurse,
			IsActive:     true,
		}
		createTestUser(t, db, existingUser)

		// Try to register with same email
		req := &models.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
			Name:     "New User",
			Role:     models.RoleProvider,
		}

		user, err := svc.Register(ctx, req)
		
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrEmailExists, err, "should return ErrEmailExists")
	})

	t.Run("password is properly hashed", func(t *testing.T) {
		svc, _, _ := newTestService(t)

		req := &models.RegisterRequest{
			Email:    "test@example.com",
			Password: "mySecurePassword123!",
			Name:     "Test User",
			Role:     models.RoleNurse,
		}

		user, err := svc.Register(ctx, req)
		
		require.NoError(t, err)
		require.NotNil(t, user)
		
		// Password should be hashed with bcrypt
		assert.NotEqual(t, req.Password, user.PasswordHash)
		assert.Greater(t, len(user.PasswordHash), 50, "bcrypt hash should be long")
		
		// Should be able to verify password
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		assert.NoError(t, err)
		
		// Wrong password should fail
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("wrongpassword"))
		assert.Error(t, err)
	})
}

func TestService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully authenticates valid credentials", func(t *testing.T) {
		svc, db, mockJWT := newTestService(t)

		// Create a test user with known password
		password := "securepassword123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Test User",
			Role:         models.RoleNurse,
			IsActive:     true,
		}
		user = createTestUser(t, db, user)

		// Mock JWT generation
		expectedToken := "mock.jwt.token"
		expectedExpiresAt := time.Now().Add(24 * time.Hour).Unix()
		mockJWT.On("GenerateToken", mock.MatchedBy(func(u *models.User) bool {
			return u.ID == user.ID && u.Email == user.Email
		})).Return(expectedToken, expectedExpiresAt, nil)

		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: password,
		}

		resp, err := svc.Login(ctx, req)
		
		require.NoError(t, err)
		require.NotNil(t, resp)
		
		assert.Equal(t, expectedToken, resp.Token)
		assert.Equal(t, expectedExpiresAt, resp.ExpiresAt)
		assert.NotNil(t, resp.User)
		assert.Equal(t, user.ID, resp.User.ID)
		assert.Equal(t, user.Email, resp.User.Email)
		assert.Equal(t, user.Name, resp.User.Name)
		assert.Equal(t, user.Role, resp.User.Role)
		
		mockJWT.AssertExpectations(t)
	})

	t.Run("rejects wrong password", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create user with known password
		correctPassword := "correctpassword"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Test User",
			Role:         models.RoleNurse,
			IsActive:     true,
		}
		createTestUser(t, db, user)

		// Try to login with wrong password
		req := &models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		resp, err := svc.Login(ctx, req)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, ErrInvalidPassword, err, "should return ErrInvalidPassword")
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		svc, _, _ := newTestService(t)

		req := &models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		resp, err := svc.Login(ctx, req)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, ErrUserNotFound, err, "should return ErrUserNotFound")
	})

	t.Run("rejects inactive user", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create inactive user
		password := "password123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			Email:        "inactive@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Inactive User",
			Role:         models.RoleNurse,
			IsActive:     false, // User is inactive
		}
		createTestUser(t, db, user)

		// Try to login with correct credentials
		req := &models.LoginRequest{
			Email:    "inactive@example.com",
			Password: password,
		}

		resp, err := svc.Login(ctx, req)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, ErrUserInactive, err, "should return ErrUserInactive")
	})
}

func TestService_GetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns user by ID", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create test user
		expectedUser := &models.User{
			Email:         "test@example.com",
			PasswordHash:  "hashedpassword",
			Name:          "Test User",
			Role:          models.RoleProvider,
			UnitIDs:       models.StringArray{"unit-1"},
			DepartmentIDs: models.StringArray{"dept-1", "dept-2"},
			IsActive:      true,
		}
		expectedUser = createTestUser(t, db, expectedUser)

		user, err := svc.GetUserByID(ctx, expectedUser.ID)
		
		require.NoError(t, err)
		require.NotNil(t, user)
		
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Role, user.Role)
		assert.Equal(t, expectedUser.UnitIDs, user.UnitIDs)
		assert.Equal(t, expectedUser.DepartmentIDs, user.DepartmentIDs)
		assert.Equal(t, expectedUser.IsActive, user.IsActive)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		svc, _, _ := newTestService(t)

		nonExistentID := uuid.New().String()
		user, err := svc.GetUserByID(ctx, nonExistentID)
		
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrUserNotFound, err, "should return ErrUserNotFound")
	})

	t.Run("returns inactive users", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create inactive user
		inactiveUser := &models.User{
			Email:        "inactive@example.com",
			PasswordHash: "hashedpassword",
			Name:         "Inactive User",
			Role:         models.RoleNurse,
			IsActive:     false,
		}
		inactiveUser = createTestUser(t, db, inactiveUser)

		// GetUserByID should return inactive users (unlike Login)
		user, err := svc.GetUserByID(ctx, inactiveUser.ID)
		
		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, inactiveUser.ID, user.ID)
		assert.False(t, user.IsActive)
	})
}

func TestService_ChangePassword(t *testing.T) {
	ctx := context.Background()

	t.Run("changes password with valid old password", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create user with known password
		oldPassword := "oldpassword123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Test User",
			Role:         models.RoleNurse,
			IsActive:     true,
		}
		user = createTestUser(t, db, user)

		// Change password
		newPassword := "newpassword456"
		req := &models.ChangePasswordRequest{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}

		err = svc.ChangePassword(ctx, user.ID, req)
		require.NoError(t, err)

		// Fetch updated user from database
		var updatedUser models.User
		err = db.First(&updatedUser, "id = ?", user.ID).Error
		require.NoError(t, err)

		// Verify new password works
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte(newPassword))
		assert.NoError(t, err, "new password should work")

		// Verify old password no longer works
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte(oldPassword))
		assert.Error(t, err, "old password should not work")
		assert.Equal(t, bcrypt.ErrMismatchedHashAndPassword, err)
	})

	t.Run("rejects wrong old password", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create user with known password
		actualPassword := "actualpassword"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(actualPassword), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Test User",
			Role:         models.RoleNurse,
			IsActive:     true,
		}
		user = createTestUser(t, db, user)

		// Try to change password with wrong old password
		req := &models.ChangePasswordRequest{
			OldPassword: "wrongpassword",
			NewPassword: "newpassword123",
		}

		err = svc.ChangePassword(ctx, user.ID, req)
		
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPassword, err, "should return ErrInvalidPassword")

		// Verify password was not changed
		var unchangedUser models.User
		err = db.First(&unchangedUser, "id = ?", user.ID).Error
		require.NoError(t, err)
		
		err = bcrypt.CompareHashAndPassword([]byte(unchangedUser.PasswordHash), []byte(actualPassword))
		assert.NoError(t, err, "original password should still work")
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		svc, _, _ := newTestService(t)

		nonExistentID := uuid.New().String()
		req := &models.ChangePasswordRequest{
			OldPassword: "oldpass",
			NewPassword: "newpass",
		}

		err := svc.ChangePassword(ctx, nonExistentID, req)
		
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err, "should return ErrUserNotFound")
	})

	t.Run("new password is properly hashed", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create user
		oldPassword := "oldpass123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Test User",
			Role:         models.RoleNurse,
			IsActive:     true,
		}
		user = createTestUser(t, db, user)

		// Change password
		newPassword := "newSecurePassword!@#123"
		req := &models.ChangePasswordRequest{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		}

		err = svc.ChangePassword(ctx, user.ID, req)
		require.NoError(t, err)

		// Verify new password is hashed, not plaintext
		var updatedUser models.User
		err = db.First(&updatedUser, "id = ?", user.ID).Error
		require.NoError(t, err)

		assert.NotEqual(t, newPassword, updatedUser.PasswordHash, "password should be hashed")
		assert.Greater(t, len(updatedUser.PasswordHash), 50, "bcrypt hash should be long")
	})
}

func TestService_DeactivateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("deactivates user", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create active user
		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: "hashedpassword",
			Name:         "Test User",
			Role:         models.RoleNurse,
			IsActive:     true,
		}
		user = createTestUser(t, db, user)

		// Verify user is active
		assert.True(t, user.IsActive)

		// Deactivate user
		err := svc.DeactivateUser(ctx, user.ID)
		require.NoError(t, err)

		// Verify user is now inactive
		var deactivatedUser models.User
		err = db.First(&deactivatedUser, "id = ?", user.ID).Error
		require.NoError(t, err)
		assert.False(t, deactivatedUser.IsActive, "user should be inactive")
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		svc, _, _ := newTestService(t)

		nonExistentID := uuid.New().String()
		err := svc.DeactivateUser(ctx, nonExistentID)
		
		assert.Error(t, err)
		assert.Equal(t, ErrUserNotFound, err, "should return ErrUserNotFound")
	})

	t.Run("deactivated user cannot login", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create user with known password
		password := "password123"
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: string(hashedPassword),
			Name:         "Test User",
			Role:         models.RoleNurse,
			IsActive:     true,
		}
		user = createTestUser(t, db, user)

		// Deactivate user
		err = svc.DeactivateUser(ctx, user.ID)
		require.NoError(t, err)

		// Try to login with correct credentials
		loginReq := &models.LoginRequest{
			Email:    user.Email,
			Password: password,
		}

		resp, err := svc.Login(ctx, loginReq)
		
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, ErrUserInactive, err, "deactivated user should not be able to login")
	})

	t.Run("deactivating already inactive user succeeds", func(t *testing.T) {
		svc, db, _ := newTestService(t)

		// Create already inactive user
		user := &models.User{
			Email:        "test@example.com",
			PasswordHash: "hashedpassword",
			Name:         "Test User",
			Role:         models.RoleNurse,
			IsActive:     false,
		}
		user = createTestUser(t, db, user)

		// Deactivate again (should succeed)
		err := svc.DeactivateUser(ctx, user.ID)
		require.NoError(t, err)

		// Verify still inactive
		var stillInactiveUser models.User
		err = db.First(&stillInactiveUser, "id = ?", user.ID).Error
		require.NoError(t, err)
		assert.False(t, stillInactiveUser.IsActive)
	})
}

// Integration-style test for password hashing and verification
func TestPasswordHashingRoundTrip(t *testing.T) {
	passwords := []string{
		"password123",
		"VerySecureP@ssw0rd!",
		"12345678",
		"admin@123",
	}

	for _, password := range passwords {
		t.Run("password: "+password, func(t *testing.T) {
			// Hash password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			assert.NoError(t, err)
			assert.NotEmpty(t, hashedPassword)

			// Verify correct password
			err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
			assert.NoError(t, err)

			// Verify wrong password fails
			err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password+"wrong"))
			assert.Error(t, err)
		})
	}
}
