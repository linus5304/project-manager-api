package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/linus5304/project-manager-api/internal/domain"
	"github.com/linus5304/project-manager-api/internal/store"
)

type createProjectInput struct {
	Name string `json:"name"`
}

type metadata struct {
	Page         int `json:"page"`
	PageSize     int `json:"pageSize"`
	TotalRecords int `json:"totalRecords"`
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

func (app *Application) listProjects(w http.ResponseWriter, r *http.Request) {
	page, err := readIntQuery(r, "page", 1)
	if err != nil {
		badRequestResponse(w, r, err)
		return
	}

	pageSize, err := readIntQuery(r, "page_size", 20)
	if err != nil {
		badRequestResponse(w, r, err)
		return
	}

	if err := validatePageParams(page, pageSize); err != nil {
		badRequestResponse(w, r, err)
		return
	}

	all, err := app.store.ListProjects()
	if err != nil {
		serverErrorResponse(w, r, err)
		return
	}

	total := len(all)
	offset := (page - 1) * pageSize

	// If offset is beyond the end, return an empty page (not an error)
	var pageItems []domain.Project
	if offset < total {
		end := offset + pageSize
		if end > total {
			end = total
		}
		pageItems = all[offset:end]
	} else {
		pageItems = []domain.Project{}
	}

	env := map[string]any{
		"projects": pageItems,
		"metadata": metadata{
			Page:         page,
			PageSize:     pageSize,
			TotalRecords: total,
		},
	}

	_ = writeJSON(w, http.StatusOK, env, nil)
}
