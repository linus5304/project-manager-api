package httpapi

import "github.com/linus5304/project-manager-api/internal/store"

type Application struct {
	store store.ProjectStore
}

func NewApplication() *Application {
	return &Application{
		store: store.NewMemoryStore(),
	}
}
