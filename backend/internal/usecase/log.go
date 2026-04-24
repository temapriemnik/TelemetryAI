package usecase

import (
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"telemetryai/internal/models"
	"telemetryai/internal/repository"
)

var (
	ErrProjectNotFoundByKey = errors.New("project not found by api key")
)

type NotificationService interface {
	Send(level models.LogLevel, message string)
}

type LogService struct {
	logRepo       repository.LogRepository
	projectRepo   repository.ProjectRepository
	notification  NotificationService
	logLevelRegex *regexp.Regexp
	logger        *slog.Logger
}

func NewLogService(
	logRepo repository.LogRepository,
	projectRepo repository.ProjectRepository,
	notification NotificationService,
) *LogService {
	regex := regexp.MustCompile(`(?i)(error|warn|warning)`)
	return &LogService{
		logRepo:       logRepo,
		projectRepo:   projectRepo,
		notification:  notification,
		logLevelRegex: regex,
	}
}

type ReceiveLogInput struct {
	APIKey   string    `json:"api_key"`
	Timestamp time.Time `json:"timestamp"`
	Message string    `json:"message"`
}

type LogOutput struct {
	ID        int       `json:"id"`
	ProjectID int      `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

func (s *LogService) Receive(input ReceiveLogInput) (LogOutput, error) {
	slog.Debug("receiving log", "api_key", input.APIKey, "message", input.Message)
	
	project, err := s.projectRepo.GetByAPIKey(input.APIKey)
	if err != nil {
		slog.Error("failed to get project by api key", "error", err, "api_key", input.APIKey)
		return LogOutput{}, err
	}
	if project == nil {
		slog.Warn("project not found by api key", "api_key", input.APIKey)
		return LogOutput{}, ErrProjectNotFoundByKey
	}

	level := s.parseLogLevel(input.Message)

	logEntry := &models.Log{
		ProjectID: project.ID,
		Timestamp: input.Timestamp,
		Level:     string(level),
		Message:   input.Message,
	}

	if err := s.logRepo.Create(logEntry); err != nil {
		slog.Error("failed to create log", "error", err, "project_id", project.ID)
		return LogOutput{}, err
	}

	if level == models.LogLevelError || level == models.LogLevelWarn {
		s.notification.Send(level, input.Message)
	}

	slog.Info("log received", "project_id", project.ID, "log_id", logEntry.ID, "level", level)
	return LogOutput{
		ID:        logEntry.ID,
		ProjectID: project.ID,
		Timestamp: logEntry.Timestamp,
		Level:     string(level),
		Message:   logEntry.Message,
	}, nil
}

func (s *LogService) parseLogLevel(message string) models.LogLevel {
	re := regexp.MustCompile(`^\[(\w+)\]`)
	matches := re.FindStringSubmatch(message)
	if len(matches) > 1 {
		level := strings.ToLower(matches[1])
		switch level {
		case "error":
			return models.LogLevelError
		case "warn", "warning":
			return models.LogLevelWarn
		case "info":
			return models.LogLevelInfo
		}
	}

	idx := strings.Index(message, "[")
	if idx != -1 {
		message = message[idx:]
	}

	matches = s.logLevelRegex.FindStringSubmatch(message)
	if len(matches) == 0 {
		return models.LogLevelInfo
	}

	level := strings.ToLower(matches[1])
	switch level {
	case "error":
		return models.LogLevelError
	case "warn", "warning":
		return models.LogLevelWarn
	}

	return models.LogLevelInfo
}

func (s *LogService) GetByProjectID(projectID, userID int) ([]LogOutput, error) {
	slog.Debug("getting logs for project", "project_id", projectID, "user_id", userID)
	
	project, err := s.projectRepo.GetByID(projectID)
	if err != nil {
		slog.Error("failed to get project", "error", err, "project_id", projectID)
		return nil, err
	}
	if project == nil {
		slog.Warn("project not found", "project_id", projectID)
		return nil, ErrProjectNotFoundByKey
	}
	slog.Debug("project found", "project_id", projectID, "project_user_id", project.UserID)
	
	if project.UserID != userID {
		slog.Warn("access denied to project", "project_id", projectID, "user_id", userID, "project_user_id", project.UserID)
		return nil, ErrProjectForbidden
	}
	
	logs, err := s.logRepo.GetByProjectID(projectID)
	if err != nil {
		slog.Error("failed to get logs", "error", err, "project_id", projectID)
		return nil, err
	}

	slog.Debug("logs fetched", "project_id", projectID, "count", len(logs))
	
	result := make([]LogOutput, len(logs))
	for i, l := range logs {
		result[i] = LogOutput{
			ID:        l.ID,
			ProjectID: l.ProjectID,
			Timestamp: l.Timestamp,
			Level:     l.Level,
			Message:   l.Message,
		}
	}

	return result, nil
}

type MockNotificationService struct{}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{}
}

func (s *MockNotificationService) Send(level models.LogLevel, message string) {
	slog.Info("notification sent", "level", level, "message", message)
}