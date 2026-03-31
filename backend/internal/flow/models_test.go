package flow

import "testing"

func TestFlowStateTransition_TableName(t *testing.T) {
	f := FlowStateTransition{}
	if f.TableName() != "flow_state_transitions" {
		t.Errorf("TableName() = %q, want %q", f.TableName(), "flow_state_transitions")
	}
}

func TestFlowState_Values(t *testing.T) {
	states := []FlowState{
		StateArrived,
		StateTriage,
		StateProviderEval,
		StateDiagnostics,
		StateAdmitted,
		StateDischargeReady,
		StateDischarged,
	}
	seen := make(map[FlowState]bool)
	for _, s := range states {
		if string(s) == "" {
			t.Errorf("empty FlowState value")
		}
		if seen[s] {
			t.Errorf("duplicate FlowState: %q", s)
		}
		seen[s] = true
	}
}

func TestActorType_Values(t *testing.T) {
	actors := []ActorType{
		ActorTypeUser,
		ActorTypeSystem,
	}
	seen := make(map[ActorType]bool)
	for _, a := range actors {
		if string(a) == "" {
			t.Errorf("empty ActorType value")
		}
		if seen[a] {
			t.Errorf("duplicate ActorType: %q", a)
		}
		seen[a] = true
	}
}

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name  string
		from  FlowState
		to    FlowState
		valid bool
	}{
		{"arrived to triage", StateArrived, StateTriage, true},
		{"triage to provider eval", StateTriage, StateProviderEval, true},
		{"triage to discharge ready", StateTriage, StateDischargeReady, true},
		{"provider eval to diagnostics", StateProviderEval, StateDiagnostics, true},
		{"provider eval to admitted", StateProviderEval, StateAdmitted, true},
		{"provider eval to discharge ready", StateProviderEval, StateDischargeReady, true},
		{"diagnostics to provider eval", StateDiagnostics, StateProviderEval, true},
		{"diagnostics to admitted", StateDiagnostics, StateAdmitted, true},
		{"admitted to discharge ready", StateAdmitted, StateDischargeReady, true},
		{"discharge ready to discharged", StateDischargeReady, StateDischarged, true},
		{"arrived to admitted - invalid", StateArrived, StateAdmitted, false},
		{"arrived to discharged - invalid", StateArrived, StateDischarged, false},
		{"discharged to anything - invalid", StateDischarged, StateArrived, false},
		{"triage to admitted - invalid", StateTriage, StateAdmitted, false},
		{"admitted to triage - invalid", StateAdmitted, StateTriage, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTransition(tt.from, tt.to)
			if result != tt.valid {
				t.Errorf("IsValidTransition(%v, %v) = %v, want %v", tt.from, tt.to, result, tt.valid)
			}
		})
	}
}
