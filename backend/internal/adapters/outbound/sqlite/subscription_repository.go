package sqlite

import (
	"context"
	"database/sql"

	"hacelo/internal/application/port"
	"hacelo/internal/domain"
)

// SubscriptionRepository implements port.SubscriptionRepository.
type SubscriptionRepository struct {
	db *sql.DB
}

var _ port.SubscriptionRepository = (*SubscriptionRepository)(nil)

func NewSubscriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Save(ctx context.Context, sub domain.PushSubscription) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO push_subscriptions (user_id, endpoint, p256dh_key, auth_key)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(endpoint) DO UPDATE SET
			p256dh_key = excluded.p256dh_key,
			auth_key = excluded.auth_key`,
		singleUserID, sub.Endpoint, sub.P256dhKey, sub.AuthKey)
	return err
}

func (r *SubscriptionRepository) Delete(ctx context.Context, endpoint string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM push_subscriptions WHERE endpoint = ? AND user_id = ?`, endpoint, singleUserID)
	return err
}

func (r *SubscriptionRepository) List(ctx context.Context) ([]domain.PushSubscription, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT endpoint, p256dh_key, auth_key FROM push_subscriptions WHERE user_id = ?`, singleUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := []domain.PushSubscription{}
	for rows.Next() {
		var s domain.PushSubscription
		if err := rows.Scan(&s.Endpoint, &s.P256dhKey, &s.AuthKey); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}
