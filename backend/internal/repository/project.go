package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"telemetryai/internal/models"
)

type projectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) ProjectRepository {
	return &projectRepository{pool: pool}
}

func (r *projectRepository) Create(project *models.Project) error {
	query := `
		INSERT INTO projects (user_id, name, api_key, description, architecture)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return r.pool.QueryRow(context.Background(), query,
		project.UserID, project.Name, project.APIKey, project.Description, project.Architecture).
		Scan(&project.ID, &project.CreatedAt)
}

func (r *projectRepository) GetByID(id int) (*models.Project, error) {
	query := `SELECT id, user_id, name, api_key, COALESCE(description, ''), COALESCE(architecture, ''), created_at 
		FROM projects WHERE id = $1`

	var project models.Project
	err := r.pool.QueryRow(context.Background(), query, id).
		Scan(&project.ID, &project.UserID, &project.Name, &project.APIKey, &project.Description, &project.Architecture, &project.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *projectRepository) GetByUserID(userID int) ([]*models.Project, error) {
	query := `SELECT id, user_id, name, api_key, COALESCE(description, ''), COALESCE(architecture, ''), created_at 
		FROM projects WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		var project models.Project
		if err := rows.Scan(&project.ID, &project.UserID, &project.Name, &project.APIKey, &project.Description, &project.Architecture, &project.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

func (r *projectRepository) GetByAPIKey(apiKey string) (*models.Project, error) {
	query := `SELECT id, user_id, name, api_key, COALESCE(description, ''), COALESCE(architecture, ''), created_at 
		FROM projects WHERE api_key = $1`

	var project models.Project
	err := r.pool.QueryRow(context.Background(), query, apiKey).
		Scan(&project.ID, &project.UserID, &project.Name, &project.APIKey, &project.Description, &project.Architecture, &project.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *projectRepository) Delete(id int) error {
	query := `DELETE FROM projects WHERE id = $1`
	_, err := r.pool.Exec(context.Background(), query, id)
	return err
}

func (r *projectRepository) Update(project *models.Project) error {
	query := `UPDATE projects SET name = $1, description = $2, architecture = $3 WHERE id = $4`
	_, err := r.pool.Exec(context.Background(), query, project.Name, project.Description, project.Architecture, project.ID)
	return err
}