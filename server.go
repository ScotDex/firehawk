package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func startHealthCheckServer() {
	// Cloud Run provides the port to listen on via the 'PORT' environment variable.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified (for local testing)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Firehawk bot is running.")
	})

	log.Printf("Health check server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start health check server: %v", err)
	}
}
