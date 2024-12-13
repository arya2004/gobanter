package main

import (
	"log"
	"net/http"

	"github.com/arya2004/gobanter/pkg/routes"
)

func main(){
	mux := routes.Routes()

	log.Println("starting on port 8083")

	_ = http.ListenAndServe(":8083", mux)

}