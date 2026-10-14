// Package webpush implements port.Notifier using the Web Push protocol
// (VAPID), delivering notifications to the PWA's service worker.
package webpush

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	webpushlib "github.com/SherClockHolmes/webpush-go"

	"hacelo/internal/application/port"
	"hacelo/internal/domain"
)

// Action identifiers understood by the service worker (web/public/sw.js).
const (
	actionSnooze     = "snooze"
	actionDone       = "done"
	actionReschedule = "reschedule"
)

// payload is the JSON received by the service worker's "push" event.
type payload struct {
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	TaskID  int64    `json:"task_id"`
	Actions []action `json:"actions"`
}

type action struct {
	Action string `json:"action"`
	Title  string `json:"title"` // button label shown to the user
}

// Notifier implements port.Notifier.
type Notifier struct {
	publicKey  string
	privateKey string
	subscriber string // mailto: or https: contact required by VAPID
	logger     *slog.Logger
}

var _ port.Notifier = (*Notifier)(nil)

func NewNotifier(publicKey, privateKey, subscriber string, logger *slog.Logger) *Notifier {
	return &Notifier{publicKey: publicKey, privateKey: privateKey, subscriber: subscriber, logger: logger}
}

// buildPayload picks the buttons according to the task state: the first
// reminder offers "can't now" / "on it", the follow-up asks whether it is
// done. Title, body and button labels are shown to the user, so they stay
// in Spanish.
func buildPayload(task domain.Task) payload {
	if task.Status == domain.StatusVerifying {
		return payload{
			Title:  "¿Ya lo hiciste?",
			Body:   task.Description,
			TaskID: task.ID,
			Actions: []action{
				{Action: actionSnooze, Title: "Ya lo hago"},
				{Action: actionDone, Title: "Ya lo hice"},
			},
		}
	}
	return payload{
		Title:  "Tenés una tarea pendiente",
		Body:   task.Description,
		TaskID: task.ID,
		Actions: []action{
			{Action: actionReschedule, Title: "Ahora no puedo"},
			{Action: actionSnooze, Title: "Ya lo hago"},
		},
	}
}

func (n *Notifier) Notify(ctx context.Context, task domain.Task, subs []domain.PushSubscription) ([]string, error) {
	body, err := json.Marshal(buildPayload(task))
	if err != nil {
		return nil, fmt.Errorf("encoding payload: %w", err)
	}

	var expired []string
	for _, sub := range subs {
		resp, err := webpushlib.SendNotification(body, &webpushlib.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpushlib.Keys{
				P256dh: sub.P256dhKey,
				Auth:   sub.AuthKey,
			},
		}, &webpushlib.Options{
			Subscriber:      n.subscriber,
			VAPIDPublicKey:  n.publicKey,
			VAPIDPrivateKey: n.privateKey,
			TTL:             300,
		})
		if err != nil {
			n.logger.Error("failed to push notification", "endpoint", sub.Endpoint, "error", err)
			continue
		}
		resp.Body.Close()

		// 404/410 mean the browser dropped the subscription (app removed,
		// storage cleared...), so it can be discarded.
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
			expired = append(expired, sub.Endpoint)
		}
	}
	return expired, nil
}
