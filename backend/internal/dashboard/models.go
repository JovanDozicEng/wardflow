package dashboard

import (
	"time"

	"github.com/wardflow/backend/internal/flow"
	"github.com/wardflow/backend/internal/task"
)

// HuddleMetrics contains aggregated metrics for the daily huddle dashboard
type HuddleMetrics struct {
	UnitID       *string              `json:"unitId,omitempty"`
	DepartmentID *string              `json:"departmentId,omitempty"`
	GeneratedAt  time.Time            `json:"generatedAt"`
	
	// Census metrics
	ActiveEncounters     int64 `json:"activeEncounters"`
	ExpectedDischarges   int64 `json:"expectedDischarges"`
	ExpectedAdmissions   int64 `json:"expectedAdmissions"`
	
	// Flow state distribution
	FlowDistribution     FlowDistribution `json:"flowDistribution"`
	
	// Risk indicators
	RiskIndicators       RiskIndicators `json:"riskIndicators"`
	
	// Task metrics
	TaskMetrics          TaskMetrics `json:"taskMetrics"`
	
	// Drill-down lists
	OverdueTasks         []TaskSummary      `json:"overdueTasks,omitempty"`
	LongStayPatients     []EncounterSummary `json:"longStayPatients,omitempty"`
	PendingDischarges    []EncounterSummary `json:"pendingDischarges,omitempty"`
}

// FlowDistribution shows count of encounters in each flow state
type FlowDistribution struct {
	Arrived        int64 `json:"arrived"`
	Triage         int64 `json:"triage"`
	ProviderEval   int64 `json:"providerEval"`
	Diagnostics    int64 `json:"diagnostics"`
	Admitted       int64 `json:"admitted"`
	DischargeReady int64 `json:"dischargeReady"`
	Discharged     int64 `json:"discharged"`
}

// RiskIndicators highlights potential issues
type RiskIndicators struct {
	PatientsInTriageOver2hrs  int64 `json:"patientsInTriageOver2hrs"`
	PatientsWaitingForBedOver4hrs int64 `json:"patientsWaitingForBedOver4hrs"`
	OverdueHighPriorityTasks  int64 `json:"overdueHighPriorityTasks"`
	UnassignedUrgentTasks     int64 `json:"unassignedUrgentTasks"`
	EncountersWithoutCareTeam int64 `json:"encountersWithoutCareTeam"`
}

// TaskMetrics provides task board overview
type TaskMetrics struct {
	TotalOpen      int64 `json:"totalOpen"`
	TotalOverdue   int64 `json:"totalOverdue"`
	HighPriority   int64 `json:"highPriority"`
	Urgent         int64 `json:"urgent"`
	Unassigned     int64 `json:"unassigned"`
	CompletedToday int64 `json:"completedToday"`
}

// TaskSummary is a lightweight task representation for lists
type TaskSummary struct {
	ID        string             `json:"id"`
	Title     string             `json:"title"`
	Priority  task.TaskPriority  `json:"priority"`
	SLADueAt  *time.Time         `json:"slaDueAt,omitempty"`
	OwnerName *string            `json:"ownerName,omitempty"`
	ScopeType task.ScopeType     `json:"scopeType"`
	ScopeID   string             `json:"scopeId"`
}

// EncounterSummary is a lightweight encounter representation for lists
type EncounterSummary struct {
	ID           string           `json:"id"`
	PatientID    string           `json:"patientId"`
	UnitID       string           `json:"unitId"`
	DepartmentID string           `json:"departmentId"`
	CurrentState *flow.FlowState  `json:"currentState,omitempty"`
	StartedAt    time.Time        `json:"startedAt"`
	LengthOfStay string           `json:"lengthOfStay"` // e.g., "2d 3h"
}

// FilterParams holds filter parameters for dashboard queries
type FilterParams struct {
	UnitID       *string
	DepartmentID *string
	StartTime    *time.Time // For time-range queries
	EndTime      *time.Time
}

// OverdueTaskSummary provides context for overdue tasks
type OverdueTaskSummary struct {
	TaskID       string            `json:"taskId"`
	Title        string            `json:"title"`
	Priority     task.TaskPriority `json:"priority"`
	SLADueAt     time.Time         `json:"slaDueAt"`
	OverdueBy    string            `json:"overdueBy"` // e.g., "2h 15m"
	AssignedTo   *string           `json:"assignedTo,omitempty"`
	EncounterID  *string           `json:"encounterId,omitempty"`
}

// DelayedConsultSummary provides context for delayed consults
// (This would be used if we had consult module, kept for future)
type DelayedConsultSummary struct {
	ConsultID   string    `json:"consultId"`
	Service     string    `json:"service"`
	RequestedAt time.Time `json:"requestedAt"`
	DelayedBy   string    `json:"delayedBy"`
	Urgency     string    `json:"urgency"`
}

// StaffingMetrics provides staffing overview (future enhancement)
type StaffingMetrics struct {
	NursesOnDuty     int64   `json:"nursesOnDuty"`
	ProvidersOnDuty  int64   `json:"providersOnDuty"`
	NurseToPatientRatio float64 `json:"nurseToPatientRatio"`
}
