package main

import (
	"fmt"
	"net/http"
)

// Will be extended later on.
func (app *application) logError(r *http.Request, err error) {
	app.logger.Print(err)
}

// Helper to send JSON-formatted error messages to the client.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	err := app.writeJSON(w, status, envelope{"error": message}, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Wrapper around errorResponse() that is used when our app encounters an unexpected problem at runtime.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// server errors are logged
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// Wrapper around errorResponse() that is used when the user submits a bad request.
func (app *application) notFoundReponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// Wrapper around errorResponse() that is used when the user tries to perform an invalid action against a resource.
func (app *application) methodNotAllowedRespone(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}
