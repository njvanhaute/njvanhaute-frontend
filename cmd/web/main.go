package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /tune/view/{id}", tuneView)
	mux.HandleFunc("GET /tune/create", tuneCreate)
	mux.HandleFunc("POST /tune/create", tuneCreatePost)
	log.Print("starting server on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
