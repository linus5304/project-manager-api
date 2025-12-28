package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/linus5304/project-manager-api/internal/store"
)

type createTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (app *Application) createTask(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		badRequestResponse(w, r, errors.New("invalid project id"))
		return
	}

	var input createTaskInput
	if err := readJSON(w, r, &input); err != nil {
		badRequestResponse(w, r, err)
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		badRequestResponse(w, r, errors.New("title is required"))
		return
	}

	t, err := app.store.InsertTask(projectID, input.Title, input.Description)
	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			notFoundResponse(w, r)
			return
		}
		serverErrorResponse(w, r, err)
		return
	}

	_ = writeJSON(w, http.StatusCreated, t, nil)
}

func (app *Application) listTasks(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		badRequestResponse(w, r, errors.New("invalid project id"))
		return
	}
	tasks, err := app.store.ListTasks(projectID)
	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			notFoundResponse(w, r)
			return
		}
		serverErrorResponse(w, r, err)
		return
	}

	env := map[string]any{
		"tasks": tasks,
		"metadata": metadata{
			Page:         1,
			PageSize:     len(tasks),
			TotalRecords: len(tasks),
		},
	}

	_ = writeJSON(w, http.StatusOK, env, nil)
}

type updateTaskInput struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

func (app *Application) updateTask(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.PathValue("projectId")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		badRequestResponse(w, r, errors.New("invalid project id"))
		return
	}
	taskIDStr := r.PathValue("taskId")
	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		badRequestResponse(w, r, errors.New("invalid task id"))
		return
	}

	var input updateTaskInput
	if err := readJSON(w, r, &input); err != nil {
		badRequestResponse(w, r, err)
		return
	}

	// Must provide at least one field for PATCH
	if input.Title == nil && input.Description == nil && input.Status == nil {
		badRequestResponse(w, r, errors.New("body must contain at least one of title, description or status"))
		return
	}

	// Normalize + validate
	if input.Title != nil {
		t := strings.TrimSpace(*input.Title)
		if t == "" {
			badRequestResponse(w, r, errors.New("title cannot be empty"))
			return
		}
		input.Title = &t
	}

	if input.Description != nil {
		d := strings.TrimSpace(*input.Description)
		input.Description = &d
	}

	if input.Status != nil {
		s := strings.TrimSpace(*input.Status)
		switch s {
		case "todo", "doing", "done":
			// ok
		default:
			badRequestResponse(w, r, errors.New("status must be one of: todo, doing, done"))
			return
		}
		input.Status = &s
	}

	update := store.TaskUpdate{
		Title:       input.Title,
		Description: input.Description,
		Status:      input.Status,
	}

	updated, err := app.store.UpdateTask(projectID, taskID, update)
	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) || errors.Is(err, store.ErrTaskNotFound) {
			notFoundResponse(w, r)
			return
		}
		serverErrorResponse(w, r, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, updated, nil)

}
