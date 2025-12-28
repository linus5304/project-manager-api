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
