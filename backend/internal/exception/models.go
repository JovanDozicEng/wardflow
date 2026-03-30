package exception

import "time"

// ExceptionStatus represents the status of an exception event
type ExceptionStatus string

const (
	ExceptionStatusDraft     ExceptionStatus = "draft"
	ExceptionStatusFinalized ExceptionStatus = "finalized"
	ExceptionStatusCorrected ExceptionStatus = "corrected"
)

// ExceptionEvent represents an exception event
type ExceptionEvent struct {
	ID             string          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID    string          `json:"encounterId" gorm:"not null;index"`
	Type           string          `json:"type" gorm:"not null;index"`
	Status         ExceptionStatus `json:"status" gorm:"type:varchar(20);not null;default:'draft';index"`
	RequiredFields string          `json:"requiredFields" gorm:"type:jsonb;not null"`
	Data           string          `json:"data" gorm:"type:jsonb;not null"`

	InitiatedBy string    `json:"initiatedBy" gorm:"not null"`
	InitiatedAt time.Time `json:"initiatedAt"`

	FinalizedBy *string    `json:"finalizedBy,omitempty"`
	FinalizedAt *time.Time `json:"finalizedAt,omitempty"`

	CorrectedByEventID *string `json:"correctedByEventId,omitempty" gorm:"type:uuid"`
	CorrectionReason   *string `json:"correctionReason,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName returns the table name for ExceptionEvent
func (ExceptionEvent) TableName() string {
	return "exception_events"
}

// CreateExceptionRequest is the request body for creating an exception event
type CreateExceptionRequest struct {
	EncounterID string                 `json:"encounterId"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
}

// UpdateExceptionRequest is the request body for updating an exception event
type UpdateExceptionRequest struct {
	Data map[string]interface{} `json:"data"`
}

// CorrectExceptionRequest is the request body for correcting an exception event
type CorrectExceptionRequest struct {
	Reason string                 `json:"reason"`
	Data   map[string]interface{} `json:"data"`
}

// ListExceptionsFilter holds filters for listing exception events
type ListExceptionsFilter struct {
	EncounterID string
	Type        string
	Status      ExceptionStatus
	Limit       int
	Offset      int
}
