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
	app := newTestApp()
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
	app := newTestApp()
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

func TestGetProject_200(t *testing.T) {
	app := newTestApp()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	// First, create a project to retrieve later
	createBody := []byte(`{"name": "Alpha"}`)
	res, err := http.Post(ts.URL+"/v1/projects", "application/json", bytes.NewReader(createBody))
	if err != nil {
		t.Fatalf("POST /v1/projects failed: %v", err)
	}
	defer res.Body.Close()

	var created map[string]any
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response body: %v", err)
	}
	id, ok := created["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty project ID, got %#v", created["id"])
	}

	// Now, retrieve the created project
	getRes, err := http.Get(ts.URL + "/v1/projects/" + id)
	if err != nil {
		t.Fatalf("GET /v1/projects/%s failed: %v", id, err)
	}
	defer getRes.Body.Close()

	if getRes.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(getRes.Body)
		t.Fatalf("expected status 200 OK; got %d; body=%s", getRes.StatusCode, string(b))
	}

	var got map[string]any
	if err := json.NewDecoder(getRes.Body).Decode(&got); err != nil {
		t.Fatalf("decode get response body: %v", err)
	}

	if got["id"] != id {
		t.Fatalf("expected project ID %#v; got %#v", id, got["id"])
	}
}

func TestGetProject_400_InvalidID(t *testing.T) {
	app := newTestApp()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/v1/projects/invalid-uuid")
	if err != nil {
		t.Fatalf("GET /v1/projects/invalid-uuid failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 400 Bad Request; got %d; body=%s", res.StatusCode, string(b))
	}
}

func TestGetProject_404_NotFound(t *testing.T) {
	app := newTestApp()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	nonExistentID := "123e4567-e89b-12d3-a456-426614174000"
	res, err := http.Get(ts.URL + "/v1/projects/" + nonExistentID)
	if err != nil {
		t.Fatalf("GET /v1/projects/%s failed: %v", nonExistentID, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 404 Not Found; got %d; body=%s", res.StatusCode, string(b))
	}
}

func TestListProjects_200_Empty(t *testing.T) {
	app := newTestApp()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/v1/projects")
	if err != nil {
		t.Fatalf("GET /v1/projects failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200 OK; got %d; body=%s", res.StatusCode, string(b))
	}

	var env map[string]any
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	projects, ok := env["projects"].([]any)
	if !ok {
		t.Fatalf("expected projects array, got %#v", env["projects"])
	}
	if len(projects) != 0 {
		t.Fatalf("expected 0 projects, got %d", len(projects))
	}

	md, ok := env["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("expected metadata object, got %#v", env["metadata"])
	}
	if md["totalRecords"] != float64(0) {
		t.Fatalf("expected totalRecords 0, got %v", md["totalRecords"])
	}
}

func TestListProjects_200_NewestFirst(t *testing.T) {
	app := newTestApp()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	create := func(name string) {
		body := []byte(`{"name": "` + name + `"}`)
		res, err := http.Post(ts.URL+"/v1/projects", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("create %s failed: %v", name, err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(res.Body)
			t.Fatalf("expected status 201 Created; got %d; body=%s", res.StatusCode, string(b))
		}
	}

	create("Alpha")
	create("Beta")
	create("Gamma")

	res, err := http.Get(ts.URL + "/v1/projects")
	if err != nil {
		t.Fatalf("GET /v1/projects failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 200 OK; got %d; body=%s", res.StatusCode, string(b))
	}

	var env map[string]any
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	raw, ok := env["projects"].([]any)
	if !ok {
		t.Fatalf("expected projects array, got %#v", env["projects"])
	}
	if len(raw) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(raw))
	}

	// Projects[0] should be the newest (Project C)
	first, ok := raw[0].(map[string]any)
	if !ok {
		t.Fatalf("expected project object, got %#v", raw[0])
	}
	name, _ := first["name"].(string)
	if name != "Gamma" {
		t.Fatalf("expected first project to be 'Gamma', got %q", name)
	}
}

func TestListProjects_400_InvalidPage(t *testing.T) {
	app := newTestApp()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	res, err := http.Get(ts.URL + "/v1/projects?page=0")
	if err != nil {
		t.Fatalf("GET /v1/projects failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 400 Bad Request; got %d; body=%s", res.StatusCode, string(b))
	}
}
