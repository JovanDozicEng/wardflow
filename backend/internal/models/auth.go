package models

// RegisterRequest for user registration
type RegisterRequest struct {
	Email         string      `json:"email" binding:"required,email"`
	Password      string      `json:"password" binding:"required,min=8"`
	Name          string      `json:"name" binding:"required"`
	Role          Role        `json:"role" binding:"required"`
	UnitIDs       StringArray `json:"unitIds"`
	DepartmentIDs StringArray `json:"departmentIds"`
}

// LoginRequest for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse returned after successful login
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt int64     `json:"expiresAt"`
	User      *UserInfo `json:"user"`
}

// UserInfo contains safe user information (without password)
type UserInfo struct {
	ID            string      `json:"id"`
	Email         string      `json:"email"`
	Name          string      `json:"name"`
	Role          Role        `json:"role"`
	UnitIDs       StringArray `json:"unitIds"`
	DepartmentIDs StringArray `json:"departmentIds"`
	IsActive      bool        `json:"isActive"`
}

// ToUserInfo converts User to UserInfo (safe representation)
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Role:          u.Role,
		UnitIDs:       u.UnitIDs,
		DepartmentIDs: u.DepartmentIDs,
		IsActive:      u.IsActive,
	}
}

// ChangePasswordRequest for password updates
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}
