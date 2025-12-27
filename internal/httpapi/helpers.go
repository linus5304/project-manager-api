package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

func readIntQuery(r *http.Request, key string, defaultValue int) (int, error) {
	qs := r.URL.Query()
	s := qs.Get(key)
	if s == "" {
		return defaultValue, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("query parameter %q must be an integer", key)
	}

	return i, nil
}

func validatePageParams(page, pageSize int) error {
	if page < 1 {
		return errors.New("page must be >= 1")
	}

	if pageSize < 1 {
		return errors.New("page_size must be >= 1")
	}

	if pageSize > 100 {
		return errors.New("page_size must be <= 100")
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

func badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func notFoundResponse(w http.ResponseWriter, r *http.Request) {
	errorResponse(w, r, http.StatusNotFound, "the requested resource could not be found")
}

var errInvalidJson = errors.New("invalid JSON")
