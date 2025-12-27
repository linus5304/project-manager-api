package httpapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ctxKey int

const requestIDKey ctxKey = iota

func getRequestID(r *http.Request) string {
	v, _ := r.Context().Value(requestIDKey).(string)
	return v
}

func (app *Application) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Read incoming request id if provided
		rid := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if rid == "" {
			// 2. Otherwise generate one
			rid = uuid.NewString()
		}

		// 3. Put it on the response so Clients can correlate
		w.Header().Set("X-Request-ID", rid)

		// 4. Store it in context so downstream handlers/middleware can use it
		ctx := context.WithValue(r.Context(), requestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) recoverPanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				rid := getRequestID(r)
				log.Printf("PANIC: request_id=%s, panic=%v", rid, rec)

				// if you want to be extra safe, you can also ensure the connection closes:
				w.Header().Set("Connection", "close")

				// Standard 500 response (do not leak internals)
				serverErrorResponse(w, r, fmt.Errorf("panic recovered: %v", rec))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(status int) {
	sr.status = status
	sr.ResponseWriter.WriteHeader(status)
}

func (app *Application) logRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sr, r)

		rid := getRequestID(r)
		dur := time.Since(start)

		log.Printf("INFO: request_id=%s, method=%s, path=%s, status=%d, duration=%s", rid, r.Method, r.URL.Path, sr.status, dur)
	})
}
