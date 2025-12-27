package httpapi

import "net/http"

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", app.healthz)
	mux.HandleFunc("POST /v1/projects", app.createProject)
	mux.HandleFunc("GET /v1/projects/{id}", app.getProject)
	mux.HandleFunc("GET /v1/projects", app.listProjects)

	h := http.Handler(mux)
	h = app.logRequestMiddleware(h)
	h = app.recoverPanicMiddleware(h)
	h = app.requestIDMiddleware(h)
	return h
}
