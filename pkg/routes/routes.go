package routes

import (
	"net/http"

	"github.com/arya2004/gobanter/pkg/handlers"
	"github.com/bmizerany/pat"
)

// Routes initializes the application's routing configuration
func Routes() http.Handler {
	mux := pat.New()

	// Define route for the home page
	mux.Get("/", http.HandlerFunc(handlers.Home))

	// Define route for WebSocket endpoint
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))

	// Serve static assets from the "assets" directory
	fileServer := http.FileServer(http.Dir("./assets/"))
	mux.Get("/assets/", http.StripPrefix("/assets/", fileServer))

	return mux
}
