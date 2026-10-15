// Package domain holds the core business entities and rules of hacelo.
// It has no dependencies on frameworks, databases or transport layers.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Status is the position of a task in its reminder lifecycle.
//
//	pending   -> waiting for the initial reminder
//	verifying -> the user said "I'm on it"; waiting to confirm it is done
//	overdue   -> verification rounds exhausted; retried the next day
//	done      -> completed, never reminded again
type Status string

const (
	StatusPending   Status = "pending"
	StatusVerifying Status = "verifying"
	StatusOverdue   Status = "overdue"
	StatusDone      Status = "done"
)

// Urgency is how important the user marked a task.
type Urgency string

const (
	UrgencyLow    Urgency = "low"
	UrgencyMedium Urgency = "medium"
	UrgencyHigh   Urgency = "high"
)

const (
	// MaxVerificationRounds is how many follow-up reminders a task gets
	// before it becomes overdue.
	MaxVerificationRounds = 3
	// DefaultSnoozeMinutes is the follow-up interval when none is given.
	DefaultSnoozeMinutes = 15
)

var (
	ErrNotFound    = errors.New("task not found")
	ErrInvalid     = errors.New("invalid input")
	ErrAlreadyDone = errors.New("task is already done")
)

// Valid reports whether u is a known urgency.
func (u Urgency) Valid() bool {
	return u == UrgencyLow || u == UrgencyMedium || u == UrgencyHigh
}

// Task is the central entity: it fires a reminder at NextReminderAt and
// moves through the Status lifecycle according to the user's answers.
type Task struct {
	ID                int64
	Description       string
	Urgency           Urgency
	Status            Status
	OriginalTime      time.Time
	SnoozeMinutes     int
	NextReminderAt    *time.Time
	VerificationRound int
	LastSentAt        *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	CompletedAt       *time.Time
}

// NewTask builds a pending task scheduled at scheduledAt. An empty urgency
// defaults to medium and a non-positive snoozeMinutes to the default.
func NewTask(description string, urgency Urgency, scheduledAt time.Time, snoozeMinutes int, now time.Time) (*Task, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil, fmt.Errorf("%w: description is required", ErrInvalid)
	}
	if urgency == "" {
		urgency = UrgencyMedium
	}
	if !urgency.Valid() {
		return nil, fmt.Errorf("%w: unknown urgency %q", ErrInvalid, urgency)
	}
	if snoozeMinutes <= 0 {
		snoozeMinutes = DefaultSnoozeMinutes
	}

	next := scheduledAt
	return &Task{
		Description:    description,
		Urgency:        urgency,
		Status:         StatusPending,
		OriginalTime:   scheduledAt,
		SnoozeMinutes:  snoozeMinutes,
		NextReminderAt: &next,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// UpdateDetails changes the editable fields; nil arguments are left as is.
func (t *Task) UpdateDetails(description *string, urgency *Urgency, snoozeMinutes *int, now time.Time) error {
	if description != nil {
		trimmed := strings.TrimSpace(*description)
		if trimmed == "" {
			return fmt.Errorf("%w: description cannot be empty", ErrInvalid)
		}
		t.Description = trimmed
	}
	if urgency != nil {
		if !urgency.Valid() {
			return fmt.Errorf("%w: unknown urgency %q", ErrInvalid, *urgency)
		}
		t.Urgency = *urgency
	}
	if snoozeMinutes != nil {
		if *snoozeMinutes <= 0 {
			return fmt.Errorf("%w: snooze minutes must be positive", ErrInvalid)
		}
		t.SnoozeMinutes = *snoozeMinutes
	}
	t.UpdatedAt = now
	return nil
}

// Reschedule manually sets the next reminder and restarts the cycle. The new
// time also becomes the task's reference time, so "same time tomorrow" after
// an overdue is computed from it and not from the stale original time.
func (t *Task) Reschedule(at time.Time, now time.Time) {
	t.Status = StatusPending
	t.VerificationRound = 0
	t.OriginalTime = at
	t.NextReminderAt = &at
	t.LastSentAt = nil
	t.UpdatedAt = now
}

// Snooze handles the "I'm on it" answer. While rounds remain, it schedules
// a verification reminder SnoozeMinutes from now; once they are exhausted
// the task becomes overdue and is retried tomorrow at the original time.
func (t *Task) Snooze(now time.Time) error {
	if t.Status == StatusDone {
		return ErrAlreadyDone
	}

	var next time.Time
	if t.VerificationRound >= MaxVerificationRounds {
		t.Status = StatusOverdue
		t.VerificationRound = 0
		next = t.OriginalTime.AddDate(0, 0, 1)
	} else {
		t.Status = StatusVerifying
		t.VerificationRound++
		next = now.Add(time.Duration(t.SnoozeMinutes) * time.Minute)
	}

	t.NextReminderAt = &next
	t.LastSentAt = nil
	t.UpdatedAt = now
	return nil
}

// Complete handles the "I did it" answer.
func (t *Task) Complete(now time.Time) {
	t.Status = StatusDone
	t.NextReminderAt = nil
	t.LastSentAt = nil
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// IsDue reports whether a reminder must be sent now: the scheduled time has
// passed and no reminder was sent yet for that time.
func (t *Task) IsDue(now time.Time) bool {
	if t.Status == StatusDone || t.NextReminderAt == nil {
		return false
	}
	if t.NextReminderAt.After(now) {
		return false
	}
	return t.LastSentAt == nil || t.LastSentAt.Before(*t.NextReminderAt)
}

// MarkReminderSent records that the reminder went out. An overdue task goes
// back to pending: a new day starts a fresh verification cycle.
func (t *Task) MarkReminderSent(now time.Time) {
	t.LastSentAt = &now
	if t.Status == StatusOverdue {
		t.Status = StatusPending
		t.VerificationRound = 0
	}
	t.UpdatedAt = now
}
