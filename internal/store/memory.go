package store

import (
	"errors"
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
