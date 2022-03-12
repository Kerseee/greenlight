package main

import (
	"fmt"
	"net/http"
)

// logError log the error via app.logger.
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err.Error(), map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})
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

// badRequestResponse sends the Bad Request response to the client.
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// failedValidationResponse sends the Unprocessable Entity Error response to the client
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// editConflictResponse sends the Conflict Error response to the client.
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	msg := "unable to update due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, msg)
}

// rateLimitExceededResponse sends the Rate Limit Exceed Error response to the client.
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	msg := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, msg)
}

// invalidCredentialsResponse sends the Status Unauthorized Error response to the client.
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	msg := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// invalidAuthenticationTokenResponse adds "WWW-Authenticate":"Bearer" into response header
// and sends the Status Unauthorized Error response to the client.
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	msg := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// authenticationRequiredResponse sends Status Unauthorized Error resonse to the client.
// Called when authentication is needed but the client is not authenticated.
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	msg := "require authentication to this end point"
	app.errorResponse(w, r, http.StatusUnauthorized, msg)
}

// invalidAccountResponse sends Forbidden Error response to the client.
// Called when the client's account is not activated.
func (app *application) invalidAccountResponse(w http.ResponseWriter, r *http.Request) {
	msg := "require activation of your account to this end point"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}

// notPermittedResponse sends Forbidden Error response to the client.
// Called when the client is permitted for an end point.
func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	msg := "your account has no permission for this resource"
	app.errorResponse(w, r, http.StatusForbidden, msg)
}
