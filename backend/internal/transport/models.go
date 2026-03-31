package transport

import "time"

type TransportStatus string

const (
	TransportStatusPending   TransportStatus = "pending"
	TransportStatusAssigned  TransportStatus = "assigned"
	TransportStatusInTransit TransportStatus = "in_transit"
	TransportStatusCompleted TransportStatus = "completed"
	TransportStatusCancelled TransportStatus = "cancelled"
)

// TransportRequest represents a patient transport request
type TransportRequest struct {
	ID          string          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID string          `json:"encounterId" gorm:"not null;index"`
	Origin      string          `json:"origin" gorm:"not null"`
	Destination string          `json:"destination" gorm:"not null"`
	Priority    string          `json:"priority" gorm:"type:varchar(20);not null;default:'routine'"` // routine, urgent, emergent
	Status      TransportStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	AssignedTo  *string         `json:"assignedTo,omitempty" gorm:"type:uuid"`
	AssignedAt  *time.Time      `json:"assignedAt,omitempty"`
	CreatedBy   string          `json:"createdBy" gorm:"not null"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// TableName returns the table name for TransportRequest
func (TransportRequest) TableName() string {
	return "transport_requests"
}

// TransportChangeEvent represents a transport request change event
type TransportChangeEvent struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RequestID     string    `json:"requestId" gorm:"not null;index"`
	ChangedFields string    `json:"changedFields" gorm:"type:jsonb;not null"` // JSON map of field->newValue
	ChangedBy     string    `json:"changedBy" gorm:"not null"`
	Reason        *string   `json:"reason,omitempty"`
	ChangedAt     time.Time `json:"changedAt" gorm:"not null"`
	CreatedAt     time.Time `json:"createdAt"`
}

// TableName returns the table name for TransportChangeEvent
func (TransportChangeEvent) TableName() string {
	return "transport_change_events"
}

// CreateTransportRequest represents the request body for creating a transport request
type CreateTransportRequest struct {
	EncounterID string `json:"encounterId"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Priority    string `json:"priority"`
}

// UpdateTransportRequest represents the request body for updating a transport request
type UpdateTransportRequest struct {
	Origin      *string `json:"origin,omitempty"`
	Destination *string `json:"destination,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	Reason      *string `json:"reason,omitempty"`
}

// AcceptTransportRequest represents the request body for accepting a transport request
type AcceptTransportRequest struct {
	AssignedTo string `json:"assignedTo"` // user ID of transport staff
}

// ListTransportFilter represents filters for listing transport requests
type ListTransportFilter struct {
	Status  string
	UnitID  string
	UnitIDs []string // For filtering by multiple units (non-admin users)
	Limit   int
	Offset  int
}
