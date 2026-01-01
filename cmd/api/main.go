package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/linus5304/project-manager-api/internal/httpapi"
	"github.com/linus5304/project-manager-api/internal/store"
)

type closer interface{ Close() }

func main() {
	addr := ":4000"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	// Shutdown timeout (default 10s)
	shutdownTimeout := 10 * time.Second
	if v := os.Getenv("SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			shutdownTimeout = d
		} else {
			log.Fatalf("invalid SHUTDOWN_TIMEOUT: %v", err)
		}
	}

	// Store selection
	var st store.ProjectStore
	var stCloser closer

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Printf("INFO: DATABASE_URL not set; using MemoryStore")
		st = store.NewMemoryStore()
	} else {
		startCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pg, err := store.NewPostgresStore(startCtx, dsn)
		if err != nil {
			log.Fatalf("unable to connect to database: %v", err)
		}
		st = pg
		stCloser = pg // close later, after shutdown
	}

	app := httpapi.NewApplication(st)

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	errCh := make(chan error, 1)
	go func() {
		log.Printf("starting server on %s", addr)
		errCh <- srv.ListenAndServe()
	}()

	// Wait for signal OR server error
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-sigCtx.Done():
		log.Printf("INFO: shutdown signal received")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
		// If ErrServerClosed, it means shutdown happened elsewhere; continue to cleanup.
	}

	// Graceful shutdown with deadline
	if err := shutdownServer(srv, shutdownTimeout); err != nil {
		log.Printf("ERROR: shutdown: %v", err)
	}

	err := <-errCh
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("ERROR: server returned: %v", err)
	}

	if stCloser != nil {
		stCloser.Close()
	}
	log.Printf("INFO: server stopped")
}
