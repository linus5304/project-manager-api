package httpapi

import "net/http"

func (app *Application) livez(w http.ResponseWriter, r *http.Request) {
	_ = writeJSON(w, http.StatusOK, map[string]string{"status": "OK"}, nil)
}
