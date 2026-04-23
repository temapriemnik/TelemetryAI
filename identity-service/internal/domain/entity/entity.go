package entity

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Password  string    `json:"-"`
	AuthToken string    `json:"auth_token,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID        uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	Name     string   `json:"name"`
	APIKey   string   `json:"api_key,omitempty"`
	CreateAt time.Time `json:"created_at"`
}

type APIKey struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	KeyHash  string   `json:"-"`
	Level    string   `json:"level"`
	CreateAt time.Time `json:"created_at"`
}