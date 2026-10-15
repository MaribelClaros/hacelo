package domain

import (
	"errors"
	"testing"
	"time"
)

var baseTime = time.Date(2026, 10, 14, 18, 0, 0, 0, time.UTC)

func newTestTask(t *testing.T) *Task {
	t.Helper()
	task, err := NewTask("buy vegetables", UrgencyMedium, baseTime, 15, baseTime.Add(-time.Hour))
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	return task
}

func TestNewTaskDefaults(t *testing.T) {
	task, err := NewTask("  call dentist ", "", baseTime, 0, baseTime)
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	if task.Description != "call dentist" {
		t.Errorf("description = %q, want trimmed", task.Description)
	}
	if task.Urgency != UrgencyMedium {
		t.Errorf("urgency = %q, want medium", task.Urgency)
	}
	if task.SnoozeMinutes != DefaultSnoozeMinutes {
		t.Errorf("snooze = %d, want %d", task.SnoozeMinutes, DefaultSnoozeMinutes)
	}
	if task.Status != StatusPending {
		t.Errorf("status = %q, want pending", task.Status)
	}
}

func TestNewTaskValidation(t *testing.T) {
	if _, err := NewTask("   ", UrgencyLow, baseTime, 15, baseTime); !errors.Is(err, ErrInvalid) {
		t.Errorf("empty description: err = %v, want ErrInvalid", err)
	}
	if _, err := NewTask("x", Urgency("urgent"), baseTime, 15, baseTime); !errors.Is(err, ErrInvalid) {
		t.Errorf("bad urgency: err = %v, want ErrInvalid", err)
	}
}

func TestSnoozeSchedulesVerification(t *testing.T) {
	task := newTestTask(t)
	now := baseTime.Add(time.Minute)

	if err := task.Snooze(now); err != nil {
		t.Fatalf("Snooze: %v", err)
	}
	if task.Status != StatusVerifying || task.VerificationRound != 1 {
		t.Errorf("status=%q round=%d, want verifying/1", task.Status, task.VerificationRound)
	}
	want := now.Add(15 * time.Minute)
	if task.NextReminderAt == nil || !task.NextReminderAt.Equal(want) {
		t.Errorf("next reminder = %v, want %v", task.NextReminderAt, want)
	}
}

func TestSnoozeBecomesOverdueAfterMaxRounds(t *testing.T) {
	task := newTestTask(t)
	now := baseTime

	for i := 0; i < MaxVerificationRounds; i++ {
		if err := task.Snooze(now); err != nil {
			t.Fatalf("Snooze %d: %v", i, err)
		}
	}
	if task.VerificationRound != MaxVerificationRounds {
		t.Fatalf("round = %d, want %d", task.VerificationRound, MaxVerificationRounds)
	}

	if err := task.Snooze(now); err != nil {
		t.Fatalf("final Snooze: %v", err)
	}
	if task.Status != StatusOverdue || task.VerificationRound != 0 {
		t.Errorf("status=%q round=%d, want overdue/0", task.Status, task.VerificationRound)
	}
	want := baseTime.AddDate(0, 0, 1)
	if task.NextReminderAt == nil || !task.NextReminderAt.Equal(want) {
		t.Errorf("next reminder = %v, want %v (same time tomorrow)", task.NextReminderAt, want)
	}
}

func TestCompleteStopsReminders(t *testing.T) {
	task := newTestTask(t)
	task.Complete(baseTime)

	if task.Status != StatusDone || task.NextReminderAt != nil || task.CompletedAt == nil {
		t.Errorf("unexpected state after Complete: %+v", task)
	}
	if task.IsDue(baseTime.Add(24 * time.Hour)) {
		t.Error("a done task must never be due")
	}
	if err := task.Snooze(baseTime); !errors.Is(err, ErrAlreadyDone) {
		t.Errorf("Snooze on done task: err = %v, want ErrAlreadyDone", err)
	}
}

func TestRescheduleRestartsCycle(t *testing.T) {
	task := newTestTask(t)
	_ = task.Snooze(baseTime)

	at := baseTime.Add(3 * time.Hour)
	task.Reschedule(at, baseTime)

	if task.Status != StatusPending || task.VerificationRound != 0 {
		t.Errorf("status=%q round=%d, want pending/0", task.Status, task.VerificationRound)
	}
	if task.NextReminderAt == nil || !task.NextReminderAt.Equal(at) {
		t.Errorf("next reminder = %v, want %v", task.NextReminderAt, at)
	}
	if !task.OriginalTime.Equal(at) {
		t.Errorf("original time = %v, want %v", task.OriginalTime, at)
	}
}

func TestOverdueAfterRescheduleUsesNewTime(t *testing.T) {
	task := newTestTask(t)
	rescheduled := baseTime.AddDate(0, 0, 3)
	task.Reschedule(rescheduled, baseTime)

	for i := 0; i <= MaxVerificationRounds; i++ {
		if err := task.Snooze(rescheduled); err != nil {
			t.Fatalf("Snooze %d: %v", i, err)
		}
	}

	want := rescheduled.AddDate(0, 0, 1)
	if task.Status != StatusOverdue || task.NextReminderAt == nil || !task.NextReminderAt.Equal(want) {
		t.Errorf("status=%q next=%v, want overdue/%v", task.Status, task.NextReminderAt, want)
	}
}

func TestIsDueAndMarkReminderSent(t *testing.T) {
	task := newTestTask(t)

	if task.IsDue(baseTime.Add(-time.Minute)) {
		t.Error("task should not be due before its time")
	}
	if !task.IsDue(baseTime) {
		t.Error("task should be due at its time")
	}

	task.MarkReminderSent(baseTime)
	if task.IsDue(baseTime.Add(time.Minute)) {
		t.Error("task should not be due again once the reminder was sent")
	}
}

func TestMarkReminderSentResetsOverdue(t *testing.T) {
	task := newTestTask(t)
	task.Status = StatusOverdue

	task.MarkReminderSent(baseTime)
	if task.Status != StatusPending {
		t.Errorf("status = %q, want pending", task.Status)
	}
}
