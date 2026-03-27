package encounter

import "time"

// EncounterStatus represents the status of an encounter
type EncounterStatus string

const (
	EncounterStatusActive     EncounterStatus = "active"
	EncounterStatusDischarged EncounterStatus = "discharged"
	EncounterStatusCancelled  EncounterStatus = "cancelled"
)

// Encounter represents a patient encounter
type Encounter struct {
	ID           string          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PatientID    string          `json:"patientId" gorm:"not null;index"`
	UnitID       string          `json:"unitId" gorm:"not null;index"`
	DepartmentID string          `json:"departmentId" gorm:"not null;index"`
	Status       EncounterStatus `json:"status" gorm:"type:varchar(50);not null;default:'active';index"`
	StartedAt    time.Time       `json:"startedAt" gorm:"not null"`
	EndedAt      *time.Time      `json:"endedAt,omitempty"`
	CreatedBy    string          `json:"createdBy" gorm:"not null"`
	UpdatedBy    string          `json:"updatedBy" gorm:"not null"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

// TableName returns the table name for Encounter
func (Encounter) TableName() string {
	return "encounters"
}

// CreateEncounterRequest is the request body for creating an encounter
type CreateEncounterRequest struct {
	PatientID    string     `json:"patientId"`
	UnitID       string     `json:"unitId"`
	DepartmentID string     `json:"departmentId"`
	StartedAt    *time.Time `json:"startedAt,omitempty"` // defaults to now
}

// UpdateEncounterRequest is the request body for updating an encounter
type UpdateEncounterRequest struct {
	Status  *EncounterStatus `json:"status,omitempty"`
	EndedAt *time.Time       `json:"endedAt,omitempty"`
	UnitID  *string          `json:"unitId,omitempty"`
}

// ListEncountersFilter holds filters for listing encounters
type ListEncountersFilter struct {
	UnitID       string
	DepartmentID string
	Status       string
	Limit        int
	Offset       int
}
