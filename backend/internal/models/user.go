package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Role represents user roles in the system
type Role string

const (
	RoleNurse           Role = "nurse"
	RoleProvider        Role = "provider"
	RoleChargeNurse     Role = "charge_nurse"
	RoleOperations      Role = "operations"
	RoleConsult         Role = "consult"
	RoleTransport       Role = "transport"
	RoleQualitySafety   Role = "quality_safety"
	RoleAdmin           Role = "admin"
)

// StringArray is a custom type for storing string arrays as JSON
type StringArray []string

// Scan implements sql.Scanner interface
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = []string{}
		return nil
	}
	
	// Handle different types PostgreSQL might return
	switch v := value.(type) {
	case []byte:
		// PostgreSQL json/jsonb returns as []byte
		if len(v) == 0 || string(v) == "null" {
			*sa = []string{}
			return nil
		}
		return json.Unmarshal(v, sa)
	case string:
		// Sometimes returns as string
		if v == "" || v == "null" {
			*sa = []string{}
			return nil
		}
		return json.Unmarshal([]byte(v), sa)
	default:
		return fmt.Errorf("failed to unmarshal StringArray: unsupported type %T", value)
	}
}

// Value implements driver.Valuer interface
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil || len(sa) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(sa)
}

// User represents a system user
type User struct {
	ID           string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string         `json:"-" gorm:"not null"`
	Name         string         `json:"name" gorm:"not null"`
	IsActive     bool           `json:"isActive" gorm:"default:true;not null"`
	Role         Role           `json:"role" gorm:"type:varchar(50);not null;index"`
	
	// Unit/Department assignment for visibility boundaries (stored as JSON)
	UnitIDs       StringArray    `json:"unitIds" gorm:"type:jsonb;default:'[]'"`
	DepartmentIDs StringArray    `json:"departmentIds" gorm:"type:jsonb;default:'[]'"`
	
	CreatedAt    time.Time      `json:"createdAt" gorm:"not null"`
	UpdatedAt    time.Time      `json:"updatedAt" gorm:"not null"`
	DeletedAt    gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// HasRole checks if user has a specific role
func (u *User) HasRole(role Role) bool {
	return u.Role == role || u.Role == RoleAdmin
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanAccessUnit checks if user can access a specific unit
func (u *User) CanAccessUnit(unitID string) bool {
	if u.IsAdmin() {
		return true
	}
	for _, id := range u.UnitIDs {
		if id == unitID {
			return true
		}
	}
	return false
}

// CanAccessDepartment checks if user can access a specific department
func (u *User) CanAccessDepartment(deptID string) bool {
	if u.IsAdmin() {
		return true
	}
	for _, id := range u.DepartmentIDs {
		if id == deptID {
			return true
		}
	}
	return false
}
