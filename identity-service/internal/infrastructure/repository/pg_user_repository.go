package repository

import (
	"context"
	"errors"

	"identity-service/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGUserRepository struct {
	pool *pgxpool.Pool
}

func NewPGUserRepository(pool *pgxpool.Pool) *PGUserRepository {
	return &PGUserRepository{pool: pool}
}

func (r *PGUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, auth_token, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query,
		user.ID, user.Email, user.Password, user.AuthToken, user.CreatedAt)
	return err
}

func (r *PGUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `SELECT id, email, password_hash, auth_token, created_at FROM users WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var user entity.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.AuthToken, &user.CreatedAt)
	if errors.Is(err, context.Canceled) {
		return nil, err
	}
	return &user, err
}

func (r *PGUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, email, password_hash, auth_token, created_at FROM users WHERE email = $1`
	row := r.pool.QueryRow(ctx, query, email)

	var user entity.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.AuthToken, &user.CreatedAt)
	if errors.Is(err, context.Canceled) {
		return nil, err
	}
	return &user, err
}

func (r *PGUserRepository) GetByAuthToken(ctx context.Context, token string) (*entity.User, error) {
	query := `SELECT id, email, password_hash, auth_token, created_at FROM users WHERE auth_token = $1`
	row := r.pool.QueryRow(ctx, query, token)

	var user entity.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.AuthToken, &user.CreatedAt)
	if errors.Is(err, context.Canceled) {
		return nil, err
	}
	return &user, err
}