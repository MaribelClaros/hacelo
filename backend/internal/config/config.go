// Package config loads the runtime configuration from environment variables.
package config

import (
	"os"
	"time"
)

type Config struct {
	Port              string
	DBPath            string
	APIToken          string
	VAPIDPublicKey    string
	VAPIDPrivateKey   string
	VAPIDSubscriber   string
	SchedulerInterval time.Duration
}

// Load reads the configuration, applying defaults for local development.
func Load() Config {
	return Config{
		Port:              getenv("PORT", "8080"),
		DBPath:            getenv("DB_PATH", "hacelo.db"),
		APIToken:          os.Getenv("API_TOKEN"),
		VAPIDPublicKey:    os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:   os.Getenv("VAPID_PRIVATE_KEY"),
		VAPIDSubscriber:   getenv("VAPID_SUBSCRIBER", "mailto:me@example.com"),
		SchedulerInterval: 30 * time.Second,
	}
}

// HasVAPIDKeys reports whether Web Push can be used.
func (c Config) HasVAPIDKeys() bool {
	return c.VAPIDPublicKey != "" && c.VAPIDPrivateKey != ""
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
