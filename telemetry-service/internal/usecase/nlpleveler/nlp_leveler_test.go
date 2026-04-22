package nlpleveler

import (
	"errors"
	"testing"

	"log/slog"
	"os"

	"telemetry-service/internal/domain"
)

type mockModelClient struct {
	resp string
	err  error
}

func (m *mockModelClient) GetLevel(log string) (string, error) {
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
		{"whitespace only defaults to INFO", "   ", "INFO", nil, domain.LevelInfo},
		{"model returns INFO", "log message", "INFO", nil, domain.LevelInfo},
		{"model returns WARN", "log", "WARN", nil, domain.LevelWarn},
		{"model returns ERROR", "log", "ERROR", nil, domain.LevelError},
		{"invalid level defaults to INFO", "log", "UNKNOWN", nil, domain.LevelInfo},
		{"model error defaults to INFO", "log", "", errors.New("model error"), domain.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockModelClient{resp: tt.mockResp, err: tt.mockErr}
			svc := NewNLPLevelerService(logger, mock)

			got := svc.DetectLevel(tt.logLine)
			if got != tt.want {
				t.Errorf("DetectLevel(%q) = %v, want %v", tt.logLine, got, tt.want)
			}
		})
	}
}