package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"hacelo/internal/application/port"
	"hacelo/internal/domain"
)

// TaskRepository implements port.TaskRepository.
type TaskRepository struct {
	db *sql.DB
}

var _ port.TaskRepository = (*TaskRepository)(nil)

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

const taskColumns = `id, description, urgency, status, original_time,
	snooze_minutes, next_reminder_at, verification_round, last_sent_at,
	created_at, updated_at, completed_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner) (*domain.Task, error) {
	var t domain.Task
	var originalTime, createdAt, updatedAt string
	var nextReminderAt, lastSentAt, completedAt sql.NullString

	err := row.Scan(
		&t.ID, &t.Description, &t.Urgency, &t.Status, &originalTime,
		&t.SnoozeMinutes, &nextReminderAt, &t.VerificationRound, &lastSentAt,
		&createdAt, &updatedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	if t.OriginalTime, err = parseTime(originalTime); err != nil {
		return nil, fmt.Errorf("parsing original_time: %w", err)
	}
	if t.CreatedAt, err = parseTime(createdAt); err != nil {
		return nil, fmt.Errorf("parsing created_at: %w", err)
	}
	if t.UpdatedAt, err = parseTime(updatedAt); err != nil {
		return nil, fmt.Errorf("parsing updated_at: %w", err)
	}
	if t.NextReminderAt, err = parseNullableTime(nextReminderAt); err != nil {
		return nil, fmt.Errorf("parsing next_reminder_at: %w", err)
	}
	if t.LastSentAt, err = parseNullableTime(lastSentAt); err != nil {
		return nil, fmt.Errorf("parsing last_sent_at: %w", err)
	}
	if t.CompletedAt, err = parseNullableTime(completedAt); err != nil {
		return nil, fmt.Errorf("parsing completed_at: %w", err)
	}
	return &t, nil
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (user_id, description, urgency, status, original_time,
			snooze_minutes, next_reminder_at, verification_round, last_sent_at,
			created_at, updated_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		singleUserID, task.Description, string(task.Urgency), string(task.Status),
		formatTime(task.OriginalTime), task.SnoozeMinutes,
		formatNullableTime(task.NextReminderAt), task.VerificationRound,
		formatNullableTime(task.LastSentAt), formatTime(task.CreatedAt),
		formatTime(task.UpdatedAt), formatNullableTime(task.CompletedAt),
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = id
	return nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id int64) (*domain.Task, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT `+taskColumns+` FROM tasks WHERE id = ? AND user_id = ?`, id, singleUserID)
	task, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return task, err
}

func (r *TaskRepository) List(ctx context.Context, filter port.TaskFilter) ([]domain.Task, error) {
	query := `SELECT ` + taskColumns + ` FROM tasks WHERE user_id = ?`
	args := []any{singleUserID}

	if filter.Date != "" {
		query += ` AND date(original_time) = date(?)`
		args = append(args, filter.Date)
	}
	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, string(filter.Status))
	}
	query += ` ORDER BY original_time ASC`

	return r.queryTasks(ctx, query, args...)
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET description = ?, urgency = ?, status = ?, original_time = ?,
			snooze_minutes = ?, next_reminder_at = ?, verification_round = ?,
			last_sent_at = ?, updated_at = ?, completed_at = ?
		WHERE id = ? AND user_id = ?`,
		task.Description, string(task.Urgency), string(task.Status),
		formatTime(task.OriginalTime), task.SnoozeMinutes,
		formatNullableTime(task.NextReminderAt), task.VerificationRound,
		formatNullableTime(task.LastSentAt), formatTime(task.UpdatedAt),
		formatNullableTime(task.CompletedAt), task.ID, singleUserID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM tasks WHERE id = ? AND user_id = ?`, id, singleUserID)
	return err
}

func (r *TaskRepository) FindDue(ctx context.Context, now time.Time) ([]domain.Task, error) {
	return r.queryTasks(ctx, `
		SELECT `+taskColumns+` FROM tasks
		WHERE user_id = ?
			AND status != ?
			AND next_reminder_at IS NOT NULL
			AND next_reminder_at <= ?
			AND (last_sent_at IS NULL OR last_sent_at < next_reminder_at)`,
		singleUserID, string(domain.StatusDone), formatTime(now))
}

func (r *TaskRepository) queryTasks(ctx context.Context, query string, args ...any) ([]domain.Task, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []domain.Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}
	return tasks, rows.Err()
}
