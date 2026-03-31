package encounter

import "testing"

func TestEncounter_TableName(t *testing.T) {
	e := Encounter{}
	if e.TableName() != "encounters" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "encounters")
	}
}

func TestEncounterStatus_Values(t *testing.T) {
	statuses := []EncounterStatus{
		EncounterStatusActive,
		EncounterStatusDischarged,
		EncounterStatusCancelled,
	}
	seen := make(map[EncounterStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty EncounterStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate EncounterStatus: %q", s)
		}
		seen[s] = true
	}
}
