package discharge

import "time"

type ChecklistStatus string

const (
	ChecklistStatusInProgress       ChecklistStatus = "in_progress"
	ChecklistStatusComplete         ChecklistStatus = "complete"
	ChecklistStatusOverrideComplete ChecklistStatus = "override_complete"
)

type ItemStatus string

const (
	ItemStatusOpen   ItemStatus = "open"
	ItemStatusDone   ItemStatus = "done"
	ItemStatusWaived ItemStatus = "waived"
)

// DischargeChecklist represents a discharge checklist for an encounter
type DischargeChecklist struct {
	ID             string          `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID    string          `json:"encounterId" gorm:"not null;uniqueIndex"` // one per encounter
	DischargeType  string          `json:"dischargeType" gorm:"type:varchar(50);not null"`
	Status         ChecklistStatus `json:"status" gorm:"type:varchar(30);not null;default:'in_progress'"`
	CompletedBy    *string         `json:"completedBy,omitempty" gorm:"type:uuid"`
	CompletedAt    *time.Time      `json:"completedAt,omitempty"`
	OverrideReason *string         `json:"overrideReason,omitempty"`
	CreatedBy      string          `json:"createdBy" gorm:"not null"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	Items          []DischargeChecklistItem `json:"items,omitempty" gorm:"-"`
}

// TableName returns the table name for DischargeChecklist
func (DischargeChecklist) TableName() string {
	return "discharge_checklists"
}

// DischargeChecklistItem represents an item in a discharge checklist
type DischargeChecklistItem struct {
	ID          string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ChecklistID string     `json:"checklistId" gorm:"not null;index"`
	Code        string     `json:"code" gorm:"not null"`
	Label       string     `json:"label" gorm:"not null"`
	Required    bool       `json:"required" gorm:"not null;default:true"`
	Status      ItemStatus `json:"status" gorm:"type:varchar(10);not null;default:'open'"`
	CompletedBy *string    `json:"completedBy,omitempty" gorm:"type:uuid"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// TableName returns the table name for DischargeChecklistItem
func (DischargeChecklistItem) TableName() string {
	return "discharge_checklist_items"
}

// InitChecklistRequest represents the request body for initializing a discharge checklist
type InitChecklistRequest struct {
	DischargeType string `json:"dischargeType"` // standard, ama, lwbs
}

// CompleteItemRequest represents the request body for completing a checklist item
type CompleteItemRequest struct {
	// no body needed; completedBy comes from auth context
}

// CompleteDischargeRequest represents the request body for completing discharge
type CompleteDischargeRequest struct {
	Override bool    `json:"override"`
	Reason   *string `json:"reason,omitempty"` // required when override=true
}

// DefaultItemDefinition represents a default checklist item
type DefaultItemDefinition struct {
	Code     string
	Label    string
	Required bool
}

// DefaultItems returns the default checklist items for a discharge type
func DefaultItems(dischargeType string) []DefaultItemDefinition {
	base := []DefaultItemDefinition{
		{"patient_education", "Patient education completed", true},
		{"medication_reconciliation", "Medication reconciliation done", true},
		{"follow_up_appointment", "Follow-up appointment scheduled", true},
		{"discharge_summary", "Discharge summary signed", true},
		{"transport_arranged", "Transport arranged", false},
		{"belongings_returned", "Patient belongings returned", false},
	}

	switch dischargeType {
	case "ama":
		return []DefaultItemDefinition{
			{"ama_form_signed", "AMA form signed by patient", true},
			{"risks_explained", "Risks of leaving explained and documented", true},
			{"physician_notified", "Attending physician notified", true},
		}
	case "lwbs":
		return []DefaultItemDefinition{
			{"lwbs_documented", "LWBS documented in chart", true},
			{"contact_attempted", "Contact attempt documented", true},
		}
	default:
		return base
	}
}
