package repository

import (
	"context"
	"errors"

	"identity-service/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGProjectRepository struct {
	pool *pgxpool.Pool
}

func NewPGProjectRepository(pool *pgxpool.Pool) *PGProjectRepository {
	return &PGProjectRepository{pool: pool}
}

func (r *PGProjectRepository) Create(ctx context.Context, project *entity.Project) error {
	query := `
		INSERT INTO projects (id, user_id, name, created_at)
		VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query,
		project.ID, project.UserID, project.Name, project.CreateAt)
	return err
}

func (r *PGProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	query := `SELECT id, user_id, name, created_at FROM projects WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var project entity.Project
	err := row.Scan(&project.ID, &project.UserID, &project.Name, &project.CreateAt)
	if errors.Is(err, context.Canceled) {
		return nil, err
	}
	return &project, err
}

func (r *PGProjectRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Project, error) {
	query := `SELECT id, user_id, name, created_at FROM projects WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*entity.Project
	for rows.Next() {
		var project entity.Project
		if err := rows.Scan(&project.ID, &project.UserID, &project.Name, &project.CreateAt); err != nil {
			return nil, err
		}
		projects = append(projects, &project)
	}
	return projects, rows.Err()
}

func (r *PGProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}