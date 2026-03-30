package careteam

import (
	"time"
)

// RoleType represents the clinical role in a care team
type RoleType string

const (
	RolePrimaryNurse       RoleType = "primary_nurse"
	RoleSecondaryNurse     RoleType = "secondary_nurse"
	RoleAttendingProvider  RoleType = "attending_provider"
	RoleResidentProvider   RoleType = "resident_provider"
	RoleConsultProvider    RoleType = "consult_provider"
	RoleChargeNurse        RoleType = "charge_nurse"
	RoleCaseManager        RoleType = "case_manager"
	RoleSocialWorker       RoleType = "social_worker"
)

// CriticalRoles are roles that require handoff notes during transfers
var CriticalRoles = map[RoleType]bool{
	RolePrimaryNurse:      true,
	RoleAttendingProvider: true,
}

// CareTeamAssignment represents a historical care team assignment
// This table is append-only; assignments are never updated or deleted
type CareTeamAssignment struct {
	ID             string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID    string     `json:"encounterId" gorm:"type:uuid;not null;index:idx_care_team_encounter"`
	UserID         string     `json:"userId" gorm:"type:uuid;not null;index:idx_care_team_user"`
	RoleType       RoleType   `json:"roleType" gorm:"type:varchar(50);not null"`
	StartsAt       time.Time  `json:"startsAt" gorm:"not null"`
	EndsAt         *time.Time `json:"endsAt,omitempty" gorm:"index:idx_care_team_active"`
	CreatedBy      string     `json:"createdBy" gorm:"type:uuid;not null"`
	CreatedAt      time.Time  `json:"createdAt"`
	HandoffNoteID  *string    `json:"handoffNoteId,omitempty" gorm:"type:uuid"`
	
	// Virtual fields (not persisted)
	HandoffNote *HandoffNote `json:"handoffNote,omitempty" gorm:"foreignKey:HandoffNoteID;references:ID"`
}

// TableName returns the table name for GORM
func (CareTeamAssignment) TableName() string {
	return "care_team_assignments"
}

// IsActive returns true if the assignment is currently active (EndsAt is nil)
func (a *CareTeamAssignment) IsActive() bool {
	return a.EndsAt == nil
}

// HandoffNote represents structured handoff documentation
// Immutable after creation
type HandoffNote struct {
	ID                  string                 `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID         string                 `json:"encounterId" gorm:"type:uuid;not null;index:idx_handoff_encounter"`
	FromUserID          string                 `json:"fromUserId" gorm:"type:uuid;not null"`
	ToUserID            string                 `json:"toUserId" gorm:"type:uuid;not null"`
	RoleType            RoleType               `json:"roleType" gorm:"type:varchar(50);not null"`
	Note                string                 `json:"note" gorm:"type:text;not null"`
	StructuredFieldsJSON map[string]interface{} `json:"structuredFields,omitempty" gorm:"type:jsonb;serializer:json"`
	CreatedAt           time.Time              `json:"createdAt" gorm:"index:idx_handoff_created"`
}

// TableName returns the table name for GORM
func (HandoffNote) TableName() string {
	return "handoff_notes"
}

// --- Request/Response DTOs ---

// AssignRoleRequest is the request body for assigning a role to an encounter
type AssignRoleRequest struct {
	UserID   string    `json:"userId" binding:"required"`
	RoleType RoleType  `json:"roleType" binding:"required"`
	StartsAt *time.Time `json:"startsAt,omitempty"` // defaults to now if omitted
}

// TransferRoleRequest is the request body for transferring a role to another user
type TransferRoleRequest struct {
	ToUserID         string                 `json:"toUserId" binding:"required"`
	HandoffNote      string                 `json:"handoffNote" binding:"required"`
	StructuredFields map[string]interface{} `json:"structuredFields,omitempty"`
	EndsAt           *time.Time             `json:"endsAt,omitempty"` // defaults to now if omitted
}

// ListAssignmentsResponse contains care team assignments with optional filters
type ListAssignmentsResponse struct {
	Assignments []CareTeamAssignment `json:"assignments"`
	Total       int64                `json:"total"`
}

// ListHandoffsResponse contains handoff notes for an encounter
type ListHandoffsResponse struct {
	Handoffs []HandoffNote `json:"handoffs"`
	Total    int64         `json:"total"`
}

// CareTeamMember represents an active care team member with user details populated
type CareTeamMember struct {
	AssignmentID string    `json:"assignmentId"`
	UserID       string    `json:"userId"`
	UserName     string    `json:"userName"`
	UserEmail    string    `json:"userEmail"`
	RoleType     RoleType  `json:"roleType"`
	StartsAt     time.Time `json:"startsAt"`
}

// CareTeamResponse is the response for getting the current care team
type CareTeamResponse struct {
	EncounterID string           `json:"encounterId"`
	Members     []CareTeamMember `json:"members"`
}
