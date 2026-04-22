package rbleveler

import (
	"testing"

	"log/slog"
	"os"

	"telemetry-service/internal/domain"
)

func TestDetectLevel(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	svc := NewRBLevelerService(logger)

	tests := []struct {
		name    string
		logLine string
		want    domain.Level
		found  bool
	}{
		{"empty string", "", domain.LevelInfo, false},
		{"whitespace only", "   ", domain.LevelInfo, false},
		{"info exact", "INFO: service started", domain.LevelInfo, true},
		{"INFO uppercase", "INFO Starting", domain.LevelInfo, true},
		{"info lowercase", "info message", domain.LevelInfo, true},
		{"debug", "DEBUG: some debug", domain.LevelInfo, true},
		{"trace", "TRACE: tracing", domain.LevelInfo, true},
		{"verbose", "VERBOSE: verbose log", domain.LevelInfo, true},
		{"notice", "NOTICE: notice", domain.LevelInfo, true},
		{"config", "CONFIG: config loaded", domain.LevelInfo, true},
		{"dbg abbreviation", "[DBG] debug", domain.LevelInfo, true},
		{"trc abbreviation", "TRC trace", domain.LevelInfo, true},
		{"warn exact", "WARN: warning", domain.LevelWarn, true},
		{"warn uppercase", "WARNING message", domain.LevelWarn, true},
		{"warn lowercase", "warn: low", domain.LevelWarn, true},
		{"alert", "ALERT: alert", domain.LevelWarn, true},
		{"caution", "CAUTION: caution", domain.LevelWarn, true},
		{"attention", "ATTENTION: attention", domain.LevelWarn, true},
		{"wrg abbreviation", "[WRG] warning", domain.LevelWarn, true},
		{"error exact", "ERROR: error", domain.LevelError, true},
		{"error uppercase", "ERROR", domain.LevelError, true},
		{"error lowercase", "error msg", domain.LevelError, true},
		{"err abbreviation", "ERR: error", domain.LevelError, true},
		{"fatal", "FATAL: fatal error", domain.LevelError, true},
		{"panic", "PANIC: panic!", domain.LevelError, true},
		{"panic lowercase", "panic: occurred", domain.LevelError, true},
		{"critical", "CRITICAL: critical", domain.LevelError, true},
		{"crash", "CRASH: crash", domain.LevelError, true},
		{"severe", "SEVERE: severe", domain.LevelError, true},
		{"exception", "EXCEPTION: exception", domain.LevelError, true},
		{"fail", "FAIL: failed", domain.LevelError, true},
		{"failed", "failed operation", domain.LevelError, true},
		{"emerg", "EMERG: emergency", domain.LevelError, true},
		{"crit abbreviation", "CRIT: critical", domain.LevelError, true},
		{"halt", "HALT: halting", domain.LevelError, true},
		{"abort", "ABORT: aborted", domain.LevelError, true},
		{"aborted", "process aborted", domain.LevelError, true},
		{"mixed lowercase error", "some error occurred", domain.LevelError, true},
		{"mixed lowercase warn", "warning message", domain.LevelWarn, true},
		{"unknown not found", "random log message", domain.LevelInfo, false},
		{"log with timestamp", "2024-01-01 10:00:00 INFO starting", domain.LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := svc.DetectLevel(tt.logLine)
			if got != tt.want || found != tt.found {
				t.Errorf("DetectLevel(%q) = (%v, %v), want (%v, %v)", tt.logLine, got, found, tt.want, tt.found)
			}
		})
	}
}

func TestLevelConstants(t *testing.T) {
	if domain.LevelInfo != "INFO" {
		t.Errorf("LevelInfo = %v, want INFO", domain.LevelInfo)
	}
	if domain.LevelWarn != "WARN" {
		t.Errorf("LevelWarn = %v, want WARN", domain.LevelWarn)
	}
	if domain.LevelError != "ERROR" {
		t.Errorf("LevelError = %v, want ERROR", domain.LevelError)
	}
}