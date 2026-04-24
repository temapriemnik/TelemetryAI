package nats

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"agent/internal/config"
)

type ErrorNotification struct {
	ProjectID  int       `json:"project_id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type Consumer struct {
	nc    *nats.Conn
	cfg   *config.NATSConfig
	ch    chan ErrorNotification
	logger *slog.Logger
}

func NewConsumer(cfg *config.NATSConfig) *Consumer {
	return &Consumer{
		cfg:   cfg,
		ch:    make(chan ErrorNotification, 100),
		logger: slog.Default(),
	}
}

func (c *Consumer) Connect(natsURL string) error {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return err
	}
	c.nc = nc
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
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

func (c *Consumer) handleMessage(msg *nats.Msg) {
	var notification ErrorNotification
	if err := json.Unmarshal(msg.Data, &notification); err != nil {
		c.logger.Error("failed to unmarshal error notification", "error", err)
		msg.Ack()
		return
	}

	c.ch <- notification
	msg.Ack()
}

func (c *Consumer) Channel() <-chan ErrorNotification {
	return c.ch
}