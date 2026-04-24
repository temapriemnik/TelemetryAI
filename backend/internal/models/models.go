package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string   `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
}

type Project struct {
	ID           int       `json:"id"`
	UserID      int       `json:"user_id"`
	Name        string    `json:"name"`
	APIKey      string    `json:"api_key"`
	Architecture string   `json:"architecture,omitempty"`
	Description string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Log struct {
	ID        int       `json:"id"`
	ProjectID int      `json:"project_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)