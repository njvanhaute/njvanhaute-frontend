package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"frontend.njvanhaute.com/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) tuneView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Display a specific tune with ID %d...", id)
}

func (app *application) tuneCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Display a form for creating a new tune..."))
}

func (app *application) tuneCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Save a new tune..."))
}

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.html", data)
}

type userSignupApiError struct {
	Error struct {
		Email string `json:"email"`
	} `json:"error"`
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
	}

	postBody, err := json.Marshal(map[string]string{
		"name":     form.Name,
		"email":    form.Email,
		"password": form.Password,
	})

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	postBuffer := bytes.NewBuffer(postBody)
	rawResp, err := app.httpClient.Post(app.buildURL("/v1/users"),
		"application/json", postBuffer)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	defer rawResp.Body.Close()

	var parsedResp userSignupApiError

	err = app.readJSON(rawResp, &parsedResp)
	if err != nil {
		app.serverError(w, r, err)
	}

	if parsedResp.Error.Email != "" {
		form.AddFieldError("email", "Email address is already in use")
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		app.logger.Info("Response from json", "json", parsedResp)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please check your email for more information.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Display a form for logging in a user...")
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Authenticate and login the user...")
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Logout the user...")
}
