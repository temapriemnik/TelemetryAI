package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"telemetry-service/internal/transport/kafka"
	"telemetry-service/internal/usecase"
	"telemetry-service/internal/usecase/nlpleveler"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	levelDetection := usecase.NewLevelDetectionService(logger, &nopModelClient{})

	kafkaConsumer, err := kafka.NewConsumer(kafka.Config{
		Brokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		Group:   getEnv("KAFKA_GROUP", "telemetry-service"),
		Topic:   getEnv("KAFKA_TOPIC", "raw.logs"),
	}, levelDetection, logger)
	if err != nil {
		log.Fatal("failed to create kafka consumer: ", err)
	}

	metricsPort := getEnv("METRICS_PORT", "8080")
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		logger.Info("starting metrics server", "port", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil && err != http.ErrServerClosed {
			log.Fatal("metrics server error: ", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("received shutdown signal")
		cancel()
	}()

	log.Println("starting telemetry service...")
	if err := kafkaConsumer.Start(ctx); err != nil {
		log.Fatal("consumer error: ", err)
	}
}

type nopModelClient struct{}

func (c *nopModelClient) GetLevel(log string) (string, error) {
	return "", nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var _ nlpleveler.ModelClient = (*nopModelClient)(nil)