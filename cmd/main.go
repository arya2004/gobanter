package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arya2004/gobanter/pkg/config"
	"github.com/arya2004/gobanter/pkg/handlers"
	"github.com/arya2004/gobanter/pkg/routes"
)

// main is the entry point of the application
func main() {
	cfg := config.Load()

	// Initialize the application routes
	mux := routes.Routes()

	// Start a goroutine to listen for WebSocket messages
	log.Println("Starting WebSocket channel listener...")
	go handlers.ListenToWsChannel()

	// Start the HTTP server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on port %s...", cfg.Port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
