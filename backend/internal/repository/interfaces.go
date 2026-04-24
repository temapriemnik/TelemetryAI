package repository

import "telemetryai/internal/models"

type UserRepository interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByID(id int) (*models.User, error)
}

type ProjectRepository interface {
	Create(project *models.Project) error
	GetByID(id int) (*models.Project, error)
	GetByUserID(userID int) ([]*models.Project, error)
	GetByAPIKey(apiKey string) (*models.Project, error)
	Delete(id int) error
	Update(project *models.Project) error
}

type LogRepository interface {
	Create(log *models.Log) error
	GetByProjectID(projectID int) ([]*models.Log, error)
}