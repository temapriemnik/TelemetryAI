package repository

import (
	"context"
	"errors"

	"identity-service/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGAPIKeyRepository struct {
	pool *pgxpool.Pool
}

func NewPGAPIKeyRepository(pool *pgxpool.Pool) *PGAPIKeyRepository {
	return &PGAPIKeyRepository{pool: pool}
}

func (r *PGAPIKeyRepository) Create(ctx context.Context, key *entity.APIKey) error {
	query := `
		INSERT INTO api_keys (id, project_id, key_hash, level, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.pool.Exec(ctx, query,
		key.ID, key.ProjectID, key.KeyHash, key.Level, key.CreateAt)
	return err
}

func (r *PGAPIKeyRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*entity.APIKey, error) {
	query := `SELECT id, project_id, key_hash, level, created_at FROM api_keys WHERE project_id = $1`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*entity.APIKey
	for rows.Next() {
		var key entity.APIKey
		if err := rows.Scan(&key.ID, &key.ProjectID, &key.KeyHash, &key.Level, &key.CreateAt); err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}
	return keys, rows.Err()
}

func (r *PGAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	query := `SELECT id, project_id, key_hash, level, created_at FROM api_keys WHERE key_hash = $1`
	row := r.pool.QueryRow(ctx, query, keyHash)

	var key entity.APIKey
	err := row.Scan(&key.ID, &key.ProjectID, &key.KeyHash, &key.Level, &key.CreateAt)
	if errors.Is(err, context.Canceled) {
		return nil, err
	}
	return &key, err
}

func (r *PGAPIKeyRepository) Delete(ctx context.Context, keyHash string) error {
	query := `DELETE FROM api_keys WHERE key_hash = $1`
	_, err := r.pool.Exec(ctx, query, keyHash)
	return err
}

func (r *PGAPIKeyRepository) List(ctx context.Context) ([]*entity.APIKey, error) {
	query := `SELECT id, project_id, key_hash, level, created_at FROM api_keys`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*entity.APIKey
	for rows.Next() {
		var key entity.APIKey
		if err := rows.Scan(&key.ID, &key.ProjectID, &key.KeyHash, &key.Level, &key.CreateAt); err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}
	return keys, rows.Err()
}