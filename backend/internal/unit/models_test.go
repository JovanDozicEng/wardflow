package unit

import "testing"

func TestUnit_TableName(t *testing.T) {
	u := Unit{}
	if u.TableName() != "units" {
		t.Errorf("TableName() = %q, want %q", u.TableName(), "units")
	}
}
