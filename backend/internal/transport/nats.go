package transport

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"telemetryai/internal/config"
	"telemetryai/internal/models"
	"telemetryai/internal/usecase"
)

type NATSConsumer struct {
	nc        *nats.Conn
	logService *usecase.LogService
	cfg       *config.NATSConfig
	logger    *slog.Logger
}

type AgentLogEntry struct {
	Container  string `json:"container"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	APIKey    string `json:"api_key"`
}

type ErrorNotification struct {
	ProjectID  int       `json:"project_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func NewNATSConsumer(nc *nats.Conn, logService *usecase.LogService, cfg *config.NATSConfig) *NATSConsumer {
	return &NATSConsumer{
		nc:        nc,
		logService: logService,
		cfg:       cfg,
		logger:    slog.Default(),
	}
}

func (c *NATSConsumer) Start(ctx context.Context) error {
	sub, err := c.nc.Subscribe(c.cfg.Topic, c.handleMessage)
	if err != nil {
		return err
	}

	c.logger.Info("NATS consumer started", "topic", c.cfg.Topic)

	go func() {
		<-ctx.Done()
		if err := sub.Unsubscribe(); err != nil {
			c.logger.Error("failed to unsubscribe", "error", err)
		}
		c.nc.Close()
		c.logger.Info("NATS consumer stopped")
	}()

	return nil
}

func (c *NATSConsumer) handleMessage(msg *nats.Msg) {
	var entry AgentLogEntry
	if err := json.Unmarshal(msg.Data, &entry); err != nil {
		c.logger.Error("failed to unmarshal log entry", "error", err)
		msg.Ack()
		return
	}

	timestamp, err := time.Parse(time.RFC3339, entry.Timestamp)
	if err != nil {
		timestamp = time.Now().UTC()
	}

	input := usecase.ReceiveLogInput{
		APIKey:    entry.APIKey,
		Timestamp: timestamp,
		Message:   entry.Message,
	}

	if _, err := c.logService.Receive(input); err != nil {
		c.logger.Error("failed to process log", "error", err, "api_key", entry.APIKey)
	}

	msg.Ack()
}

type NATSNotificationPublisher struct {
	nc    *nats.Conn
	topic string
}

func NewNATSNotificationPublisher(nc *nats.Conn, topic string) *NATSNotificationPublisher {
	return &NATSNotificationPublisher{
		nc:    nc,
		topic: topic,
	}
}

func (p *NATSNotificationPublisher) Send(level models.LogLevel, projectID int, message string) {
	if level != models.LogLevelError && level != models.LogLevelWarn {
		return
	}

	notification := ErrorNotification{
		ProjectID:  projectID,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}

	data, err := json.Marshal(notification)
	if err != nil {
		slog.Error("failed to marshal error notification", "error", err)
		return
	}

	if err := p.nc.Publish(p.topic, data); err != nil {
		slog.Error("failed to publish error notification", "error", err)
		return
	}

	slog.Info("error notification published", "topic", p.topic, "project_id", projectID)
}