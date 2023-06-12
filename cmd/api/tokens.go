package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/ricci2511/greenlight-api/internal/data"
	"github.com/ricci2511/greenlight-api/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			v.AddError("email", "no user exists with this email address")
			app.invalidCredentialsResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse email and validate it before generating a password reset token for it.
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		} else {
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// Make sure the user's account is activated.
	if !user.Activated {
		v.AddError("email", "user account must be activated to reset your password")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 45*time.Minute, data.ScopePasswordReset)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the user an email with the password reset token in a background goroutine.
	app.background(func() {
		data := map[string]any{
			"passwordResetToken": token.Plaintext,
		}

		err := app.mailer.Send(user.Email, "token_password_reset.html", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	res := envelope{"message": "an email will be sent to you containing the password reset instructions"}
	err = app.writeJSON(w, http.StatusAccepted, res, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
