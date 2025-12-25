package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	return err
}

func errorResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	env := map[string]any{
		"error": map[string]string{
			"message": message,
		},
	}

	_ = writeJSON(w, status, env, nil)
}

func badRequestResponse(w http.ResponseWriter, r *http.Request, message string) {
	errorResponse(w, r, http.StatusBadRequest, message)
}

func serverErrorResponse(w http.ResponseWriter, r *http.Request) {
	errorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func notFoundResponse(w http.ResponseWriter, r *http.Request) {
	errorResponse(w, r, http.StatusNotFound, "the requested resource could not be found")
}

var errInvalidJson = errors.New("invalid JSON")
