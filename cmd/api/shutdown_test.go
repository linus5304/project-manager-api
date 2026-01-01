package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestShutdownServer_TimeoutForcesClose(t *testing.T) {
	started := make(chan struct{})
	block := make(chan struct{})

	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-block:
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// If the server closes the connection, request ctx is canceled.
			return
		}
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })

	client := &http.Client{Timeout: 2 * time.Second}

	// Fire request in background
	done := make(chan error, 1)
	go func() {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+ln.Addr().String()+"/slow", nil)
		_, err := client.Do(req)
		done <- err
	}()

	<-started // ensure request is running

	// Now shut down with a tiny timeout: should time out and escalate to Close().
	err = shutdownServer(srv, 50*time.Millisecond)
	if err == nil {
		t.Fatalf("expected shutdown error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded; got %v", err)
	}

	// Unblock handler so we don't leak goroutines in the test.
	close(block)

	// Client should fail because we escalated to Close().
	if clientErr := <-done; clientErr == nil {
		t.Fatalf("expected client error after forced close, got nil")
	}
}
