package main

import (
	"fmt"
	"net/http"
)

// createMovieHandler create a movie and store into DB.
func (app *application) createMovieHandler(w http.ResponseWriter, r * http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// showMovieHandler shows a movie information.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Get the params in the request context.
	id, err := app.readIDParam(r)
	if err != nil{
		http.NotFound(w, r)
		return
	}

	// Show the movie with given id.
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}