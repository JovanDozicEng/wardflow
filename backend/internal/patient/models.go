package patient

import "time"

// Patient represents a patient in the system
type Patient struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FirstName   string     `json:"firstName" gorm:"not null"`
	LastName    string     `json:"lastName" gorm:"not null"`
	DateOfBirth *time.Time `json:"dateOfBirth,omitempty"`
	MRN         string     `json:"mrn" gorm:"uniqueIndex;not null"` // Medical Record Number
	CreatedBy   string     `json:"createdBy" gorm:"not null"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// TableName returns the table name for Patient
func (Patient) TableName() string {
	return "patients"
}

// CreatePatientRequest represents the request body for creating a patient
type CreatePatientRequest struct {
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	DateOfBirth *string `json:"dateOfBirth,omitempty"` // ISO date string "2000-01-15"
	MRN         string  `json:"mrn"`
}

// ListPatientsFilter represents filters for listing patients
type ListPatientsFilter struct {
	Q      string // search: first_name, last_name, mrn ILIKE
	Limit  int
	Offset int
}
