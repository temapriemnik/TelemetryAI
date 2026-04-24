package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"telemetryai/internal/models"
)

type MockProjectRepoForTest struct {
	mock.Mock
}

func (m *MockProjectRepoForTest) Create(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepoForTest) GetByID(id int) (*models.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepoForTest) GetByUserID(userID int) ([]*models.Project, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockProjectRepoForTest) GetByAPIKey(apiKey string) (*models.Project, error) {
	args := m.Called(apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepoForTest) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProjectRepoForTest) Update(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func TestProjectService_Create(t *testing.T) {
	mockRepo := new(MockProjectRepoForTest)
	projectService := NewProjectService(mockRepo)

	mockRepo.On("Create", mock.AnythingOfType("*models.Project")).Return(nil).Run(func(args mock.Arguments) {
		project := args.Get(0).(*models.Project)
		project.ID = 1
		project.CreatedAt = time.Now()
	})

	input := CreateProjectInput{
		UserID: 1,
		Name:   "Test Project",
	}

	output, err := projectService.Create(input)
	require.NoError(t, err)
	assert.Equal(t, "Test Project", output.Name)
	assert.NotEmpty(t, output.APIKey)

	mockRepo.AssertExpectations(t)
}

func TestProjectService_GetByUserID(t *testing.T) {
	mockRepo := new(MockProjectRepoForTest)
	projectService := NewProjectService(mockRepo)

	projects := []*models.Project{
		{ID: 1, UserID: 1, Name: "Project 1", APIKey: "key1", CreatedAt: time.Now()},
		{ID: 2, UserID: 1, Name: "Project 2", APIKey: "key2", CreatedAt: time.Now()},
	}

	mockRepo.On("GetByUserID", 1).Return(projects, nil)

	result, err := projectService.GetByUserID(1)
	require.NoError(t, err)
	assert.Len(t, result, 2)

	mockRepo.AssertExpectations(t)
}

func TestProjectService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockProjectRepoForTest)
	projectService := NewProjectService(mockRepo)

	project := &models.Project{
		ID:     1,
		UserID: 1,
		Name:   "Test Project",
		APIKey: "key1",
	}

	mockRepo.On("GetByID", 1).Return(project, nil)

	result, err := projectService.GetByID(1, 1)
	require.NoError(t, err)
	assert.Equal(t, "Test Project", result.Name)

	mockRepo.AssertExpectations(t)
}

func TestProjectService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockProjectRepoForTest)
	projectService := NewProjectService(mockRepo)

	mockRepo.On("GetByID", 1).Return(nil, nil)

	_, err := projectService.GetByID(1, 1)
	assert.Error(t, err)
	assert.Equal(t, ErrProjectNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestProjectService_GetByID_Forbidden(t *testing.T) {
	mockRepo := new(MockProjectRepoForTest)
	projectService := NewProjectService(mockRepo)

	project := &models.Project{
		ID:     1,
		UserID: 2,
		Name:   "Test Project",
		APIKey: "key1",
	}

	mockRepo.On("GetByID", 1).Return(project, nil)

	_, err := projectService.GetByID(1, 1)
	assert.Error(t, err)
	assert.Equal(t, ErrProjectForbidden, err)

	mockRepo.AssertExpectations(t)
}

func TestProjectService_Delete_Success(t *testing.T) {
	mockRepo := new(MockProjectRepoForTest)
	projectService := NewProjectService(mockRepo)

	project := &models.Project{
		ID:     1,
		UserID: 1,
		Name:   "Test Project",
	}

	mockRepo.On("GetByID", 1).Return(project, nil)
	mockRepo.On("Delete", 1).Return(nil)

	err := projectService.Delete(1, 1)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}