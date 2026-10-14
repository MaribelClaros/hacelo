package service

import (
	"context"
	"log/slog"

	"hacelo/internal/application/port"
)

// ReminderService implements port.ReminderUseCases.
type ReminderService struct {
	tasks    port.TaskRepository
	subs     port.SubscriptionRepository
	notifier port.Notifier
	clock    port.Clock
	logger   *slog.Logger
}

var _ port.ReminderUseCases = (*ReminderService)(nil)

func NewReminderService(
	tasks port.TaskRepository,
	subs port.SubscriptionRepository,
	notifier port.Notifier,
	clock port.Clock,
	logger *slog.Logger,
) *ReminderService {
	return &ReminderService{tasks: tasks, subs: subs, notifier: notifier, clock: clock, logger: logger}
}

func (s *ReminderService) DispatchDueReminders(ctx context.Context) error {
	now := s.clock.Now()

	candidates, err := s.tasks.FindDue(ctx, now)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	subscriptions, err := s.subs.List(ctx)
	if err != nil {
		return err
	}

	for i := range candidates {
		task := candidates[i]
		if !task.IsDue(now) {
			continue
		}

		expired, err := s.notifier.Notify(ctx, task, subscriptions)
		if err != nil {
			s.logger.Error("failed to send reminder", "task_id", task.ID, "error", err)
			continue
		}
		for _, endpoint := range expired {
			if err := s.subs.Delete(ctx, endpoint); err != nil {
				s.logger.Error("failed to delete expired subscription", "endpoint", endpoint, "error", err)
			}
		}

		task.MarkReminderSent(now)
		if err := s.tasks.Update(ctx, &task); err != nil {
			s.logger.Error("failed to record sent reminder", "task_id", task.ID, "error", err)
		}
	}
	return nil
}
