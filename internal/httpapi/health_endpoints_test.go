package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/linus5304/project-manager-api/internal/store"
)

type pingStore struct {
	store.ProjectStore
	err error
}

func (ps pingStore) Ping(ctx context.Context) error { return ps.err }

func TestLivez_200(t *testing.T) {
	app := NewApplication(store.NewMemoryStore())
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/livez")
	if err != nil {
		t.Fatalf("GET /livez: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestReadyz_200_WithMemoryStore(t *testing.T) {
	app := NewApplication(store.NewMemoryStore())
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestReadyz_503_WhenPingFails(t *testing.T) {
	base := store.NewMemoryStore()
	app := NewApplication(pingStore{ProjectStore: base, err: errors.New("db down")})
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", res.StatusCode)
	}
}
