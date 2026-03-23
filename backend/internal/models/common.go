package models

import "time"

// Common audit fields embedded in domain models
type AuditFields struct {
	CreatedAt time.Time  `json:"createdAt" gorm:"not null"`
	UpdatedAt time.Time  `json:"updatedAt" gorm:"not null"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" gorm:"index"`
}

// CreatedBy tracks who created a record
type CreatedByField struct {
	CreatedBy string `json:"createdBy" gorm:"not null;index"`
}

// UpdatedBy tracks who last updated a record
type UpdatedByField struct {
	UpdatedBy string `json:"updatedBy" gorm:"not null;index"`
}

// SoftDelete provides soft delete capability
type SoftDelete struct {
	DeletedAt *time.Time `json:"deletedAt,omitempty" gorm:"index"`
	DeletedBy *string    `json:"deletedBy,omitempty"`
}
