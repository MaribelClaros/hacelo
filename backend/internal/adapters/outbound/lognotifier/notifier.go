// Package lognotifier is a port.Notifier that only logs. It is used when
// VAPID keys are not configured (e.g. local development).
package lognotifier

import (
	"context"
	"log/slog"

	"hacelo/internal/application/port"
	"hacelo/internal/domain"
)

type Notifier struct {
	logger *slog.Logger
}

var _ port.Notifier = (*Notifier)(nil)

func New(logger *slog.Logger) *Notifier {
	return &Notifier{logger: logger}
}

func (n *Notifier) Notify(_ context.Context, task domain.Task, subs []domain.PushSubscription) ([]string, error) {
	n.logger.Info("reminder not pushed (VAPID keys not configured)",
		"task_id", task.ID, "status", string(task.Status), "subscriptions", len(subs))
	return nil, nil
}
