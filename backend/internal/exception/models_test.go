package exception

import "testing"

func TestExceptionEvent_TableName(t *testing.T) {
	e := ExceptionEvent{}
	if e.TableName() != "exception_events" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "exception_events")
	}
}

func TestExceptionStatus_Values(t *testing.T) {
	statuses := []ExceptionStatus{
		ExceptionStatusDraft,
		ExceptionStatusFinalized,
		ExceptionStatusCorrected,
	}
	seen := make(map[ExceptionStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty ExceptionStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate ExceptionStatus: %q", s)
		}
		seen[s] = true
	}
}
