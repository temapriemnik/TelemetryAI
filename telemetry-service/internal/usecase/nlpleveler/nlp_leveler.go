package nlpleveler

import (
	"log/slog"
	"strings"

	"telemetry-service/internal/domain"
)

type ModelClient interface {
	GetLevel(log string) (level string, err error)
}

type NLPLevelerService struct {
	logger      *slog.Logger
	modelClient ModelClient
}

func NewNLPLevelerService(logger *slog.Logger, modelClient ModelClient) *NLPLevelerService {
	return &NLPLevelerService{
		logger:      logger,
		modelClient: modelClient,
	}
}

func (s *NLPLevelerService) DetectLevel(logLine string) domain.Level {
	logLine = strings.TrimSpace(logLine)
	if logLine == "" {
		s.logger.Debug("empty log line, defaulting to INFO")
		return domain.LevelInfo
	}

	s.logger.Debug("determining log level via model", "log", logLine)

	levelStr, err := s.modelClient.GetLevel(logLine)
	if err != nil {
		s.logger.Warn("model error, defaulting to INFO", "error", err.Error())
		return domain.LevelInfo
	}

	s.logger.Debug("model returned level", "level", levelStr)

	normalized := strings.ToUpper(strings.TrimSpace(levelStr))
	switch domain.Level(normalized) {
	case domain.LevelInfo, domain.LevelWarn, domain.LevelError:
		return domain.Level(normalized)
	default:
		s.logger.Warn("invalid level from model", "level", levelStr)
		return domain.LevelInfo
	}
}