package bed

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringSlice is a []string that serializes to/from JSON for GORM JSONB columns.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (s *StringSlice) Scan(value any) error {
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("cannot scan type %T into StringSlice", value)
	}
	return json.Unmarshal(bytes, s)
}

func (s StringSlice) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("[]"), nil
	}
	type plain []string
	return json.Marshal(plain(s))
}

type BedStatus string

const (
	BedStatusAvailable   BedStatus = "available"
	BedStatusOccupied    BedStatus = "occupied"
	BedStatusBlocked     BedStatus = "blocked"
	BedStatusCleaning    BedStatus = "cleaning"
	BedStatusMaintenance BedStatus = "maintenance"
)

// Bed represents a hospital bed in the system
type Bed struct {
	ID                 string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UnitID             string     `json:"unitId" gorm:"not null;index"`
	Room               string     `json:"room" gorm:"not null"`
	Label              string     `json:"label" gorm:"not null"` // e.g. "Bed A", "Room 101-1"
	Capabilities       StringSlice `json:"capabilities" gorm:"type:jsonb;default:'[]'"`
	CurrentStatus      BedStatus  `json:"currentStatus" gorm:"type:varchar(20);not null;default:'available';index"`
	CurrentEncounterID *string    `json:"currentEncounterId,omitempty" gorm:"type:uuid;index"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// TableName returns the table name for Bed
func (Bed) TableName() string {
	return "beds"
}

// BedStatusEvent represents a bed status change event
type BedStatusEvent struct {
	ID         string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BedID      string     `json:"bedId" gorm:"not null;index"`
	FromStatus *BedStatus `json:"fromStatus,omitempty" gorm:"type:varchar(20)"`
	ToStatus   BedStatus  `json:"toStatus" gorm:"type:varchar(20);not null"`
	Reason     *string    `json:"reason,omitempty"`
	ChangedBy  string     `json:"changedBy" gorm:"not null"`
	ChangedAt  time.Time  `json:"changedAt" gorm:"not null"`
	CreatedAt  time.Time  `json:"createdAt"`
}

// TableName returns the table name for BedStatusEvent
func (BedStatusEvent) TableName() string {
	return "bed_status_events"
}

type BedRequestStatus string

const (
	BedRequestStatusPending   BedRequestStatus = "pending"
	BedRequestStatusAssigned  BedRequestStatus = "assigned"
	BedRequestStatusCancelled BedRequestStatus = "cancelled"
)

// BedRequest represents a request for a bed assignment
type BedRequest struct {
	ID                   string           `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID          string           `json:"encounterId" gorm:"not null;index"`
	RequiredCapabilities StringSlice      `json:"requiredCapabilities" gorm:"type:jsonb;default:'[]'"`
	Priority             string           `json:"priority" gorm:"type:varchar(20);not null;default:'routine'"` // routine, urgent, emergent
	Status               BedRequestStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	AssignedBedID        *string          `json:"assignedBedId,omitempty" gorm:"type:uuid"`
	CreatedBy            string           `json:"createdBy" gorm:"not null"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

// TableName returns the table name for BedRequest
func (BedRequest) TableName() string {
	return "bed_requests"
}

// CreateBedRequest represents the request body for creating a bed
type CreateBedRequest struct {
	UnitID       string   `json:"unitId"`
	Room         string   `json:"room"`
	Label        string   `json:"label"`
	Capabilities []string `json:"capabilities"`
}

// UpdateBedStatusRequest represents the request body for updating bed status
type UpdateBedStatusRequest struct {
	Status BedStatus `json:"status"`
	Reason *string   `json:"reason,omitempty"`
}

// CreateBedRequestRequest represents the request body for creating a bed request
type CreateBedRequestRequest struct {
	RequiredCapabilities []string `json:"requiredCapabilities"`
	Priority             string   `json:"priority"`
}

// AssignBedRequest represents the request body for assigning a bed
type AssignBedRequest struct {
	BedID string `json:"bedId"`
}

// ListBedsFilter represents filters for listing beds
type ListBedsFilter struct {
	UnitID string
	Status string
	Limit  int
	Offset int
}
