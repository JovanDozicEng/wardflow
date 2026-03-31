package bed

import "testing"

func TestBedStatus_Values(t *testing.T) {
	statuses := []BedStatus{
		BedStatusAvailable,
		BedStatusOccupied,
		BedStatusBlocked,
		BedStatusCleaning,
		BedStatusMaintenance,
	}
	seen := make(map[BedStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty BedStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate BedStatus: %q", s)
		}
		seen[s] = true
	}
}

func TestBedRequestStatus_Values(t *testing.T) {
	statuses := []BedRequestStatus{
		BedRequestStatusPending,
		BedRequestStatusAssigned,
		BedRequestStatusCancelled,
	}
	seen := make(map[BedRequestStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty BedRequestStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate BedRequestStatus: %q", s)
		}
		seen[s] = true
	}
}

func TestBed_TableName(t *testing.T) {
	b := Bed{}
	if b.TableName() != "beds" {
		t.Errorf("TableName() = %q, want %q", b.TableName(), "beds")
	}
}

func TestBedStatusEvent_TableName(t *testing.T) {
	e := BedStatusEvent{}
	if e.TableName() != "bed_status_events" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "bed_status_events")
	}
}

func TestBedRequest_TableName(t *testing.T) {
	r := BedRequest{}
	if r.TableName() != "bed_requests" {
		t.Errorf("TableName() = %q, want %q", r.TableName(), "bed_requests")
	}
}

func TestBedStatus_AssignableStatuses(t *testing.T) {
	// Only available beds can be assigned
	assignable := map[BedStatus]bool{
		BedStatusAvailable: true,
	}
	notAssignable := []BedStatus{
		BedStatusOccupied,
		BedStatusBlocked,
		BedStatusCleaning,
		BedStatusMaintenance,
	}
	for _, s := range notAssignable {
		if assignable[s] {
			t.Errorf("status %q should not be assignable", s)
		}
	}
	if !assignable[BedStatusAvailable] {
		t.Error("BedStatusAvailable should be assignable")
	}
}
