package port

import (
	"context"
	"time"

	"hacelo/internal/domain"
)

// TaskRepository persists tasks.
type TaskRepository interface {
	// Create stores a new task and fills in its ID.
	Create(ctx context.Context, task *domain.Task) error
	// FindByID returns domain.ErrNotFound when the task does not exist.
	FindByID(ctx context.Context, id int64) (*domain.Task, error)
	List(ctx context.Context, filter TaskFilter) ([]domain.Task, error)
	// Update returns domain.ErrNotFound when the task does not exist.
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id int64) error
	// FindDue returns candidate tasks whose reminder time has passed.
	FindDue(ctx context.Context, now time.Time) ([]domain.Task, error)
}

// SubscriptionRepository persists push subscriptions.
type SubscriptionRepository interface {
	Save(ctx context.Context, sub domain.PushSubscription) error
	Delete(ctx context.Context, endpoint string) error
	List(ctx context.Context) ([]domain.PushSubscription, error)
}

// Notifier delivers a reminder to the user's devices.
type Notifier interface {
	// Notify sends the reminder to every subscription and returns the
	// endpoints that are no longer valid so the caller can discard them.
	Notify(ctx context.Context, task domain.Task, subs []domain.PushSubscription) (expiredEndpoints []string, err error)
}

// Clock abstracts time so business logic can be tested deterministically.
type Clock interface {
	Now() time.Time
}
