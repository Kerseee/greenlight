package main

import (
	"fmt"
	"net/http"
	"time"

	"greenlight.kerseeehuang.com/internal/data"
)

// createMovieHandler decodes the information from the request,
// create a movie and store it into DB.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to holding decoded input.
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Decode the movie information from the request.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Print the decoded struct just for now.
	fmt.Fprintf(w, "%+v\n", input)
}

// showMovieHandler shows a movie information.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Get the params in the request context.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Create a dummy movie.
	movie := data.Movie{
		ID:       id,
		CreateAt: time.Now(),
		Title:    "Jujutsu Kaisen 0",
		Runtime:  105,
		Genres:   []string{"anime", "horror", "romance"},
		Version:  1,
	}

	// Write the responses with movie in the JSON form.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
