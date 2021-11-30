package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes routes requests to the corresponding handlers.
func (app *application) routes() *httprouter.Router {
	// Initialize the router.
	router := httprouter.New()

	// Register the methods, URL pattern, and handlers.
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	return router
}