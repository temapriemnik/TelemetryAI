package service

import (
	"context"
	"errors"
	"time"

	"identity-service/internal/domain/entity"
	"identity-service/internal/domain/repository"

	"github.com/google/uuid"
)

type ProjectService struct {
	projectRepo repository.ProjectRepository
	apiKeyRepo repository.APIKeyRepository
}

func NewProjectService(projectRepo repository.ProjectRepository, apiKeyRepo repository.APIKeyRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		apiKeyRepo:  apiKeyRepo,
	}
}

func (s *ProjectService) Create(ctx context.Context, userID uuid.UUID, name string) (*entity.Project, error) {
	project := &entity.Project{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		CreateAt: time.Now(),
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrProjectNotFound
	}

	if project.UserID != userID {
		return nil, ErrProjectAccess
	}

	return project, nil
}

func (s *ProjectService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Project, error) {
	return s.projectRepo.GetByUserID(ctx, userID)
}

func (s *ProjectService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, id)
	if err != nil {
		return ErrProjectNotFound
	}

	if project.UserID != userID {
		return ErrProjectAccess
	}

	return s.projectRepo.Delete(ctx, id)
}

func (s *ProjectService) GetAPIKey(ctx context.Context, projectID, userID uuid.UUID) (string, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return "", ErrProjectNotFound
	}

	if project.UserID != userID {
		return "", ErrProjectAccess
	}

	keys, err := s.apiKeyRepo.GetByProjectID(ctx, projectID)
	if err != nil {
		return "", err
	}

	if len(keys) == 0 {
		return "", errors.New("api key not found")
	}

	return keys[0].KeyHash, nil
}