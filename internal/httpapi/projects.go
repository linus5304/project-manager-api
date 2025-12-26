package httpapi

import (
	"errors"
	"net/http"
	"strings"
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
