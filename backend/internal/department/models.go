package department

import "time"

// Department represents a hospital department (e.g., Emergency, Cardiology)
type Department struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex"`
	Code      string    `json:"code" gorm:"not null;uniqueIndex"` // e.g. "EMERGENCY", "CARDIOLOGY"
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (Department) TableName() string {
	return "departments"
}

// CreateDepartmentRequest represents the request body for creating a department
type CreateDepartmentRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
