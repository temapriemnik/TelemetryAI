package service

import (
	"context"
	"testing"

	"identity-service/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByAuthToken(ctx context.Context, token string) (*entity.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, assert.AnError)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

	token, user, err := authService.Register(context.Background(), "test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register_UserExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	existingUser := &entity.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}
	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	_, _, err := authService.Register(context.Background(), "test@example.com", "password123")

	assert.Equal(t, ErrUserExists, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, assert.AnError)

	_, _, err := authService.Login(context.Background(), "test@example.com", "password123")

	assert.Equal(t, ErrInvalidCreds, err)
	mockRepo.AssertExpectations(t)
}