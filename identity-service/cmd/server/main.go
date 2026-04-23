package main

import (
	"context"
	"log/slog"
	"net"
	httplib "net/http"
	"os"
	"os/signal"
	"syscall"

	"identity-service/internal/application/service"
	"identity-service/internal/config"
	"identity-service/internal/infrastructure/database"
	"identity-service/internal/infrastructure/repository"
	grpcserver "identity-service/internal/transport/grpc"
	httprouter "identity-service/internal/transport/http"
	"identity-service/internal/transport/http/handler"

	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.Load()
	if cfg == nil {
		log.Error("failed to load config")
		os.Exit(1)
	}

	pool, err := database.NewPool(ctx, database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	})
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewPGUserRepository(pool)
	projectRepo := repository.NewPGProjectRepository(pool)
	apiKeyRepo := repository.NewPGAPIKeyRepository(pool)

	authService := service.NewAuthService(userRepo)
	projectService := service.NewProjectService(projectRepo, apiKeyRepo)
	apiKeyService := service.NewAPIKeyService(apiKeyRepo, projectRepo)

	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	detectHandler := handler.NewDetectHandler(apiKeyService)

	httpRouter := httprouter.New(authHandler, projectHandler, detectHandler, authService, log)

	httpServer := &httplib.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: httpRouter,
	}

	grpcServer := grpc.NewServer()

	grpcAPIKeyServer := grpcserver.NewAPIKeyServer(apiKeyService, projectRepo)
	grpcAPIKeyServer.Register(grpcServer)

	grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	go func() {
		log.Info("starting gRPC server", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Error("gRPC server error", "error", err)
		}
	}()

	go func() {
		log.Info("starting HTTP server", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != httplib.ErrServerClosed {
			log.Error("HTTP server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP shutdown error", "error", err)
	}

	grpcServer.GracefulStop()
	log.Info("server stopped")
}