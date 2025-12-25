package httpapi

import "net/http"

func (app *Application) healthz(w http.ResponseWriter, r *http.Request) {
	env := map[string]string{"status": "ok"}
	_ = writeJSON(w, http.StatusOK, env, nil)
}
