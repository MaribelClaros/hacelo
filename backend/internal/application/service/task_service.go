// Package service implements the application use cases on top of the
// domain and the outbound ports.
package service

import (
	"context"

	"hacelo/internal/application/port"
	"hacelo/internal/domain"
)

// TaskService implements port.TaskUseCases.
type TaskService struct {
	tasks port.TaskRepository
	clock port.Clock
}

var _ port.TaskUseCases = (*TaskService)(nil)

func NewTaskService(tasks port.TaskRepository, clock port.Clock) *TaskService {
	return &TaskService{tasks: tasks, clock: clock}
}

func (s *TaskService) CreateTask(ctx context.Context, cmd port.CreateTaskCommand) (*domain.Task, error) {
	task, err := domain.NewTask(cmd.Description, cmd.Urgency, cmd.ScheduledAt, cmd.SnoozeMinutes, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.tasks.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) GetTask(ctx context.Context, id int64) (*domain.Task, error) {
	return s.tasks.FindByID(ctx, id)
}

func (s *TaskService) ListTasks(ctx context.Context, filter port.TaskFilter) ([]domain.Task, error) {
	return s.tasks.List(ctx, filter)
}

func (s *TaskService) UpdateTask(ctx context.Context, id int64, cmd port.UpdateTaskCommand) (*domain.Task, error) {
	task, err := s.tasks.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	if err := task.UpdateDetails(cmd.Description, cmd.Urgency, cmd.SnoozeMinutes, now); err != nil {
		return nil, err
	}
	if cmd.NextReminderAt != nil {
		task.Reschedule(*cmd.NextReminderAt, now)
	}

	if err := s.tasks.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	return s.tasks.Delete(ctx, id)
}

func (s *TaskService) SnoozeTask(ctx context.Context, id int64) (*domain.Task, error) {
	task, err := s.tasks.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := task.Snooze(s.clock.Now()); err != nil {
		return nil, err
	}
	if err := s.tasks.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) CompleteTask(ctx context.Context, id int64) (*domain.Task, error) {
	task, err := s.tasks.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	task.Complete(s.clock.Now())
	if err := s.tasks.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}
