package repository

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"telemetryai/internal/models"
)

type logRepository struct {
	pool *pgxpool.Pool
}

func NewLogRepository(pool *pgxpool.Pool) LogRepository {
	return &logRepository{pool: pool}
}

func (r *logRepository) Create(log *models.Log) error {
	query := `
		INSERT INTO logs (project_id, timestamp, level, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	slog.Debug("inserting log", "project_id", log.ProjectID, "level", log.Level)
	
	err := r.pool.QueryRow(context.Background(), query,
		log.ProjectID, log.Timestamp, log.Level, log.Message).
		Scan(&log.ID, &log.CreatedAt)
	if err != nil {
		slog.Error("failed to insert log", "error", err, "project_id", log.ProjectID)
		return err
	}
	
	slog.Debug("log inserted", "log_id", log.ID)
	return nil
}

func (r *logRepository) GetByProjectID(projectID int) ([]*models.Log, error) {
	query := `SELECT id, project_id, timestamp, level, message, created_at 
		FROM logs WHERE project_id = $1 ORDER BY timestamp DESC LIMIT 100`

	slog.Debug("querying logs", "project_id", projectID)
	
	rows, err := r.pool.Query(context.Background(), query, projectID)
	if err != nil {
		slog.Error("query error", "error", err)
		return nil, err
	}
	defer rows.Close()

	var logs []*models.Log
	for rows.Next() {
		var log models.Log
		if err := rows.Scan(&log.ID, &log.ProjectID, &log.Timestamp, &log.Level, &log.Message, &log.CreatedAt); err != nil {
			slog.Error("scan error", "error", err)
			return nil, err
		}
		logs = append(logs, &log)
	}

	slog.Debug("logs result", "project_id", projectID, "count", len(logs))
	
	if logs == nil {
		return []*models.Log{}, nil
	}

	return logs, nil
}

func ParseLogLevel(message string) models.LogLevel {
	lower := func(s string) string {
		result := make([]byte, len(s))
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			result[i] = c
		}
		return string(result)
	}(message)

	switch {
	case contains(lower, "error"):
		return models.LogLevelError
	case contains(lower, "warn"), contains(lower, "warning"):
		return models.LogLevelWarn
	default:
		return models.LogLevelInfo
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}