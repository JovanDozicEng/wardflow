package consult

import "time"

// ConsultUrgency represents the urgency level of a consult request
type ConsultUrgency string

const (
	ConsultUrgencyRoutine  ConsultUrgency = "routine"
	ConsultUrgencyUrgent   ConsultUrgency = "urgent"
	ConsultUrgencyEmergent ConsultUrgency = "emergent"
)

// ConsultStatus represents the status of a consult request
type ConsultStatus string

const (
	ConsultStatusPending    ConsultStatus = "pending"
	ConsultStatusAccepted   ConsultStatus = "accepted"
	ConsultStatusDeclined   ConsultStatus = "declined"
	ConsultStatusCompleted  ConsultStatus = "completed"
	ConsultStatusRedirected ConsultStatus = "redirected"
	ConsultStatusCancelled  ConsultStatus = "cancelled"
)

// ConsultRequest represents a consult request
type ConsultRequest struct {
	ID            string         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID   string         `json:"encounterId" gorm:"not null;index"`
	TargetService string         `json:"targetService" gorm:"not null;index"`
	Reason        string         `json:"reason" gorm:"not null"`
	Urgency       ConsultUrgency `json:"urgency" gorm:"type:varchar(20);not null"`
	Status        ConsultStatus  `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`

	CreatedBy string    `json:"createdBy" gorm:"not null"`
	CreatedAt time.Time `json:"createdAt"`

	AcceptedBy *string    `json:"acceptedBy,omitempty"`
	AcceptedAt *time.Time `json:"acceptedAt,omitempty"`

	ClosedAt    *time.Time `json:"closedAt,omitempty"`
	CloseReason *string    `json:"closeReason,omitempty"`

	RedirectedTo *string `json:"redirectedTo,omitempty"`

	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName returns the table name for ConsultRequest
func (ConsultRequest) TableName() string {
	return "consult_requests"
}

// CreateConsultRequest is the request body for creating a consult
type CreateConsultRequest struct {
	EncounterID   string         `json:"encounterId"`
	TargetService string         `json:"targetService"`
	Reason        string         `json:"reason"`
	Urgency       ConsultUrgency `json:"urgency"`
}

// DeclineConsultRequest is the request body for declining a consult
type DeclineConsultRequest struct {
	Reason string `json:"reason"` // Required
}

// RedirectConsultRequest is the request body for redirecting a consult
type RedirectConsultRequest struct {
	TargetService string `json:"targetService"`
	Reason        string `json:"reason"`
}

// RedirectResult holds both the closed original consult and the newly created one
type RedirectResult struct {
	Original  *ConsultRequest `json:"original"`
	NewConsult *ConsultRequest `json:"newConsult"`
}

// ListConsultsFilter holds filters for listing consult requests
type ListConsultsFilter struct {
	UnitID        string
	Status        ConsultStatus
	TargetService string
	Limit         int
	Offset        int
}
