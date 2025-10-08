// Our previous code use
// log.Fatal(http.ListenAndServe(":8080", mux))

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/arya2004/gobanter/pkg/config"
	"github.com/arya2004/gobanter/pkg/handlers"
	"github.com/arya2004/gobanter/pkg/routes"
)

// main is the entry point of the application

func main() {

	// initialize the application routes

	cfg := config.Load()

	// Initialize the application routes
	mux := routes.Routes()

	//start a goroutine to listen for websocket messages

	log.Println("Starting WebSocket Channel listener...")
	go handlers.ListenToWsChannel()

	// creating a custon HTTP server for better control

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// channel to catch OS signals like ctrl + c

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// starting the  server in a goroutine

	go func() {
		log.Println("Server starting on port 8080....")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server startup error: %v", err)
		}
	}()

	// block until an interrupt signal is received

	<-stop
	log.Println("Shutdown signal received, attempting graceful shutdown...")

	// creating a deadline for shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// attempt to gracefully shut down the server

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	} else {
		log.Println("Server shut down cleanly")
	// Start the HTTP server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on port %s...", cfg.Port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
