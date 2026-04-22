package kafka

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/hamba/avro/v2"

	"telemetry-service/internal/usecase"
)

type RawLogRecord struct {
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

type Config struct {
	Brokers string
	Group  string
	Topic  string
}

type Consumer struct {
	consumer *kafka.Consumer
	service  *usecase.LevelDetectionService
	logger   *slog.Logger
	cfg      Config
}

func NewConsumer(cfg Config, service *usecase.LevelDetectionService, logger *slog.Logger) (*Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.Brokers,
		"group.id":          cfg.Group,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		service:  service,
		logger:   logger,
		cfg:      cfg,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	if err := c.consumer.Subscribe(c.cfg.Topic, nil); err != nil {
		return err
	}

	c.logger.Info("kafka consumer started", "topic", c.cfg.Topic, "brokers", c.cfg.Brokers)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("shutting down kafka consumer")
			return c.consumer.Close()
		default:
			msg, err := c.consumer.ReadMessage(5 * time.Second)
			if err != nil {
				var kafkaErr kafka.Error
				if errors.As(err, &kafkaErr) && kafkaErr.IsTimeout() {
					continue
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

func (c *Consumer) processMessage(msg *kafka.Message) error {
	var record RawLogRecord
	if err := avro.Unmarshal(rawLogSchema, msg.Value, &record); err != nil {
		return err
	}

	c.logger.Debug("received log record", "api_key", record.APIKey, "time", record.Time)

	level := c.service.DetectLevel(record.LogMessage)

	c.logger.Info("detected level", "api_key", record.APIKey, "level", level)

	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}