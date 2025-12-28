package store

import (
	"github.com/google/uuid"
	"github.com/linus5304/project-manager-api/internal/domain"
)

type TaskUpdate struct {
	Title       *string
	Description *string
	Status      *string
}

type ProjectStore interface {
	InsertProject(name string) (domain.Project, error)
	GetProject(id uuid.UUID) (domain.Project, error)
	ListProjects() ([]domain.Project, error)

	InsertTask(projectID uuid.UUID, title, description string) (domain.Task, error)
	ListTasks(projectID uuid.UUID) ([]domain.Task, error)
	UpdateTask(projectID, taskID uuid.UUID, update TaskUpdate) (domain.Task, error)
}
