package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

// routes routes requests to the corresponding handlers.
func (app *application) routes() http.Handler {
	// Initialize the router.
	router := httprouter.New()

	// Customize the router.NotFound handler so that we can send JSON-encoded
	// error messages when error happends during routing.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register the methods, URL pattern, and handlers.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// Create a middleware chain.
	chain := alice.New(app.recoverPanic)
	if app.config.limiter.enabled {
		chain = chain.Append(app.rateLimit)
	}

	return chain.Then(router)
}
