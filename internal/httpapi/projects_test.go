package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateProject_201(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	body := []byte(`{"name": "New Project"}`)
	res, err := http.Post(ts.URL+"/projects", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /projects failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created; got %d", res.StatusCode)
	}

	var got map[string]any
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if got["name"] != "New Project" {
		t.Fatalf("expected project name %q; got %q", "New Project", got["name"])
	}
	if got["id"] == "" {
		t.Fatal("expected non-empty project ID")
	}
}

func TestCreateProject_400_WhenNameMissing(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	body := []byte(`{"name": ""}`)
	res, err := http.Post(ts.URL+"/projects", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400; got %d", res.StatusCode)
	}
}
