package consult

import "testing"

func TestConsultRequest_TableName(t *testing.T) {
	c := ConsultRequest{}
	if c.TableName() != "consult_requests" {
		t.Errorf("TableName() = %q, want %q", c.TableName(), "consult_requests")
	}
}

func TestConsultUrgency_Values(t *testing.T) {
	urgencies := []ConsultUrgency{
		ConsultUrgencyRoutine,
		ConsultUrgencyUrgent,
		ConsultUrgencyEmergent,
	}
	seen := make(map[ConsultUrgency]bool)
	for _, u := range urgencies {
		if string(u) == "" {
			t.Errorf("empty ConsultUrgency value")
		}
		if seen[u] {
			t.Errorf("duplicate ConsultUrgency: %q", u)
		}
		seen[u] = true
	}
}

func TestConsultStatus_Values(t *testing.T) {
	statuses := []ConsultStatus{
		ConsultStatusPending,
		ConsultStatusAccepted,
		ConsultStatusDeclined,
		ConsultStatusCompleted,
		ConsultStatusRedirected,
		ConsultStatusCancelled,
	}
	seen := make(map[ConsultStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty ConsultStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate ConsultStatus: %q", s)
		}
		seen[s] = true
	}
}
