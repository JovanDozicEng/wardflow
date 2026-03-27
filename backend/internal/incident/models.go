package incident

import "time"

// IncidentStatus represents the status of an incident
type IncidentStatus string

const (
	IncidentStatusSubmitted   IncidentStatus = "submitted"
	IncidentStatusUnderReview IncidentStatus = "under_review"
	IncidentStatusClosed      IncidentStatus = "closed"
)

// Incident represents a safety incident report
type Incident struct {
	ID             string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID    *string        `json:"encounterId,omitempty"`
	Type           string         `json:"type" gorm:"not null;index"`
	Severity       *string        `json:"severity,omitempty"`
	HarmIndicators *string        `json:"harmIndicators,omitempty" gorm:"type:jsonb"`
	EventTime      time.Time      `json:"eventTime" gorm:"not null;index"`
	ReportedBy     string         `json:"reportedBy" gorm:"not null"`
	ReportedAt     time.Time      `json:"reportedAt"`
	Status         IncidentStatus `json:"status" gorm:"type:varchar(20);not null;default:'submitted';index"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// TableName returns the table name for Incident
func (Incident) TableName() string {
	return "incidents"
}

// IncidentStatusEvent represents a status change event for an incident
type IncidentStatusEvent struct {
	ID         string          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	IncidentID string          `json:"incidentId" gorm:"not null;index"`
	FromStatus *IncidentStatus `json:"fromStatus,omitempty" gorm:"type:varchar(20)"`
	ToStatus   IncidentStatus  `json:"toStatus" gorm:"type:varchar(20);not null"`
	ChangedBy  string          `json:"changedBy" gorm:"not null"`
	ChangedAt  time.Time       `json:"changedAt"`
	Note       *string         `json:"note,omitempty"`
}

// TableName returns the table name for IncidentStatusEvent
func (IncidentStatusEvent) TableName() string {
	return "incident_status_events"
}

// CreateIncidentRequest is the request body for creating an incident
type CreateIncidentRequest struct {
	EncounterID    *string                `json:"encounterId,omitempty"`
	Type           string                 `json:"type"`
	Severity       *string                `json:"severity,omitempty"`
	HarmIndicators map[string]interface{} `json:"harmIndicators,omitempty"`
	EventTime      time.Time              `json:"eventTime"`
}

// UpdateIncidentStatusRequest is the request body for updating incident status
type UpdateIncidentStatusRequest struct {
	Status IncidentStatus `json:"status"`
	Note   *string        `json:"note,omitempty"`
}

// ListIncidentsFilter holds filters for listing incidents
type ListIncidentsFilter struct {
	UnitID string
	Status IncidentStatus
	Type   string
	Limit  int
	Offset int
}
