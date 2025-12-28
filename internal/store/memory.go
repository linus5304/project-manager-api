package store

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linus5304/project-manager-api/internal/domain"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrProjectNotFound = errors.New("project not found")
	ErrTaskNotFound    = errors.New("task not found")
)

type MemoryStore struct {
	mu       sync.RWMutex
	projects map[uuid.UUID]domain.Project
	tasks    map[uuid.UUID]map[uuid.UUID]domain.Task
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects: make(map[uuid.UUID]domain.Project),
		tasks:    make(map[uuid.UUID]map[uuid.UUID]domain.Task),
	}
}

func (s *MemoryStore) InsertProject(name string) (domain.Project, error) {
	p := domain.Project{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	s.projects[p.ID] = p
	s.mu.Unlock()

	return p, nil
}

func (s *MemoryStore) GetProject(id uuid.UUID) (domain.Project, error) {
	s.mu.RLock()
	p, ok := s.projects[id]
	s.mu.RUnlock()

	if !ok {
		return domain.Project{}, ErrNotFound
	}

	return p, nil
}

func (s *MemoryStore) ListProjects() ([]domain.Project, error) {
	s.mu.RLock()
	projects := make([]domain.Project, 0, len(s.projects))
	for _, p := range s.projects {
		projects = append(projects, p)
	}
	s.mu.RUnlock()

	sort.Slice(projects, func(i, j int) bool {
		if projects[i].CreatedAt.Equal(projects[j].CreatedAt) {
			return projects[i].ID.String() > projects[j].ID.String()
		}
		return projects[i].CreatedAt.After(projects[j].CreatedAt)
	})
	return projects, nil
}

func (s *MemoryStore) InsertTask(projectID uuid.UUID, title, description string) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure the project exists
	if _, ok := s.projects[projectID]; !ok {
		return domain.Task{}, ErrProjectNotFound
	}

	t := domain.Task{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      "todo",
		CreatedAt:   time.Now().UTC(),
	}

	if s.tasks[projectID] == nil {
		s.tasks[projectID] = make(map[uuid.UUID]domain.Task)
	}
	s.tasks[projectID][t.ID] = t
	return t, nil
}

func (s *MemoryStore) ListTasks(projectID uuid.UUID) ([]domain.Task, error) {
	s.mu.RLock()
	if _, ok := s.projects[projectID]; !ok {
		s.mu.RUnlock()
		return []domain.Task{}, ErrProjectNotFound
	}

	projectTasks, ok := s.tasks[projectID]

	if !ok || len(projectTasks) == 0 {
		s.mu.RUnlock()
		return []domain.Task{}, nil
	}

	tasks := make([]domain.Task, 0, len(projectTasks))
	for _, t := range projectTasks {
		tasks = append(tasks, t)
	}
	s.mu.RUnlock()

	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].CreatedAt.Equal(tasks[j].CreatedAt) {
			return tasks[i].ID.String() > tasks[j].ID.String()
		}
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	return tasks, nil
}

func (s *MemoryStore) UpdateTask(projectID, taskID uuid.UUID, update TaskUpdate) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.projects[projectID]; !ok {
		return domain.Task{}, ErrProjectNotFound
	}

	taskMap, ok := s.tasks[projectID]
	if !ok {
		return domain.Task{}, ErrTaskNotFound
	}

	task, ok := taskMap[taskID]
	if !ok {
		return domain.Task{}, ErrTaskNotFound
	}
	if update.Title != nil {
		task.Title = *update.Title
	}
	if update.Description != nil {
		task.Description = *update.Description
	}
	if update.Status != nil {
		task.Status = *update.Status
	}
	s.tasks[projectID][taskID] = task
	return task, nil
}
