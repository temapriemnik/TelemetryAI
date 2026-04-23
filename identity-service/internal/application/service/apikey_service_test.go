package service

import (
	"context"
	"testing"

	"identity-service/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIKeyService_Create(t *testing.T) {
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	mockProjectRepo := new(MockProjectRepository)
	apiKeyService := NewAPIKeyService(mockAPIKeyRepo, mockProjectRepo)

	projectID := uuid.New()
	apiKey := "test-api-key"

	mockProjectRepo.On("GetByID", mock.Anything, projectID).Return(&entity.Project{ID: projectID}, nil)
	mockAPIKeyRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.APIKey")).Return(nil)

	err := apiKeyService.Create(context.Background(), projectID, apiKey)

	assert.NoError(t, err)
	mockProjectRepo.AssertExpectations(t)
	mockAPIKeyRepo.AssertExpectations(t)
}

func TestAPIKeyService_Create_ProjectNotFound(t *testing.T) {
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	mockProjectRepo := new(MockProjectRepository)
	apiKeyService := NewAPIKeyService(mockAPIKeyRepo, mockProjectRepo)

	projectID := uuid.New()
	apiKey := "test-api-key"

	mockProjectRepo.On("GetByID", mock.Anything, projectID).Return(nil, assert.AnError)

	err := apiKeyService.Create(context.Background(), projectID, apiKey)

	assert.Equal(t, ErrProjectNotFound, err)
	mockProjectRepo.AssertExpectations(t)
}

func TestAPIKeyService_Delete(t *testing.T) {
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	mockProjectRepo := new(MockProjectRepository)
	apiKeyService := NewAPIKeyService(mockAPIKeyRepo, mockProjectRepo)

	keyHash := "test-key-hash"

	mockAPIKeyRepo.On("Delete", mock.Anything, keyHash).Return(nil)

	err := apiKeyService.Delete(context.Background(), keyHash)

	assert.NoError(t, err)
	mockAPIKeyRepo.AssertExpectations(t)
}

func TestAPIKeyService_Verify(t *testing.T) {
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	mockProjectRepo := new(MockProjectRepository)
	apiKeyService := NewAPIKeyService(mockAPIKeyRepo, mockProjectRepo)

	projectID := uuid.New()
	keyHash := "test-key-hash"
	expectedKey := &entity.APIKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		KeyHash:  keyHash,
		Level:   "standard",
	}

	mockAPIKeyRepo.On("GetByHash", mock.Anything, keyHash).Return(expectedKey, nil)

	key, err := apiKeyService.Verify(context.Background(), keyHash)

	assert.NoError(t, err)
	assert.Equal(t, expectedKey.ID, key.ID)
	mockAPIKeyRepo.AssertExpectations(t)
}

func TestAPIKeyService_Verify_NotFound(t *testing.T) {
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	mockProjectRepo := new(MockProjectRepository)
	apiKeyService := NewAPIKeyService(mockAPIKeyRepo, mockProjectRepo)

	keyHash := "invalid-key"

	mockAPIKeyRepo.On("GetByHash", mock.Anything, keyHash).Return(nil, assert.AnError)

	_, err := apiKeyService.Verify(context.Background(), keyHash)

	assert.Error(t, err)
	mockAPIKeyRepo.AssertExpectations(t)
}

func TestGenerateAPIKey(t *testing.T) {
	key1 := GenerateAPIKey()
	key2 := GenerateAPIKey()

	assert.NotEmpty(t, key1)
	assert.NotEqual(t, key1, key2)
	assert.Len(t, key1, 32) // 16 bytes = 32 hex chars
}