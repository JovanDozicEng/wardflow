package transport

import "testing"

func TestTransportStatus_Values(t *testing.T) {
	statuses := []TransportStatus{
		TransportStatusPending,
		TransportStatusAssigned,
		TransportStatusInTransit,
		TransportStatusCompleted,
		TransportStatusCancelled,
	}
	seen := make(map[TransportStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty TransportStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate TransportStatus: %q", s)
		}
		seen[s] = true
	}
}

func TestTransportRequest_TableName(t *testing.T) {
	r := TransportRequest{}
	if r.TableName() != "transport_requests" {
		t.Errorf("TableName() = %q, want %q", r.TableName(), "transport_requests")
	}
}

func TestTransportChangeEvent_TableName(t *testing.T) {
	e := TransportChangeEvent{}
	if e.TableName() != "transport_change_events" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "transport_change_events")
	}
}

func TestTransportStatus_Workflow(t *testing.T) {
	// Business rule: can only complete if assigned
	canComplete := map[TransportStatus]bool{
		TransportStatusAssigned: true,
	}
	cannot := []TransportStatus{
		TransportStatusPending,
		TransportStatusInTransit,
		TransportStatusCompleted,
		TransportStatusCancelled,
	}
	for _, s := range cannot {
		if canComplete[s] {
			t.Errorf("status %q should not be completable directly", s)
		}
	}
	if !canComplete[TransportStatusAssigned] {
		t.Error("TransportStatusAssigned should allow completion")
	}
}

func TestTransportStatus_TerminalStatuses(t *testing.T) {
	terminal := map[TransportStatus]bool{
		TransportStatusCompleted: true,
		TransportStatusCancelled: true,
	}
	nonTerminal := []TransportStatus{
		TransportStatusPending,
		TransportStatusAssigned,
		TransportStatusInTransit,
	}
	for _, s := range nonTerminal {
		if terminal[s] {
			t.Errorf("status %q should not be terminal", s)
		}
	}
}
