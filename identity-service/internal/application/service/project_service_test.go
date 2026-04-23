package service

import (
	"context"
	"testing"

	"identity-service/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *MockProjectRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Project, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entity.Project), args.Error(1)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockAPIKeyRepository struct {
	mock.Mock
}

func (m *MockAPIKeyRepository) Create(ctx context.Context, key *entity.APIKey) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.APIKey, error) {
	args := m.Called(ctx, projectID)
	return args.Get(0).([]*entity.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	args := m.Called(ctx, keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.APIKey), args.Error(1)
}

func (m *MockAPIKeyRepository) Delete(ctx context.Context, keyHash string) error {
	args := m.Called(ctx, keyHash)
	return args.Error(0)
}

func (m *MockAPIKeyRepository) List(ctx context.Context) ([]*entity.APIKey, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.APIKey), args.Error(1)
}

func TestProjectService_Create(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	projectService := NewProjectService(mockProjectRepo, mockAPIKeyRepo)

	userID := uuid.New()
	mockProjectRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Project")).Return(nil)

	project, err := projectService.Create(context.Background(), userID, "Test Project")

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, project.ID)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, userID, project.UserID)
	mockProjectRepo.AssertExpectations(t)
}

func TestProjectService_GetByID(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	projectService := NewProjectService(mockProjectRepo, mockAPIKeyRepo)

	userID := uuid.New()
	projectID := uuid.New()
	expectedProject := &entity.Project{
		ID:     projectID,
		UserID: userID,
		Name:  "Test Project",
	}

	mockProjectRepo.On("GetByID", mock.Anything, projectID).Return(expectedProject, nil)

	project, err := projectService.GetByID(context.Background(), projectID, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedProject.ID, project.ID)
	mockProjectRepo.AssertExpectations(t)
}

func TestProjectService_GetByID_AccessDenied(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	projectService := NewProjectService(mockProjectRepo, mockAPIKeyRepo)

	userID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()
	project := &entity.Project{
		ID:     projectID,
		UserID: otherUserID,
		Name:  "Test Project",
	}

	mockProjectRepo.On("GetByID", mock.Anything, projectID).Return(project, nil)

	_, err := projectService.GetByID(context.Background(), projectID, userID)

	assert.Equal(t, ErrProjectAccess, err)
	mockProjectRepo.AssertExpectations(t)
}

func TestProjectService_GetByID_NotFound(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	projectService := NewProjectService(mockProjectRepo, mockAPIKeyRepo)

	projectID := uuid.New()
	userID := uuid.New()

	mockProjectRepo.On("GetByID", mock.Anything, projectID).Return(nil, assert.AnError)

	_, err := projectService.GetByID(context.Background(), projectID, userID)

	assert.Equal(t, ErrProjectNotFound, err)
	mockProjectRepo.AssertExpectations(t)
}

func TestProjectService_GetByUserID(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	mockAPIKeyRepo := new(MockAPIKeyRepository)
	projectService := NewProjectService(mockProjectRepo, mockAPIKeyRepo)

	userID := uuid.New()
	expectedProjects := []*entity.Project{
		{ID: uuid.New(), UserID: userID, Name: "Project 1"},
		{ID: uuid.New(), UserID: userID, Name: "Project 2"},
	}

	mockProjectRepo.On("GetByUserID", mock.Anything, userID).Return(expectedProjects, nil)

	projects, err := projectService.GetByUserID(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(projects))
	mockProjectRepo.AssertExpectations(t)
}