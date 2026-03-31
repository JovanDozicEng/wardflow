package task

import (
	"testing"
	"time"
)

func TestTask_TableName(t *testing.T) {
	task := Task{}
	if task.TableName() != "tasks" {
		t.Errorf("TableName() = %q, want %q", task.TableName(), "tasks")
	}
}

func TestTaskAssignmentEvent_TableName(t *testing.T) {
	e := TaskAssignmentEvent{}
	if e.TableName() != "task_assignment_events" {
		t.Errorf("TableName() = %q, want %q", e.TableName(), "task_assignment_events")
	}
}

func TestTask_IsOverdue(t *testing.T) {
	past := time.Now().UTC().Add(-1 * time.Hour)
	future := time.Now().UTC().Add(1 * time.Hour)

	t.Run("overdue when SLA deadline has passed and task is open", func(t *testing.T) {
		task := &Task{
			SLADueAt: &past,
			Status:   TaskStatusOpen,
		}
		if !task.IsOverdue() {
			t.Error("IsOverdue() = false, want true")
		}
	})

	t.Run("overdue when SLA deadline has passed and task is in progress", func(t *testing.T) {
		task := &Task{
			SLADueAt: &past,
			Status:   TaskStatusInProgress,
		}
		if !task.IsOverdue() {
			t.Error("IsOverdue() = false, want true")
		}
	})

	t.Run("not overdue when SLA deadline is in future", func(t *testing.T) {
		task := &Task{
			SLADueAt: &future,
			Status:   TaskStatusOpen,
		}
		if task.IsOverdue() {
			t.Error("IsOverdue() = true, want false")
		}
	})

	t.Run("not overdue when no SLA deadline", func(t *testing.T) {
		task := &Task{
			SLADueAt: nil,
			Status:   TaskStatusOpen,
		}
		if task.IsOverdue() {
			t.Error("IsOverdue() = true, want false")
		}
	})

	t.Run("not overdue when task is completed", func(t *testing.T) {
		task := &Task{
			SLADueAt: &past,
			Status:   TaskStatusCompleted,
		}
		if task.IsOverdue() {
			t.Error("IsOverdue() = true, want false for completed task")
		}
	})

	t.Run("not overdue when task is cancelled", func(t *testing.T) {
		task := &Task{
			SLADueAt: &past,
			Status:   TaskStatusCancelled,
		}
		if task.IsOverdue() {
			t.Error("IsOverdue() = true, want false for cancelled task")
		}
	})

	t.Run("overdue when task is escalated but deadline passed", func(t *testing.T) {
		task := &Task{
			SLADueAt: &past,
			Status:   TaskStatusEscalated,
		}
		if !task.IsOverdue() {
			t.Error("IsOverdue() = false, want true for escalated task")
		}
	})
}

func TestTaskStatus_Values(t *testing.T) {
	statuses := []TaskStatus{
		TaskStatusOpen,
		TaskStatusInProgress,
		TaskStatusCompleted,
		TaskStatusCancelled,
		TaskStatusEscalated,
	}
	seen := make(map[TaskStatus]bool)
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("empty TaskStatus value")
		}
		if seen[s] {
			t.Errorf("duplicate TaskStatus: %q", s)
		}
		seen[s] = true
	}
}

func TestTaskPriority_Values(t *testing.T) {
	priorities := []TaskPriority{
		TaskPriorityLow,
		TaskPriorityMedium,
		TaskPriorityHigh,
		TaskPriorityUrgent,
	}
	seen := make(map[TaskPriority]bool)
	for _, p := range priorities {
		if string(p) == "" {
			t.Errorf("empty TaskPriority value")
		}
		if seen[p] {
			t.Errorf("duplicate TaskPriority: %q", p)
		}
		seen[p] = true
	}
}

func TestScopeType_Values(t *testing.T) {
	scopes := []ScopeType{
		ScopeTypeEncounter,
		ScopeTypePatient,
		ScopeTypeUnit,
	}
	seen := make(map[ScopeType]bool)
	for _, s := range scopes {
		if string(s) == "" {
			t.Errorf("empty ScopeType value")
		}
		if seen[s] {
			t.Errorf("duplicate ScopeType: %q", s)
		}
		seen[s] = true
	}
}
