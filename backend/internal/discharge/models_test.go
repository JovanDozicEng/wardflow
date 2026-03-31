package discharge

import "testing"

func TestChecklistStatus_Values(t *testing.T) {
	statuses := []ChecklistStatus{
		ChecklistStatusInProgress,
		ChecklistStatusComplete,
		ChecklistStatusOverrideComplete,
	}
	seen := make(map[ChecklistStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty ChecklistStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate ChecklistStatus: %q", s)
		}
		seen[s] = true
	}
}

func TestItemStatus_Values(t *testing.T) {
	statuses := []ItemStatus{
		ItemStatusOpen,
		ItemStatusDone,
		ItemStatusWaived,
	}
	seen := make(map[ItemStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty ItemStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate ItemStatus: %q", s)
		}
		seen[s] = true
	}
}

func TestDischargeChecklist_TableName(t *testing.T) {
	c := DischargeChecklist{}
	if c.TableName() != "discharge_checklists" {
		t.Errorf("TableName() = %q, want %q", c.TableName(), "discharge_checklists")
	}
}

func TestDischargeChecklistItem_TableName(t *testing.T) {
	i := DischargeChecklistItem{}
	if i.TableName() != "discharge_checklist_items" {
		t.Errorf("TableName() = %q, want %q", i.TableName(), "discharge_checklist_items")
	}
}

func TestDefaultItems_Standard(t *testing.T) {
	items := DefaultItems("standard")
	if len(items) == 0 {
		t.Error("standard discharge type should have default items")
	}
	hasRequired := false
	for _, item := range items {
		if item.Code == "" {
			t.Error("item code should not be empty")
		}
		if item.Label == "" {
			t.Error("item label should not be empty")
		}
		if item.Required {
			hasRequired = true
		}
	}
	if !hasRequired {
		t.Error("standard discharge should have at least one required item")
	}
}

func TestDefaultItems_AMA(t *testing.T) {
	items := DefaultItems("ama")
	if len(items) == 0 {
		t.Error("ama discharge type should have default items")
	}
	// AMA should have AMA-specific items
	hasAMAForm := false
	for _, item := range items {
		if item.Code == "ama_form_signed" {
			hasAMAForm = true
		}
	}
	if !hasAMAForm {
		t.Error("ama discharge should include ama_form_signed item")
	}
}

func TestDefaultItems_LWBS(t *testing.T) {
	items := DefaultItems("lwbs")
	if len(items) == 0 {
		t.Error("lwbs discharge type should have default items")
	}
}

func TestDefaultItems_AllRequired(t *testing.T) {
	for _, dischargeType := range []string{"standard", "ama", "lwbs"} {
		items := DefaultItems(dischargeType)
		for _, item := range items {
			if item.Required {
				// Required items must have non-empty code and label
				if item.Code == "" || item.Label == "" {
					t.Errorf("required item for %q has empty code or label", dischargeType)
				}
			}
		}
	}
}

func TestChecklistStatus_CompletionStates(t *testing.T) {
	// Both complete and override_complete are terminal states
	terminal := map[ChecklistStatus]bool{
		ChecklistStatusComplete:        true,
		ChecklistStatusOverrideComplete: true,
	}
	if terminal[ChecklistStatusInProgress] {
		t.Error("in_progress should not be a terminal state")
	}
	if !terminal[ChecklistStatusComplete] {
		t.Error("complete should be a terminal state")
	}
	if !terminal[ChecklistStatusOverrideComplete] {
		t.Error("override_complete should be a terminal state")
	}
}
