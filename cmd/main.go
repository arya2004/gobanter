package main

import (
	"log"
	"net/http"

	"github.com/arya2004/gobanter/pkg/handlers"
	"github.com/arya2004/gobanter/pkg/routes"
)

func main(){
	mux := routes.Routes()

	log.Println("starting channel listener")
	go handlers.ListenToWsChannel()

	log.Println("starting on port 8080")

	_ = http.ListenAndServe(":8080", mux)

}