package httpapi

import "net/http"

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", app.healthz)

	return mux
}
