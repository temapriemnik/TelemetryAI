package rbleveler

import (
	"log/slog"
	"regexp"
	"strings"

	"telemetry-service/internal/domain"
)

type RBLevelerService struct {
	logger   *slog.Logger
	infoRe   *regexp.Regexp
	warnRe   *regexp.Regexp
	errorRe *regexp.Regexp
}

func NewRBLevelerService(logger *slog.Logger) *RBLevelerService {
	return &RBLevelerService{
		logger: logger,
		infoRe: regexp.MustCompile(`(?i)\b(info|information|debug|trace|verbose|notice|config|dbg|trc|inf)\b`),
		warnRe: regexp.MustCompile(`(?i)\b(warn|warning|alert|caution|attention|wrn|wrg)\b`),
		errorRe: regexp.MustCompile(`(?i)\b(error|err|fatal|panic|critical|crash|severe|exception|fail|failed|exception|emerg|crit|crash|halt|fatal|abort|aborted)\b`),
	}
}

func (s *RBLevelerService) DetectLevel(logLine string) (domain.Level, bool) {
	logLine = strings.TrimSpace(logLine)
	if logLine == "" {
		s.logger.Debug("empty log line provided")
		return domain.LevelInfo, false
	}

	switch {
	case s.errorRe.MatchString(logLine):
		s.logger.Debug("detected ERROR level", "match", s.errorRe.FindString(logLine))
		return domain.LevelError, true
	case s.warnRe.MatchString(logLine):
		s.logger.Debug("detected WARN level", "match", s.warnRe.FindString(logLine))
		return domain.LevelWarn, true
	case s.infoRe.MatchString(logLine):
		s.logger.Debug("detected INFO level", "match", s.infoRe.FindString(logLine))
		return domain.LevelInfo, true
	default:
		s.logger.Debug("no level detected by regex")
		return domain.LevelInfo, false
	}
}