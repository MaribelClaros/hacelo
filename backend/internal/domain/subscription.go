package domain

// PushSubscription identifies a browser/device that can receive Web Push
// notifications.
type PushSubscription struct {
	Endpoint  string
	P256dhKey string
	AuthKey   string
}
