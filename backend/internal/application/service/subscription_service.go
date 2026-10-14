package service

import (
	"context"
	"fmt"

	"hacelo/internal/application/port"
	"hacelo/internal/domain"
)

// SubscriptionService implements port.SubscriptionUseCases.
type SubscriptionService struct {
	subs port.SubscriptionRepository
}

var _ port.SubscriptionUseCases = (*SubscriptionService)(nil)

func NewSubscriptionService(subs port.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{subs: subs}
}

func (s *SubscriptionService) Subscribe(ctx context.Context, sub domain.PushSubscription) error {
	if sub.Endpoint == "" || sub.P256dhKey == "" || sub.AuthKey == "" {
		return fmt.Errorf("%w: subscription endpoint and keys are required", domain.ErrInvalid)
	}
	return s.subs.Save(ctx, sub)
}

func (s *SubscriptionService) Unsubscribe(ctx context.Context, endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("%w: endpoint is required", domain.ErrInvalid)
	}
	return s.subs.Delete(ctx, endpoint)
}
