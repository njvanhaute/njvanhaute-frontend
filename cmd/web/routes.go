package main

import (
	"net/http"

	"frontend.njvanhaute.com/ui"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /tune/view/{id}", dynamic.ThenFunc(app.tuneView))
	mux.Handle("GET /tune/create", dynamic.ThenFunc(app.tuneCreate))
	mux.Handle("POST /tune/create", dynamic.ThenFunc(app.tuneCreatePost))

	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/activate", dynamic.ThenFunc(app.userActivate))
	mux.Handle("POST /user/activate", dynamic.ThenFunc(app.userActivatePost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	mux.Handle("POST /user/logout", dynamic.ThenFunc(app.userLogoutPost))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)
}
