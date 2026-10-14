// Command server is the composition root: it builds the adapters, injects
// them into the application services and starts the HTTP server and the
// reminder scheduler.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hacelo/internal/adapters/inbound/httpapi"
	"hacelo/internal/adapters/inbound/scheduler"
	"hacelo/internal/adapters/outbound/clock"
	"hacelo/internal/adapters/outbound/lognotifier"
	"hacelo/internal/adapters/outbound/sqlite"
	"hacelo/internal/adapters/outbound/webpush"
	"hacelo/internal/application/port"
	"hacelo/internal/application/service"
	"hacelo/internal/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	if err := run(cfg, logger); err != nil {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, logger *slog.Logger) error {
	if cfg.APIToken == "" {
		logger.Warn("API_TOKEN is not set: the API is unauthenticated (local development only)")
	}

	// Outbound adapters.
	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	taskRepo := sqlite.NewTaskRepository(db)
	subscriptionRepo := sqlite.NewSubscriptionRepository(db)
	systemClock := clock.System{}

	var notifier port.Notifier
	if cfg.HasVAPIDKeys() {
		notifier = webpush.NewNotifier(cfg.VAPIDPublicKey, cfg.VAPIDPrivateKey, cfg.VAPIDSubscriber, logger)
	} else {
		notifier = lognotifier.New(logger)
	}

	// Application services.
	taskService := service.NewTaskService(taskRepo, systemClock)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo)
	reminderService := service.NewReminderService(taskRepo, subscriptionRepo, notifier, systemClock, logger)

	// Inbound adapters.
	handler := httpapi.NewHandler(taskService, subscriptionService, cfg.VAPIDPublicKey, logger)
	router, err := httpapi.NewRouter(handler, cfg.APIToken)
	if err != nil {
		return err
	}
	ticker := scheduler.NewTicker(reminderService, cfg.SchedulerInterval, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go ticker.Run(ctx)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("hacelo listening", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
