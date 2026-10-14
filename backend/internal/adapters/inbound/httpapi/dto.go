package httpapi

import (
	"time"

	"hacelo/internal/domain"
)

// taskResponse is the JSON representation of a task.
type taskResponse struct {
	ID                int64      `json:"id"`
	Description       string     `json:"description"`
	Urgency           string     `json:"urgency"`
	Status            string     `json:"status"`
	OriginalTime      time.Time  `json:"original_time"`
	SnoozeMinutes     int        `json:"snooze_minutes"`
	NextReminderAt    *time.Time `json:"next_reminder_at,omitempty"`
	VerificationRound int        `json:"verification_round"`
	LastSentAt        *time.Time `json:"last_sent_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
}

func toTaskResponse(t *domain.Task) taskResponse {
	return taskResponse{
		ID:                t.ID,
		Description:       t.Description,
		Urgency:           string(t.Urgency),
		Status:            string(t.Status),
		OriginalTime:      t.OriginalTime,
		SnoozeMinutes:     t.SnoozeMinutes,
		NextReminderAt:    t.NextReminderAt,
		VerificationRound: t.VerificationRound,
		LastSentAt:        t.LastSentAt,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
		CompletedAt:       t.CompletedAt,
	}
}

func toTaskResponses(tasks []domain.Task) []taskResponse {
	out := make([]taskResponse, 0, len(tasks))
	for i := range tasks {
		out = append(out, toTaskResponse(&tasks[i]))
	}
	return out
}

// createTaskRequest is the body of POST /api/tasks.
type createTaskRequest struct {
	Description   string `json:"description"`
	Urgency       string `json:"urgency"`
	ScheduledAt   string `json:"scheduled_at"` // RFC 3339
	SnoozeMinutes int    `json:"snooze_minutes"`
}

// updateTaskRequest is the body of PATCH /api/tasks/{id}.
type updateTaskRequest struct {
	Description    *string `json:"description"`
	Urgency        *string `json:"urgency"`
	SnoozeMinutes  *int    `json:"snooze_minutes"`
	NextReminderAt *string `json:"next_reminder_at"` // RFC 3339
}

// subscribeRequest is the body of POST /api/push/subscribe, as produced by
// PushSubscription.toJSON() in the browser.
type subscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

type unsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}
