package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	apikeysgrpc "telemetry-service/internal/transport/grpc"
	"telemetry-service/internal/transport/nats"
	"telemetry-service/internal/usecase"
	"telemetry-service/internal/usecase/nlpleveler"

	apikeys "telemetry-service/internal/storage/apikeys"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	badgerPath := getEnv("BADGER_PATH", "./badger_data")
	storage, err := apikeys.NewBadgerStorage(badgerPath)
	if err != nil {
		log.Fatal("failed to open badger: ", err)
	}
	defer storage.Close()
	logger.Info("badger storage opened", "path", badgerPath)

	levelDetection := usecase.NewLevelDetectionService(logger, &nopModelClient{})

	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	natsConsumer, err := nats.NewConsumer(nats.Config{
		URL:     natsURL,
		Subject: getEnv("NATS_SUBJECT", "raw.logs"),
	}, levelDetection, storage, logger)
	if err != nil {
		log.Fatal("failed to create nats consumer: ", err)
	}

	go func() {
		grpcServer := grpc.NewServer()
		apiKeyServer := apikeysgrpc.NewServer(storage)
		apikeysgrpc.RegisterAPIKeyServiceServer(grpcServer, apiKeyServer)
		reflection.Register(grpcServer)

		lis, err := net.Listen("tcp", ":"+getEnv("GRPC_PORT", "50051"))
		if err != nil {
			log.Fatal("failed to listen: ", err)
		}
		logger.Info("starting gRPC server", "port", getEnv("GRPC_PORT", "50051"))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
		}
	}()

	metricsPort := getEnv("METRICS_PORT", "8080")
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		http.HandleFunc("/detect", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			var req struct {
				Log    string `json:"log"`
				APIKey string `json:"api_key"`
			}
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if req.APIKey == "" {
				http.Error(w, "api_key is required", http.StatusBadRequest)
				return
			}
			projectID, err := storage.Get(context.Background(), req.APIKey)
			if err != nil {
				if err == apikeys.ErrNotFound {
					http.Error(w, "invalid api_key", http.StatusForbidden)
					return
				}
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			level := levelDetection.DetectLevel(req.Log)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"level":      string(level),
				"project_id": projectID,
			})
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
	if err := natsConsumer.Start(ctx); err != nil {
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