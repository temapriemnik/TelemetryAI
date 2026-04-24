package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"telemetryai/internal/models"
)

type MockLogRepository struct {
	mock.Mock
}

func (m *MockLogRepository) Create(log *models.Log) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockLogRepository) GetByProjectID(projectID int) ([]*models.Log, error) {
	args := m.Called(projectID)
	return args.Get(0).([]*models.Log), args.Error(1)
}

type MProjectRepository struct {
	mock.Mock
}

func (m *MProjectRepository) Create(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MProjectRepository) GetByID(id int) (*models.Project, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MProjectRepository) GetByUserID(userID int) ([]*models.Project, error) {
	args := m.Called(userID)
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MProjectRepository) GetByAPIKey(apiKey string) (*models.Project, error) {
	args := m.Called(apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MProjectRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MProjectRepository) Update(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func TestLogService_Receive_InfoLog(t *testing.T) {
	mockLogRepo := new(MockLogRepository)
	mockProjectRepo := new(MProjectRepository)
	notificationService := NewMockNotificationService()
	logService := NewLogService(mockLogRepo, mockProjectRepo, notificationService, nil, nil)

	project := &models.Project{
		ID:     1,
		UserID: 1,
		Name:   "Test Project",
		APIKey: "test-api-key",
	}

	mockProjectRepo.On("GetByAPIKey", "test-api-key").Return(project, nil)
	mockLogRepo.On("Create", mock.AnythingOfType("*models.Log")).Return(nil).Run(func(args mock.Arguments) {
		log := args.Get(0).(*models.Log)
		log.ID = 1
		log.CreatedAt = time.Now()
	})

	input := ReceiveLogInput{
		APIKey:    "test-api-key",
		Timestamp: time.Now(),
		Message:   "This is an info message",
	}

	output, err := logService.Receive(input)
	require.NoError(t, err)
	assert.Equal(t, "INFO", output.Level)

	mockLogRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
}

func TestLogService_Receive_WarnLogWithPrefix(t *testing.T) {
	mockLogRepo := new(MockLogRepository)
	mockProjectRepo := new(MProjectRepository)
	notificationService := NewMockNotificationService()
	logService := NewLogService(mockLogRepo, mockProjectRepo, notificationService, nil, nil)

	project := &models.Project{
		ID:     1,
		UserID: 1,
		Name:   "Test Project",
		APIKey: "test-api-key",
	}

	mockProjectRepo.On("GetByAPIKey", "test-api-key").Return(project, nil)
	mockLogRepo.On("Create", mock.AnythingOfType("*models.Log")).Return(nil).Run(func(args mock.Arguments) {
		log := args.Get(0).(*models.Log)
		log.ID = 1
		log.CreatedAt = time.Now()
	})

	input := ReceiveLogInput{
		APIKey:    "test-api-key",
		Timestamp: time.Now(),
		Message:   "[WARN] High memory usage detected",
	}

	output, err := logService.Receive(input)
	require.NoError(t, err)
	assert.Equal(t, "WARN", output.Level)

	mockLogRepo.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
}