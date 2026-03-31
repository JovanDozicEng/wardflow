package careteam

import (
	"testing"
	"time"
)

func TestCareTeamAssignment_TableName(t *testing.T) {
	a := CareTeamAssignment{}
	if a.TableName() != "care_team_assignments" {
		t.Errorf("TableName() = %q, want %q", a.TableName(), "care_team_assignments")
	}
}

func TestHandoffNote_TableName(t *testing.T) {
	h := HandoffNote{}
	if h.TableName() != "handoff_notes" {
		t.Errorf("TableName() = %q, want %q", h.TableName(), "handoff_notes")
	}
}

func TestCareTeamAssignment_IsActive(t *testing.T) {
	t.Run("returns true when EndsAt is nil", func(t *testing.T) {
		assignment := &CareTeamAssignment{
			ID:        "id-1",
			EndsAt:    nil,
			StartsAt:  time.Now(),
		}
		if !assignment.IsActive() {
			t.Error("IsActive() = false, want true")
		}
	})

	t.Run("returns false when EndsAt is set", func(t *testing.T) {
		endTime := time.Now().Add(-1 * time.Hour)
		assignment := &CareTeamAssignment{
			ID:       "id-1",
			EndsAt:   &endTime,
			StartsAt: time.Now().Add(-2 * time.Hour),
		}
		if assignment.IsActive() {
			t.Error("IsActive() = true, want false")
		}
	})
}

func TestRoleType_Values(t *testing.T) {
	roles := []RoleType{
		RolePrimaryNurse,
		RoleSecondaryNurse,
		RoleAttendingProvider,
		RoleResidentProvider,
		RoleConsultProvider,
		RoleChargeNurse,
		RoleCaseManager,
		RoleSocialWorker,
	}
	seen := make(map[RoleType]bool)
	for _, r := range roles {
		if string(r) == "" {
			t.Errorf("empty RoleType value")
		}
		if seen[r] {
			t.Errorf("duplicate RoleType: %q", r)
		}
		seen[r] = true
	}
}

func TestCriticalRoles(t *testing.T) {
	t.Run("primary nurse is critical", func(t *testing.T) {
		if !CriticalRoles[RolePrimaryNurse] {
			t.Error("RolePrimaryNurse should be critical")
		}
	})

	t.Run("attending provider is critical", func(t *testing.T) {
		if !CriticalRoles[RoleAttendingProvider] {
			t.Error("RoleAttendingProvider should be critical")
		}
	})

	t.Run("secondary nurse is not critical", func(t *testing.T) {
		if CriticalRoles[RoleSecondaryNurse] {
			t.Error("RoleSecondaryNurse should not be critical")
		}
	})

	t.Run("case manager is not critical", func(t *testing.T) {
		if CriticalRoles[RoleCaseManager] {
			t.Error("RoleCaseManager should not be critical")
		}
	})
}
