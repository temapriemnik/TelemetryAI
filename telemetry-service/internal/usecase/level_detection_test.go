package usecase

import (
	"errors"
	"testing"

	"log/slog"
	"os"

	"telemetry-service/internal/domain"
)

type mockNLPClient struct {
	resp string
	err  error
}

func (m *mockNLPClient) GetLevel(log string) (string, error) {
	return m.resp, m.err
}

func TestDetectLevel(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	tests := []struct {
		name     string
		logLine  string
		mockResp string
		mockErr  error
		want     domain.Level
	}{
		{"empty string defaults to INFO", "", "INFO", nil, domain.LevelInfo},
		{"regex detects INFO", "INFO: service started", "INFO", nil, domain.LevelInfo},
		{"regex detects WARN", "WARN: warning", "INFO", nil, domain.LevelWarn},
		{"regex detects ERROR", "ERROR: error occurred", "INFO", nil, domain.LevelError},
		{"regex fails nlp returns WARN", "some random log", "WARN", nil, domain.LevelWarn},
		{"regex fails nlp returns ERROR", "another random log", "ERROR", nil, domain.LevelError},
		{"nlp model error defaults to INFO", "random log", "INFO", errors.New("model error"), domain.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNLPClient{resp: tt.mockResp, err: tt.mockErr}
			svc := NewLevelDetectionService(logger, mock)

			got := svc.DetectLevel(tt.logLine)
			if got != tt.want {
				t.Errorf("DetectLevel(%q) = %v, want %v", tt.logLine, got, tt.want)
			}
		})
	}
}