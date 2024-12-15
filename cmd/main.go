package main

import (
	"log"
	"net/http"

	"github.com/arya2004/gobanter/pkg/handlers"
	"github.com/arya2004/gobanter/pkg/routes"
)

// main is the entry point of the application
func main() {
	// Initialize the application routes
	mux := routes.Routes()

	// Start a goroutine to listen for WebSocket messages
	log.Println("Starting WebSocket channel listener...")
	go handlers.ListenToWsChannel()

	// Start the HTTP server on port 8080
	log.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
