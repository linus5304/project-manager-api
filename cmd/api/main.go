package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/linus5304/project-manager-api/internal/httpapi"
	"github.com/linus5304/project-manager-api/internal/store"
)

func main() {
	addr := ":4000"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	var st store.ProjectStore

	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		st = store.NewMemoryStore()
	} else {
		startCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		pg, err := store.NewPostgresStore(startCtx, dsn)
		if err != nil {
			log.Fatalf("unable to connect to database: %v", err)
		}
		defer pg.Close()
		st = pg
	}

	app := httpapi.NewApplication(st)

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("starting server on %s", addr)
	log.Fatal(srv.ListenAndServe())
}
