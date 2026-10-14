// Package port declares the interfaces the application core exposes
// (inbound) and the ones it needs from the outside world (outbound).
package port

import (
	"context"
	"time"

	"hacelo/internal/domain"
)

// CreateTaskCommand is the input to create a task.
type CreateTaskCommand struct {
	Description   string
	Urgency       domain.Urgency
	ScheduledAt   time.Time
	SnoozeMinutes int
}

// UpdateTaskCommand carries optional changes; nil fields are untouched.
// Setting NextReminderAt reschedules the task and restarts its cycle.
type UpdateTaskCommand struct {
	Description    *string
	Urgency        *domain.Urgency
	SnoozeMinutes  *int
	NextReminderAt *time.Time
}

// TaskFilter narrows down a task listing. Empty fields mean "no filter".
type TaskFilter struct {
	Date   string // YYYY-MM-DD
	Status domain.Status
}

// TaskUseCases is the inbound port for managing tasks.
type TaskUseCases interface {
	CreateTask(ctx context.Context, cmd CreateTaskCommand) (*domain.Task, error)
	GetTask(ctx context.Context, id int64) (*domain.Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]domain.Task, error)
	UpdateTask(ctx context.Context, id int64, cmd UpdateTaskCommand) (*domain.Task, error)
	DeleteTask(ctx context.Context, id int64) error
	// SnoozeTask handles the "I'm on it" answer to a reminder.
	SnoozeTask(ctx context.Context, id int64) (*domain.Task, error)
	// CompleteTask handles the "I did it" answer to a reminder.
	CompleteTask(ctx context.Context, id int64) (*domain.Task, error)
}

// SubscriptionUseCases is the inbound port for push subscriptions.
type SubscriptionUseCases interface {
	Subscribe(ctx context.Context, sub domain.PushSubscription) error
	Unsubscribe(ctx context.Context, endpoint string) error
}

// ReminderUseCases is the inbound port driven by the scheduler.
type ReminderUseCases interface {
	// DispatchDueReminders sends a notification for every task that is due.
	DispatchDueReminders(ctx context.Context) error
}
