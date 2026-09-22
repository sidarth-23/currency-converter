package main

import (
	"log"
	"net/http"

	"github.com/example/currency-watcher/backend/internal/config"
	"github.com/example/currency-watcher/backend/internal/server"
)

// main starts the HTTP API and terminates the process when initialization or serving fails unexpectedly.
func main() {
	httpServer, err := server.NewHTTPServer(config.Load())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("currency watcher API listening on %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
