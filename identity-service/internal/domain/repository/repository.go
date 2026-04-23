package repository

import (
	"context"

	"identity-service/internal/domain/entity"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByAuthToken(ctx context.Context, token string) (*entity.User, error)
}

type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type APIKeyRepository interface {
	Create(ctx context.Context, key *entity.APIKey) error
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.APIKey, error)
	GetByHash(ctx context.Context, keyHash string) (*entity.APIKey, error)
	Delete(ctx context.Context, keyHash string) error
	List(ctx context.Context) ([]*entity.APIKey, error)
}