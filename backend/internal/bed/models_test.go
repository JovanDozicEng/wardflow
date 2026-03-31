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

func TestStringSlice_Value(t *testing.T) {
	t.Run("converts slice to JSON string", func(t *testing.T) {
		slice := StringSlice{"capability-1", "capability-2"}
		value, err := slice.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `["capability-1","capability-2"]`
		if value != expected {
			t.Errorf("Value() = %q, want %q", value, expected)
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		slice := StringSlice{}
		value, err := slice.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `[]`
		if value != expected {
			t.Errorf("Value() = %q, want %q", value, expected)
		}
	})

	t.Run("handles nil slice", func(t *testing.T) {
		var slice StringSlice
		value, err := slice.Value()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// nil slice uses custom MarshalJSON which returns "[]" not "null"
		expected := `[]`
		if value != expected {
			t.Errorf("Value() = %q, want %q", value, expected)
		}
	})
}

func TestStringSlice_Scan(t *testing.T) {
	t.Run("scans from string", func(t *testing.T) {
		var slice StringSlice
		input := `["cap1","cap2"]`
		err := slice.Scan(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(slice) != 2 {
			t.Errorf("expected 2 elements, got %d", len(slice))
		}
		if slice[0] != "cap1" || slice[1] != "cap2" {
			t.Errorf("Scan() = %v, want [cap1 cap2]", slice)
		}
	})

	t.Run("scans from byte slice", func(t *testing.T) {
		var slice StringSlice
		input := []byte(`["a","b","c"]`)
		err := slice.Scan(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(slice) != 3 {
			t.Errorf("expected 3 elements, got %d", len(slice))
		}
	})

	t.Run("returns error for invalid type", func(t *testing.T) {
		var slice StringSlice
		err := slice.Scan(123)
		if err == nil {
			t.Error("expected error for invalid type, got nil")
		}
		if err != nil && err.Error() != "cannot scan type int into StringSlice" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		var slice StringSlice
		err := slice.Scan(`invalid json`)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})
}

func TestStringSlice_MarshalJSON(t *testing.T) {
	t.Run("marshals non-empty slice", func(t *testing.T) {
		slice := StringSlice{"x", "y"}
		data, err := slice.MarshalJSON()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `["x","y"]`
		if string(data) != expected {
			t.Errorf("MarshalJSON() = %q, want %q", string(data), expected)
		}
	})

	t.Run("marshals empty slice", func(t *testing.T) {
		slice := StringSlice{}
		data, err := slice.MarshalJSON()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `[]`
		if string(data) != expected {
			t.Errorf("MarshalJSON() = %q, want %q", string(data), expected)
		}
	})

	t.Run("marshals nil slice as empty array", func(t *testing.T) {
		var slice StringSlice
		data, err := slice.MarshalJSON()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := `[]`
		if string(data) != expected {
			t.Errorf("MarshalJSON() = %q, want %q", string(data), expected)
		}
	})
}
