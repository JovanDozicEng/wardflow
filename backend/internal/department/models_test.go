package department

import "testing"

func TestDepartment_TableName(t *testing.T) {
	d := Department{}
	if d.TableName() != "departments" {
		t.Errorf("TableName() = %q, want %q", d.TableName(), "departments")
	}
}
