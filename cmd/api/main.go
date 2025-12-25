package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/linus5304/project-manager-api/internal/httpapi"
)

func main() {
	addr := ":4000"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	app := httpapi.NewApplication()

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
