package store

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/linus5304/project-manager-api/internal/domain"
)

var ErrNotFound = errors.New("not found")

type MemoryStore struct {
	mu       sync.RWMutex
	projects map[uuid.UUID]domain.Project
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects: make(map[uuid.UUID]domain.Project),
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
