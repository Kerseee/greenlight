package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"greenlight.kerseeehuang.com/internal/data"
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

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission(data.PermissionReadMovies, app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission(data.PermissionWriteMovies, app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission(data.PermissionReadMovies, app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission(data.PermissionWriteMovies, app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission(data.PermissionWriteMovies, app.deleteMovieHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Create a middleware chain.
	chain := alice.New(app.recoverPanic, app.enableCORS)
	if app.config.limiter.enabled {
		chain = chain.Append(app.rateLimit)
	}
	chain = chain.Append(app.authenticate)

	return chain.Then(router)
}
