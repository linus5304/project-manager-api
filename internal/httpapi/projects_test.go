package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject_201(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	body := []byte(`{"name": "New Project"}`)
	res, err := http.Post(ts.URL+"/v1/projects", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /v1/projects failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 201 Created; got %d; body=%s", res.StatusCode, string(b))
	}

	var got map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	name, ok := got["name"].(string)
	if !ok || name != "New Project" {
		t.Fatalf("expected project name %#v; got %q", "New Project", got["name"])
	}

	id, ok := got["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty project ID, got %#v", got["id"])
	}
}

func TestCreateProject_400_WhenNameMissing(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	body := []byte(`{"name": ""}`)
	res, err := http.Post(ts.URL+"/v1/projects", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /v1/projects failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 400; got %d; body=%s", res.StatusCode, string(b))
	}
}
