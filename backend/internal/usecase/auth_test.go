package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"telemetryai/internal/models"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo, "test-secret")

	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

	input := RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	}

	err := authService.Register(input)
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_ValidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo, "test-secret")

	user := &models.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: "$2a$10$tNMifWYSyCRWYzwC7rz4..wZ02peap1ITES6e4ZDVHqDUFwv8KuK.",
	}

	mockRepo.On("GetByEmail", "test@example.com").Return(user, nil)

	input := LoginInput{
		Email:     "test@example.com",
		Password: "password123",
	}

	output, err := authService.Login(input)
	require.NoError(t, err)
	assert.NotEmpty(t, output.Token)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo, "test-secret")

	mockRepo.On("GetByEmail", "test@example.com").Return(nil, nil)

	input := LoginInput{
		Email:     "test@example.com",
		Password: "wrong_password",
	}

	_, err := authService.Login(input)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo, "test-secret")

	mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(0).(*models.User)
		user.ID = 1
	})

	input := RegisterInput{
		Email:    "test2@example.com",
		Password: "password123",
	}

	err := authService.Register(input)
	require.NoError(t, err)

	loginInput := LoginInput{
		Email:     "test2@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByEmail", "test2@example.com").Return(&models.User{
		ID:           1,
		Email:        "test2@example.com",
		PasswordHash: "$2a$10$tNMifWYSyCRWYzwC7rz4..wZ02peap1ITES6e4ZDVHqDUFwv8KuK.",
	}, nil)

	output, err := authService.Login(loginInput)
	require.NoError(t, err)

	userID, err := authService.ValidateToken(output.Token)
	require.NoError(t, err)
	assert.Greater(t, userID, 0)
}

func TestAuthService_ValidateToken_Invalid(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo, "test-secret")

	_, err := authService.ValidateToken("invalid.token.here")
	assert.Error(t, err)
}