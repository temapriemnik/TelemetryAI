package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"identity-service/internal/domain/entity"
	"identity-service/internal/domain/repository"

	"github.com/google/uuid"
)

var ErrAPIKeyNotFound = errors.New("api key not found")

type APIKeyService struct {
	apiKeyRepo repository.APIKeyRepository
	projectRepo repository.ProjectRepository
}

func NewAPIKeyService(apiKeyRepo repository.APIKeyRepository, projectRepo repository.ProjectRepository) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo:   apiKeyRepo,
		projectRepo: projectRepo,
	}
}

func (s *APIKeyService) Create(ctx context.Context, projectID uuid.UUID, apiKey string) error {
	_, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return ErrProjectNotFound
	}

	key := &entity.APIKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		KeyHash:  apiKey,
		Level:    "standard",
		CreateAt: time.Now(),
	}

	return s.apiKeyRepo.Create(ctx, key)
}

func (s *APIKeyService) Delete(ctx context.Context, keyHash string) error {
	return s.apiKeyRepo.Delete(ctx, keyHash)
}

func (s *APIKeyService) List(ctx context.Context) ([]*entity.APIKey, error) {
	return s.apiKeyRepo.List(ctx)
}

type APIKeyPair struct {
	APIKey    string    `json:"api_key"`
	ProjectID uuid.UUID `json:"project_id"`
	Level     string    `json:"level"`
}

func (s *APIKeyService) ListAll(ctx context.Context) ([]APIKeyPair, error) {
	keys, err := s.apiKeyRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	var result []APIKeyPair
	for _, k := range keys {
		result = append(result, APIKeyPair{
			APIKey:    k.KeyHash,
			ProjectID: k.ProjectID,
			Level:    k.Level,
		})
	}
	return result, nil
}

func (s *APIKeyService) Verify(ctx context.Context, apiKey string) (*entity.APIKey, error) {
	return s.apiKeyRepo.GetByHash(ctx, apiKey)
}

func GenerateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}