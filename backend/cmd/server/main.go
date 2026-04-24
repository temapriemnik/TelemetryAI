package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"telemetryai/internal/config"
	"telemetryai/internal/middleware"
	"telemetryai/internal/repository"
	"telemetryai/internal/transport"
	"telemetryai/internal/usecase"
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

	db, err := initDB(&cfg.Database)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	pool := db.Pool

	userRepo := repository.NewUserRepository(pool)
	projectRepo := repository.NewProjectRepository(pool)
	logRepo := repository.NewLogRepository(pool)

	authService := usecase.NewAuthService(userRepo, cfg.App.JWTSecret)
	projectService := usecase.NewProjectService(projectRepo)
	notificationService := usecase.NewMockNotificationService()

	nc, err := nats.Connect(cfg.NATS.URLs)
	if err != nil {
		slog.Error("failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()
	slog.Info("connected to NATS", "url", cfg.NATS.URLs)

	natsErrorPublisher := transport.NewNATSNotificationPublisher(nc, cfg.NATS.ErrorTopic)
	nlpClient := usecase.NewNLPClient(cfg.App.NLPServiceURL)
	logService := usecase.NewLogService(logRepo, projectRepo, notificationService, natsErrorPublisher, nlpClient)

	natsConsumer := transport.NewNATSConsumer(nc, logService, &cfg.NATS)
	natsCtx, natsCancel := context.WithCancel(context.Background())
	if err := natsConsumer.Start(natsCtx); err != nil {
		slog.Error("failed to start NATS consumer", "error", err)
		os.Exit(1)
	}
	defer natsCancel()

	authMiddleware := middleware.NewAuthMiddleware(authService)

	hub := transport.NewHub()
	wsHandler := transport.NewWSHandler(hub)
	logHandler := transport.NewLogHandler(logService, hub)
	authHandler := transport.NewAuthHandler(authService)
	projectHandler := transport.NewProjectHandler(projectService)

	router := transport.NewRouter(authHandler, projectHandler, logHandler, wsHandler)
	muxRouter := router.Setup(authMiddleware.Authenticate)

	go hub.Run()

	corsRouter := middleware.CORS(muxRouter)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler: corsRouter,
	}

	go func() {
		slog.Info("starting server", "host", cfg.App.Host, "port", cfg.App.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting server down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

func initDB(cfg *config.DatabaseConfig) (*repository.DB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	slog.Info("connected to database")

	migrationFile := "migrations/001_init.sql"
	if _, err := os.Stat(migrationFile); err == nil {
		if err := runMigrations(pool, migrationFile); err != nil {
			slog.Warn("migration warning", "error", err)
		} else {
			slog.Info("migrations applied")
		}
	}

	return &repository.DB{Pool: pool}, nil
}

func runMigrations(pool *pgxpool.Pool, migrationFile string) error {
	data, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	_, err = pool.Exec(context.Background(), string(data))
	return err
}