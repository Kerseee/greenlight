package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes routes requests to the corresponding handlers.
func (app *application) routes() *httprouter.Router {
	// Initialize the router.
	router := httprouter.New()

	// Customize the router.NotFound handler so that we can send JSON-encoded
	// error messages when error happends during routing.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register the methods, URL pattern, and handlers.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	return router
}
