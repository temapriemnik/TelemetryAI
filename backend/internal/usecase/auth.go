package usecase

import (
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"telemetryai/internal/models"
	"telemetryai/internal/repository"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound    = errors.New("user not found")
)

type AuthService struct {
	userRepo repository.UserRepository
	secret  string
	logger  *slog.Logger
}

func NewAuthService(userRepo repository.UserRepository, secret string) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		secret:  secret,
	}
}

func (s *AuthService) Register(input RegisterInput) error {
	slog.Info("registering user", "email", input.Email)
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		return err
	}

	user := &models.User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		slog.Error("failed to create user", "error", err)
		return err
	}

	slog.Info("user registered successfully", "email", input.Email, "user_id", user.ID)
	return nil
}

func (s *AuthService) Login(input LoginInput) (TokenOutput, error) {
	slog.Info("login attempt", "email", input.Email)
	
	user, err := s.userRepo.GetByEmail(input.Email)
	if err != nil {
		slog.Error("failed to get user", "error", err)
		return TokenOutput{}, err
	}
	if user == nil {
		slog.Warn("user not found", "email", input.Email)
		return TokenOutput{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		slog.Warn("invalid credentials", "email", input.Email)
		return TokenOutput{}, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		return TokenOutput{}, err
	}

	slog.Info("login successful", "email", input.Email, "user_id", user.ID)
	return TokenOutput{Token: token}, nil
}

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email     string `json:"email"`
	Password string `json:"password"`
}

type TokenOutput struct {
	Token string `json:"token"`
}

func (s *AuthService) generateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *AuthService) ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.secret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return int(userID), nil
}