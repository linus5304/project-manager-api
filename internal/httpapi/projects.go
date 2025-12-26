package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/linus5304/project-manager-api/internal/store"
)

type createProjectInput struct {
	Name string `json:"name"`
}

func (app *Application) createProject(w http.ResponseWriter, r *http.Request) {
	var input createProjectInput

	if err := readJSON(w, r, &input); err != nil {
		badRequestResponse(w, r, err)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		badRequestResponse(w, r, errors.New("name is required"))
		return
	}

	p, err := app.store.InsertProject(input.Name)
	if err != nil {
		serverErrorResponse(w, r, err)
		return
	}

	_ = writeJSON(w, http.StatusCreated, p, nil)
}

func (app *Application) getProject(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		badRequestResponse(w, r, errors.New("invalid project ID"))
		return
	}

	p, err := app.store.GetProject(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			notFoundResponse(w, r)
			return
		}
		serverErrorResponse(w, r, err)
		return
	}

	_ = writeJSON(w, http.StatusOK, p, nil)
}
