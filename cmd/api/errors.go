package main

import (
	"fmt"
	"net/http"
)

// logError log the error via app.logger.
func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

// errorResponse sends JSON-encoded error messages to client with error code.
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	// Envelope the error message for better format of JSON.
	env := envelope{"error": message}

	// Write the error messages to the client.
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// serverErrorResponse logs the error and sends Server Internal Error to the client.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// Log the error.
	app.logError(r, err)

	msg := "Server Intenal Error! The server cannot process your request now."
	app.errorResponse(w, r, http.StatusInternalServerError, msg)
}

// notFoundResponse sends the Not Found Error to client in JSON form.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "The source could not be found!"
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

// methodNotAllowedResponse sends the Method Not Allowed Error to the client.
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf("The %s method is not allowed for this resource.", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, msg)
}