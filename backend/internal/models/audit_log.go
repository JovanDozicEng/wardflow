package models

import "time"

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID            string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EntityType    string     `json:"entityType" gorm:"not null;index"`
	EntityID      string     `json:"entityId" gorm:"not null;index"`
	Action        string     `json:"action" gorm:"not null"`          // CREATE, UPDATE, DELETE, OVERRIDE
	At            time.Time  `json:"at" gorm:"not null;index"`
	ByUserID      string     `json:"byUserId" gorm:"not null;index"`
	IP            *string    `json:"ip,omitempty"`
	UserAgent     *string    `json:"userAgent,omitempty"`
	Reason        *string    `json:"reason,omitempty"`
	Source        string     `json:"source" gorm:"not null;default:'user_action'"` // user_action | system_event
	BeforeJSON    *string    `json:"beforeJson,omitempty" gorm:"type:jsonb"`
	AfterJSON     *string    `json:"afterJson,omitempty" gorm:"type:jsonb"`
	CorrelationID *string    `json:"correlationId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// TableName returns the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_log"
}
