package patient

import "testing"

func TestPatient_TableName(t *testing.T) {
	p := Patient{}
	if p.TableName() != "patients" {
		t.Errorf("TableName() = %q, want %q", p.TableName(), "patients")
	}
}
