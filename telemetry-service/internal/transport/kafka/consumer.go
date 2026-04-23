package kafka

import (
	"context"
	"log/slog"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/hamba/avro/v2"

	"telemetry-service/internal/metrics"
	"telemetry-service/internal/storage/apikeys"
	"telemetry-service/internal/usecase"
)

type Config struct {
	Brokers []string
	Group  string
	Topic  string
}

type Consumer struct {
	reader  *kafkago.Reader
	service *usecase.LevelDetectionService
	storage apikeys.Storage
	logger  *slog.Logger
	cfg     Config
}

func NewConsumer(cfg Config, service *usecase.LevelDetectionService, storage apikeys.Storage, logger *slog.Logger) (*Consumer, error) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.Group,
		Topic:    cfg.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader:  reader,
		service: service,
		storage: storage,
		logger:  logger,
		cfg:     cfg,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("kafka consumer started", "topic", c.cfg.Topic, "brokers", c.cfg.Brokers)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("shutting down kafka consumer")
			return c.reader.Close()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				c.logger.Error("kafka read error", "error", err)
				continue
			}

			if err := c.processMessage(msg); err != nil {
				c.logger.Error("failed to process message", "error", err)
			}
		}
	}
}

func (c *Consumer) processMessage(msg kafkago.Message) error {
	startTime := time.Now()

	metrics.KafkaMessagesReceived.Inc()

	var record rawLogRecord
	if err := avro.Unmarshal(rawLogSchema, msg.Value, &record); err != nil {
		metrics.ProcessingErrors.WithLabelValues("unmarshal").Inc()
		return err
	}

	c.logger.Debug("received log record", "api_key", record.APIKey, "time", record.Time)

	projectID, err := c.storage.Get(context.Background(), record.APIKey)
	if err != nil {
		if err == apikeys.ErrNotFound {
			metrics.ProcessingErrors.WithLabelValues("invalid_api_key").Inc()
			c.logger.Warn("invalid api key", "api_key", record.APIKey)
			return nil
		}
		metrics.ProcessingErrors.WithLabelValues("storage_error").Inc()
		return err
	}

	level := c.service.DetectLevel(record.LogMessage)

	metrics.ProcessedLogs.WithLabelValues(string(level)).Inc()
	metrics.ProcessingDuration.Observe(time.Since(startTime).Seconds())

	c.logger.Info("detected level", "project_id", projectID, "level", level)

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

type rawLogRecord struct {
	Time       int64  `avro:"time"`
	APIKey     string `avro:"api_key"`
	LogMessage string `avro:"log_message"`
}

var rawLogSchemaStr = `{
  "type": "record",
  "name": "RawLogRecord",
  "fields": [
    { "name": "time", "type": { "type": "long", "logicalType": "timestamp-millis" } },
    { "name": "api_key", "type": "string" },
    { "name": "log_message", "type": "string" }
  ]
}`

func init() {
	var err error
	rawLogSchema, err = avro.Parse(rawLogSchemaStr)
	if err != nil {
		panic("failed to parse avro schema: " + err.Error())
	}
}

var rawLogSchema avro.Schema