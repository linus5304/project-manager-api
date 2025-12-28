package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createProject(t *testing.T, ts *httptest.Server, name string) string {
	t.Helper()

	body := []byte(`{"name": "` + name + `"}`)
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

	id, ok := got["id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty project ID, got %#v", got["id"])
	}

	return id
}

func TestCreateTask_201_DefaultTodo(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	projectID := createProject(t, ts, "Project for Task")

	body := []byte(`{"title": "Task 1", "description": "Task Description"}`)
	res, err := http.Post(ts.URL+"/v1/projects/"+projectID+"/tasks", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /v1/projects/%s/tasks failed: %v", projectID, err)
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

	if got["status"] != "todo" {
		t.Fatalf("expected default status 'todo'; got %q", got["status"])
	}
	if got["title"] != "Task 1" {
		t.Fatalf("expected title 'Task 1'; got %q", got["title"])
	}
	if got["description"] != "Task Description" {
		t.Fatalf("expected description 'Task Description'; got %q", got["description"])
	}
}

func TestCreateTask_404_ProjectMissing(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	nonExistentProjectID := "00000000-0000-0000-0000-000000000000"
	body := []byte(`{"title": "Task 1", "description": "Task Description"}`)
	res, err := http.Post(ts.URL+"/v1/projects/"+nonExistentProjectID+"/tasks", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /v1/projects/%s/tasks failed: %v", nonExistentProjectID, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 404 Not Found; got %d; body=%s", res.StatusCode, string(b))
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404 Not Found; got %d", res.StatusCode)
	}
}

func TestCreateTask_400_InvalidProjectID(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	body := []byte(`{"title": "Task 1", "description": "Task Description"}`)
	res, err := http.Post(ts.URL+"/v1/projects/invalid-uuid/tasks", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /v1/projects/invalid-uuid/tasks failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 400 Bad Request; got %d; body=%s", res.StatusCode, string(b))
	}
}

func TestListTasks_200_Empty(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	projectID := createProject(t, ts, "Project for Empty Task List")
	res, err := http.Get(ts.URL + "/v1/projects/" + projectID + "/tasks")
	if err != nil {
		t.Fatalf("GET /v1/projects/%s/tasks failed: %v", projectID, err)
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
	tasks, ok := env["tasks"].([]any)
	if !ok {
		t.Fatalf("expected tasks to be an array; got %#v", env["tasks"])
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks; got %d", len(tasks))
	}
}

func TestListTasks_404_ProjectMissing(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	nonExistentProjectID := "00000000-0000-0000-0000-000000000000"
	res, err := http.Get(ts.URL + "/v1/projects/" + nonExistentProjectID + "/tasks")
	if err != nil {
		t.Fatalf("GET /v1/projects/%s/tasks failed: %v", nonExistentProjectID, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 404 Not Found; got %d; body=%s", res.StatusCode, string(b))
	}
}

func TestCreateTask_400_BlankTitle(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	projectID := createProject(t, ts, "Project for Blank Title Task")

	body := []byte(`{"title": "   ", "description": "Task Description"}`)
	res, err := http.Post(ts.URL+"/v1/projects/"+projectID+"/tasks", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /v1/projects/%s/tasks failed: %v", projectID, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("expected status 400 Bad Request; got %d; body=%s", res.StatusCode, string(b))
	}
}

func TestListTasks_200_NewestFirst(t *testing.T) {
	app := NewApplication()
	ts := httptest.NewServer(app.Routes())
	t.Cleanup(ts.Close)

	projectID := createProject(t, ts, "Project for Task Listing")

	// Create multiple tasks
	createTasks := func(title string) {
		body := []byte(`{"title": "` + title + `", "description": "Description for ` + title + `"}`)
		res, err := http.Post(ts.URL+"/v1/projects/"+projectID+"/tasks", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("POST /v1/projects/%s/tasks failed: %v", projectID, err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			b, _ := io.ReadAll(res.Body)
			t.Fatalf("expected status 201 Created; got %d; body=%s", res.StatusCode, string(b))
		}
	}

	createTasks("first")
	createTasks("second")

	res, err := http.Get(ts.URL + "/v1/projects/" + projectID + "/tasks")
	if err != nil {
		t.Fatalf("GET /v1/projects/%s/tasks failed: %v", projectID, err)
	}
	defer res.Body.Close()

	var env map[string]any
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	tasks, ok := env["tasks"].([]any)
	if !ok || len(tasks) != 2 {
		t.Fatalf("expected 2 tasks; got %#v", env["tasks"])
	}

	first, _ := tasks[0].(map[string]any)
	if first["title"] != "second" {
		t.Fatalf("expected newest-first 'second'; got %q", first["title"])
	}
}
