package usecase

import (
	"log/slog"
	"strings"

	"telemetry-service/internal/domain"
	"telemetry-service/internal/usecase/nlpleveler"
	"telemetry-service/internal/usecase/rbleveler"
)

type LevelDetectionService struct {
	logger      *slog.Logger
	rbLeveler  *rbleveler.RBLevelerService
	nlpLeveler *nlpleveler.NLPLevelerService
}

func NewLevelDetectionService(logger *slog.Logger, nlpModelClient nlpleveler.ModelClient) *LevelDetectionService {
	return &LevelDetectionService{
		logger:      logger,
		rbLeveler:   rbleveler.NewRBLevelerService(logger),
		nlpLeveler: nlpleveler.NewNLPLevelerService(logger, nlpModelClient),
	}
}

func (s *LevelDetectionService) DetectLevel(logLine string) domain.Level {
	logLine = strings.TrimSpace(logLine)
	if logLine == "" {
		s.logger.Debug("empty log line, defaulting to INFO")
		return domain.LevelInfo
	}

	s.logger.Debug("attempting regex detection", "log", logLine)

	detected, found := s.rbLeveler.DetectLevel(logLine)
	if found {
		s.logger.Debug("regex detected level", "level", detected)
		return detected
	}

	s.logger.Debug("regex failed, falling back to NLP model")
	return s.nlpLeveler.DetectLevel(logLine)
}