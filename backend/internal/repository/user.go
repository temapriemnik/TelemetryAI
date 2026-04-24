package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"telemetryai/internal/models"
)

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at`

	return r.pool.QueryRow(context.Background(), query, user.Email, user.PasswordHash).
		Scan(&user.ID, &user.CreatedAt)
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`

	var user models.User
	err := r.pool.QueryRow(context.Background(), query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(id int) (*models.User, error) {
	query := `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`

	var user models.User
	err := r.pool.QueryRow(context.Background(), query, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}