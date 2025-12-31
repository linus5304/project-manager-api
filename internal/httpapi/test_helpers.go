package httpapi

import "github.com/linus5304/project-manager-api/internal/store"

func newTestApp() *Application {
	return NewApplication(store.NewMemoryStore())
}
