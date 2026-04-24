package usecase

import (
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"telemetryai/internal/models"
	"telemetryai/internal/repository"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrProjectForbidden = errors.New("access denied to project")
)

type ProjectService struct {
	projectRepo repository.ProjectRepository
	logger    *slog.Logger
}

func NewProjectService(projectRepo repository.ProjectRepository) *ProjectService {
	return &ProjectService{projectRepo: projectRepo}
}

type CreateProjectInput struct {
	UserID         int    `json:"user_id"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Architecture   string `json:"architecture,omitempty"`
}

type ProjectOutput struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Name          string    `json:"name"`
	APIKey        string    `json:"api_key"`
	Description  string    `json:"description,omitempty"`
	Architecture string    `json:"architecture,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type UpdateProjectInput struct {
	Name          string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	Architecture string `json:"architecture,omitempty"`
}

type AlertDataOutput struct {
	UserEmail    string `json:"user_email"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *ProjectService) Create(input CreateProjectInput) (ProjectOutput, error) {
	slog.Info("creating project", "user_id", input.UserID, "name", input.Name)
	
	project := &models.Project{
		UserID:        input.UserID,
		Name:         input.Name,
		APIKey:       uuid.New().String(),
		Architecture: input.Architecture,
		Description:  input.Description,
	}

	if err := s.projectRepo.Create(project); err != nil {
		slog.Error("failed to create project", "error", err, "name", input.Name)
		return ProjectOutput{}, err
	}

	slog.Info("project created", "project_id", project.ID, "api_key", project.APIKey)
	return ProjectOutput{
		ID:            project.ID,
		UserID:        project.UserID,
		Name:         project.Name,
		APIKey:       project.APIKey,
		Architecture: project.Architecture,
		Description:  project.Description,
		CreatedAt:    project.CreatedAt,
	}, nil
}

func (s *ProjectService) GetByUserID(userID int) ([]ProjectOutput, error) {
	slog.Debug("getting projects for user", "user_id", userID)
	
	projects, err := s.projectRepo.GetByUserID(userID)
	if err != nil {
		slog.Error("failed to get projects", "error", err)
		return nil, err
	}

	result := make([]ProjectOutput, len(projects))
	for i, p := range projects {
		result[i] = ProjectOutput{
			ID:            p.ID,
			UserID:        p.UserID,
			Name:         p.Name,
			APIKey:       p.APIKey,
			Architecture: p.Architecture,
			Description:  p.Description,
			CreatedAt:    p.CreatedAt,
		}
	}

	return result, nil
}

func (s *ProjectService) GetByID(id, userID int) (ProjectOutput, error) {
	slog.Debug("getting project", "project_id", id, "user_id", userID)
	
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		slog.Error("failed to get project", "error", err, "project_id", id)
		return ProjectOutput{}, err
	}
	if project == nil {
		slog.Warn("project not found", "project_id", id)
		return ProjectOutput{}, ErrProjectNotFound
	}

	if project.UserID != userID {
		slog.Warn("access denied to project", "project_id", id, "user_id", userID)
		return ProjectOutput{}, ErrProjectForbidden
	}

	return ProjectOutput{
		ID:            project.ID,
		UserID:        project.UserID,
		Name:         project.Name,
		APIKey:       project.APIKey,
		Architecture: project.Architecture,
		Description:  project.Description,
		CreatedAt:    project.CreatedAt,
	}, nil
}

func (s *ProjectService) Delete(id, userID int) error {
	slog.Info("deleting project", "project_id", id, "user_id", userID)
	
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		slog.Error("failed to get project for delete", "error", err)
		return err
	}
	if project == nil {
		slog.Warn("project not found for delete", "project_id", id)
		return ErrProjectNotFound
	}

	if project.UserID != userID {
		slog.Warn("access denied to delete project", "project_id", id, "user_id", userID)
		return ErrProjectForbidden
	}

	if err := s.projectRepo.Delete(id); err != nil {
		slog.Error("failed to delete project", "error", err, "project_id", id)
		return err
	}

	slog.Info("project deleted", "project_id", id)
	return nil
}

func (s *ProjectService) Update(id, userID int, input UpdateProjectInput) (ProjectOutput, error) {
	slog.Info("updating project", "project_id", id, "user_id", userID)
	
	project, err := s.projectRepo.GetByID(id)
	if err != nil {
		slog.Error("failed to get project for update", "error", err, "project_id", id)
		return ProjectOutput{}, err
	}
	if project == nil {
		slog.Warn("project not found for update", "project_id", id)
		return ProjectOutput{}, ErrProjectNotFound
	}
	if project.UserID != userID {
		slog.Warn("access denied to update project", "project_id", id, "user_id", userID)
		return ProjectOutput{}, ErrProjectForbidden
	}

	if input.Name != "" {
		project.Name = input.Name
	}
	if input.Description != "" {
		project.Description = input.Description
	}
	if input.Architecture != "" {
		project.Architecture = input.Architecture
	}

	if err := s.projectRepo.Update(project); err != nil {
		slog.Error("failed to update project", "error", err, "project_id", id)
		return ProjectOutput{}, err
	}

	slog.Info("project updated", "project_id", id)
	return ProjectOutput{
		ID:            project.ID,
		UserID:        project.UserID,
		Name:         project.Name,
		APIKey:       project.APIKey,
		Architecture: project.Architecture,
		Description:  project.Description,
		CreatedAt:    project.CreatedAt,
	}, nil
}

func (s *ProjectService) GetAlertData(projectID int) (AlertDataOutput, error) {
	project, userEmail, err := s.projectRepo.GetByIDWithUserEmail(projectID)
	if err != nil {
		slog.Error("failed to get project with user email", "error", err, "project_id", projectID)
		return AlertDataOutput{}, err
	}
	if project == nil {
		slog.Warn("project not found for alert", "project_id", projectID)
		return AlertDataOutput{}, ErrProjectNotFound
	}

	return AlertDataOutput{
		UserEmail:   userEmail,
		Name:        project.Name,
		Description: project.Description,
	}, nil
}