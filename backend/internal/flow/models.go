package flow

import (
	"time"
)

// FlowState represents a patient flow state
type FlowState string

const (
	StateArrived        FlowState = "arrived"
	StateTriage         FlowState = "triage"
	StateProviderEval   FlowState = "provider_eval"
	StateDiagnostics    FlowState = "diagnostics"
	StateAdmitted       FlowState = "admitted"
	StateDischargeReady FlowState = "discharge_ready"
	StateDischarged     FlowState = "discharged"
)

// ValidTransitions defines allowed state transitions
// Key: current state, Value: list of valid next states
var ValidTransitions = map[FlowState][]FlowState{
	StateArrived:        {StateTriage},
	StateTriage:         {StateProviderEval, StateDischargeReady},
	StateProviderEval:   {StateDiagnostics, StateAdmitted, StateDischargeReady},
	StateDiagnostics:    {StateProviderEval, StateAdmitted},
	StateAdmitted:       {StateDischargeReady},
	StateDischargeReady: {StateDischarged},
	// StateDischarged is terminal - no valid transitions
}

// IsValidTransition checks if a transition from one state to another is valid
func IsValidTransition(from, to FlowState) bool {
	validStates, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, state := range validStates {
		if state == to {
			return true
		}
	}
	return false
}

// ActorType represents who initiated the state transition
type ActorType string

const (
	ActorTypeUser   ActorType = "user"
	ActorTypeSystem ActorType = "system"
)

// FlowStateTransition represents an immutable state transition event
// This table is append-only; transitions are never updated or deleted
type FlowStateTransition struct {
	ID             string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EncounterID    string     `json:"encounterId" gorm:"type:uuid;not null;index:idx_flow_encounter"`
	FromState      *FlowState `json:"fromState,omitempty" gorm:"type:varchar(50)"` // null for initial state
	ToState        FlowState  `json:"toState" gorm:"type:varchar(50);not null;index:idx_flow_to_state"`
	TransitionedAt time.Time  `json:"transitionedAt" gorm:"not null;index:idx_flow_transitioned"`
	ActorType      ActorType  `json:"actorType" gorm:"type:varchar(50);not null"`
	ActorUserID    *string    `json:"actorUserId,omitempty" gorm:"type:uuid"` // null for system events
	Reason         *string    `json:"reason,omitempty" gorm:"type:text"`       // required for manual corrections
	SourceEventID  *string    `json:"sourceEventId,omitempty" gorm:"type:uuid"` // links to triggering event
	IsOverride     bool       `json:"isOverride" gorm:"default:false"`          // marks privileged invalid transitions
	CreatedAt      time.Time  `json:"createdAt"`
}

// TableName returns the table name for GORM
func (FlowStateTransition) TableName() string {
	return "flow_state_transitions"
}

// --- Request/Response DTOs ---

// CreateTransitionRequest is the request body for recording a state transition
type CreateTransitionRequest struct {
	ToState        FlowState  `json:"toState" binding:"required"`
	TransitionedAt *time.Time `json:"transitionedAt,omitempty"` // defaults to now if omitted
	Reason         *string    `json:"reason,omitempty"`
}

// OverrideTransitionRequest is the request body for privileged state transition override
type OverrideTransitionRequest struct {
	FromState      *FlowState `json:"fromState,omitempty"` // explicit from state for override
	ToState        FlowState  `json:"toState" binding:"required"`
	Reason         string     `json:"reason" binding:"required"` // reason is mandatory for overrides
	TransitionedAt *time.Time `json:"transitionedAt,omitempty"`
}

// FlowTimelineResponse contains the flow state transition timeline
type FlowTimelineResponse struct {
	EncounterID  string                `json:"encounterId"`
	CurrentState *FlowState            `json:"currentState,omitempty"`
	Transitions  []FlowStateTransition `json:"transitions"`
	Total        int64                 `json:"total"`
}

// TransitionWithActor includes actor details in the response
type TransitionWithActor struct {
	FlowStateTransition
	ActorName  *string `json:"actorName,omitempty"`
	ActorEmail *string `json:"actorEmail,omitempty"`
}

// FlowTimelineDetailResponse includes actor details
type FlowTimelineDetailResponse struct {
	EncounterID  string                `json:"encounterId"`
	CurrentState *FlowState            `json:"currentState,omitempty"`
	Transitions  []TransitionWithActor `json:"transitions"`
	Total        int64                 `json:"total"`
}
