package unit

import "time"

// Unit represents a hospital unit within a department (e.g., ICU, ED, Ward 4B)
type Unit struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string    `json:"name" gorm:"not null;uniqueIndex"`
	Code         string    `json:"code" gorm:"not null;uniqueIndex"` // e.g. "ICU", "ED", "WARD-4B"
	DepartmentID string    `json:"departmentId" gorm:"not null;index"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (Unit) TableName() string {
	return "units"
}

// CreateUnitRequest represents the request body for creating a unit
type CreateUnitRequest struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	DepartmentID string `json:"departmentId"`
}
