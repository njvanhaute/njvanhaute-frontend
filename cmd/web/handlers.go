package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"frontend.njvanhaute.com/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) transcriptions(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "transcriptions.html", data)
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
		return
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
		return
	}

	if parsedResp.Error.Email != "" {
		form.AddFieldError("email", "Email address is already in use")
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please check your email for more information.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

// Update the handler so it displays the login page.
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.html", data)
}

type userAuthToken struct {
	AuthToken struct {
		Token  string    `json:"token"`
		Expiry time.Time `json:"expiry"`
	} `json:"authentication_token"`
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	postBody, err := json.Marshal(map[string]string{
		"email":    form.Email,
		"password": form.Password,
	})

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	postBuffer := bytes.NewBuffer(postBody)
	resp, err := app.httpClient.Post(app.buildURL("/v1/tokens/authentication"),
		"application/json", postBuffer)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		form.AddNonFieldError("Invalid credentials. Please try again.")
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnauthorized, "login.html", data)
		return
	}

	var tokenResp userAuthToken

	err = app.readJSON(resp, &tokenResp)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserToken", tokenResp.AuthToken.Token)
	app.logger.Info("token", "token", tokenResp.AuthToken.Token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserToken")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type userActivateForm struct {
	Token               string `form:"token"`
	validator.Validator `form:"-"`
}

func (app *application) userActivate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userActivateForm{}
	app.render(w, r, http.StatusOK, "activate.html", data)
}

type userActivateApiError struct {
	Error struct {
		Token string `json:"token"`
	} `json:"error"`
}

func (app *application) userActivatePost(w http.ResponseWriter, r *http.Request) {
	var form userActivateForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Token), "token", "This field cannot be blank")
	form.CheckField(len(form.Token) == 26, "token", "The token must be 26 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "activate.html", data)
		return
	}

	putBody, err := json.Marshal(map[string]string{
		"token": form.Token,
	})

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	putBuffer := bytes.NewBuffer(putBody)
	req, err := http.NewRequest(http.MethodPut, app.buildURL("/v1/users/activate"), putBuffer)
	if err != nil {
		app.serverError(w, r, err)
	}

	req.Header.Set("Content-Type", "application/json")
	rawResp, err := app.httpClient.Do(req)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	defer rawResp.Body.Close()

	var parsedResp userActivateApiError

	err = app.readJSON(rawResp, &parsedResp)
	if nil == err && parsedResp.Error.Token != "" {
		form.AddFieldError("token", "Invalid or expired activation token")
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "activate.html", data)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your account has been activated! You can log in now.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
