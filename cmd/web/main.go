package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from me!"))
}

func tuneView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	msg := fmt.Sprintf("Display a specific tune with ID %d...", id)
	w.Write([]byte(msg))
}

func tuneCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Display a form for adding a new tune..."))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", home)
	mux.HandleFunc("/tune/view/{id}", tuneView)
	mux.HandleFunc("/tune/create", tuneCreate)

	log.Print("starting server on :4000")

	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
