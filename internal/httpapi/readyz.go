package httpapi

import (
	"context"
	"net/http"
	"time"
)

type pinger interface {
	Ping(ctx context.Context) error
}

func (app *Application) readyz(w http.ResponseWriter, r *http.Request) {
	p, ok := app.store.(pinger)
	if !ok {
		// Memorystate (or any store without Ping) => ready
		_ = writeJSON(w, http.StatusOK, map[string]string{"status": "OK"}, nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 250*time.Millisecond)
	defer cancel()

	if err := p.Ping(ctx); err != nil {
		_ = writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"}, nil)
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]string{"status": "OK"}, nil)
}
