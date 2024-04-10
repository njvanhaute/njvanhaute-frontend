package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /tune/view/{id}", app.tuneView)
	mux.HandleFunc("GET /tune/create", app.tuneCreate)
	mux.HandleFunc("POST /tune/create", app.tuneCreatePost)

	return mux
}
