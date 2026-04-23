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
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists     = errors.New("user already exists")
	ErrInvalidCreds   = errors.New("invalid credentials")
	ErrUserNotFound   = errors.New("user not found")
	ErrProjectAccess  = errors.New("access denied")
	ErrProjectNotFound = errors.New("project not found")
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, *entity.User, error) {
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return "", nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	token := generateToken()
	user := &entity.User{
		ID:         uuid.New(),
		Email:      email,
		Password:   string(hash),
		AuthToken:  token,
		CreatedAt:  time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, *entity.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCreds
	}

	return user.AuthToken, user, nil
}

func (s *AuthService) GetUserByToken(ctx context.Context, token string) (*entity.User, error) {
	user, err := s.userRepo.GetByAuthToken(ctx, token)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}