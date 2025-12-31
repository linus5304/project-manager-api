package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/linus5304/project-manager-api/internal/domain"
)

type TaskUpdate struct {
	Title       *string
	Description *string
	Status      *string
}

type ProjectStore interface {
	InsertProject(ctx context.Context, name string) (domain.Project, error)
	GetProject(ctx context.Context, id uuid.UUID) (domain.Project, error)
	ListProjects(ctx context.Context) ([]domain.Project, error)

	InsertTask(ctx context.Context, projectID uuid.UUID, title, description string) (domain.Task, error)
	ListTasks(ctx context.Context, projectID uuid.UUID) ([]domain.Task, error)
	UpdateTask(ctx context.Context, projectID, taskID uuid.UUID, update TaskUpdate) (domain.Task, error)
}
