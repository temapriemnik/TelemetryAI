package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"agent/internal/config"
	"agent/internal/handler"
	"agent/internal/nats"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	consumer := nats.NewConsumer(&cfg.NATS)
	if err := consumer.Connect(cfg.NATS.URLs); err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to NATS", "url", cfg.NATS.URLs)

	errorHandler := handler.NewErrorHandler(
		cfg.Backend.URL,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.From,
	)

	natsChan := consumer.Channel()
	errorHandler.Start(natsChan)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		slog.Error("failed to start NATS consumer", "error", err)
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	cancel()

	slog.Info("agent stopped")
}