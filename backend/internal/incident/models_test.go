package incident

import "testing"

func TestIncident_TableName(t *testing.T) {
	i := Incident{}
	if i.TableName() != "incidents" {
		t.Errorf("TableName() = %q, want %q", i.TableName(), "incidents")
	}
}

func TestIncidentStatusEvent_TableName(t *testing.T) {
	e := IncidentStatusEvent{}
	if e.TableName() != "incident_status_events" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "incident_status_events")
	}
}

func TestIncidentStatus_Values(t *testing.T) {
	statuses := []IncidentStatus{
		IncidentStatusSubmitted,
		IncidentStatusUnderReview,
		IncidentStatusClosed,
	}
	seen := make(map[IncidentStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty IncidentStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate IncidentStatus: %q", s)
		}
		seen[s] = true
	}
}
