package nats

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"telemetry-service/internal/metrics"
	"telemetry-service/internal/storage/apikeys"
	"telemetry-service/internal/usecase"
)

type Config struct {
	URL     string
	Subject string
}

type Consumer struct {
	sub     *nats.Subscription
	nc      *nats.Conn
	js      nats.JetStreamContext
	service *usecase.LevelDetectionService
	storage apikeys.Storage
	logger  *slog.Logger
	cfg     Config
}

func NewConsumer(cfg Config, service *usecase.LevelDetectionService, storage apikeys.Storage, logger *slog.Logger) (*Consumer, error) {
	nc, err := nats.Connect(cfg.URL)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:    "raw-logs",
		Subjects: []string{cfg.Subject},
	})
	if err != nil {
		js, _ = nc.JetStream()
	}

	sub, err := js.Subscribe(cfg.Subject, func(msg *nats.Msg) {
		if err := process(msg, service, storage, logger); err != nil {
			logger.Error("failed to process message", "error", err)
		}
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{
		sub:     sub,
		nc:      nc,
		js:      js,
		service: service,
		storage: storage,
		logger:  logger,
		cfg:     cfg,
	}, nil
}

func process(msg *nats.Msg, service *usecase.LevelDetectionService, storage apikeys.Storage, logger *slog.Logger) error {
	startTime := time.Now()
	metrics.NatsMessagesReceived.Inc()

	var record rawLogRecord
	if err := json.Unmarshal(msg.Data, &record); err != nil {
		metrics.ProcessingErrors.WithLabelValues("unmarshal").Inc()
		return err
	}

	logger.Debug("received log record", "api_key", record.APIKey, "time", record.Time)

	projectID, err := storage.Get(context.Background(), record.APIKey)
	if err != nil {
		if err == apikeys.ErrNotFound {
			metrics.ProcessingErrors.WithLabelValues("invalid_api_key").Inc()
			logger.Warn("invalid api key", "api_key", record.APIKey)
			return nil
		}
		metrics.ProcessingErrors.WithLabelValues("storage_error").Inc()
		return err
	}

	level := service.DetectLevel(record.LogMessage)

	metrics.ProcessedLogs.WithLabelValues(string(level)).Inc()
	metrics.ProcessingDuration.Observe(time.Since(startTime).Seconds())

	logger.Info("detected level", "project_id", projectID, "level", level)

	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("nats consumer started", "subject", c.cfg.Subject, "url", c.cfg.URL)
	<-ctx.Done()
	return c.Close()
}

func (c *Consumer) Close() error {
	if c.sub != nil {
		c.sub.Unsubscribe()
	}
	if c.nc != nil {
		c.nc.Close()
	}
	return nil
}

type rawLogRecord struct {
	Time       int64  `json:"time"`
	APIKey     string `json:"api_key"`
	LogMessage string `json:"log_message"`
}