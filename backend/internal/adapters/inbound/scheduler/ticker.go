// Package scheduler is an inbound adapter: a periodic trigger that drives
// the reminder use case. It has no business logic of its own.
package scheduler

import (
	"context"
	"log/slog"
	"time"

	"hacelo/internal/application/port"
)

// Ticker calls DispatchDueReminders at a fixed interval.
type Ticker struct {
	reminders port.ReminderUseCases
	interval  time.Duration
	logger    *slog.Logger
}

func NewTicker(reminders port.ReminderUseCases, interval time.Duration, logger *slog.Logger) *Ticker {
	return &Ticker{reminders: reminders, interval: interval, logger: logger}
}

// Run blocks until ctx is cancelled.
func (t *Ticker) Run(ctx context.Context) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := t.reminders.DispatchDueReminders(ctx); err != nil {
				t.logger.Error("failed to dispatch due reminders", "error", err)
			}
		}
	}
}
