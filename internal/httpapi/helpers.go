package httpapi

import (
	"encoding/json"
	"errors"
	"io"
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

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576 // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		if err == io.EOF {
			return errors.New("body must not be empty")
		}
		return err
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
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
