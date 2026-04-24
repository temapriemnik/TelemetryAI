package handler

import (
	"log/slog"
	"time"

	"agent/internal/backend"
	"agent/internal/nats"
	"agent/internal/smtp"
)

type ErrorHandler struct {
	backendClient *backend.Client
	smtpSender   *smtp.Sender
	logger      *slog.Logger
}

func NewErrorHandler(backendURL string, smtpHost string, smtpPort int, smtpFrom string) *ErrorHandler {
	return &ErrorHandler{
		backendClient: backend.NewClient(backendURL),
		smtpSender:   smtp.NewSender(smtpHost, smtpPort, smtpFrom),
		logger:      slog.Default(),
	}
}

func (h *ErrorHandler) Start(ch <-chan nats.ErrorNotification) {
	go h.process(ch)
}

func (h *ErrorHandler) process(ch <-chan nats.ErrorNotification) {
	for notification := range ch {
		h.handle(notification)
	}
}

func (h *ErrorHandler) handle(notification nats.ErrorNotification) {
	h.logger.Info("processing error notification", "project_id", notification.ProjectID)

	alertData, err := h.backendClient.GetAlertData(notification.ProjectID)
	if err != nil {
		h.logger.Error("failed to get alert data", "error", err, "project_id", notification.ProjectID)
		return
	}

	if alertData.UserEmail == "" {
		h.logger.Warn("no user email found", "project_id", notification.ProjectID)
		return
	}

	timestamp := notification.Timestamp.Format(time.RFC3339)

	err = h.smtpSender.SendErrorAlert(
		alertData.UserEmail,
		alertData.Name,
		notification.Message,
		timestamp,
	)
	if err != nil {
		h.logger.Error("failed to send email", "error", err, "email", alertData.UserEmail)
		return
	}

	h.logger.Info("alert sent", "email", alertData.UserEmail, "project", alertData.Name)
}