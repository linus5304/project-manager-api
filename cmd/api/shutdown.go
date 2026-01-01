package main

import (
	"context"
	"net/http"
	"time"
)

func shutdownServer(srv *http.Server, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		// Escalate: guarantee we stop accepting/serving if graceful shutdown timed out.
		_ = srv.Close()
		return err
	}
	return nil
}
