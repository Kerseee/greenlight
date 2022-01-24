package main

import (
	"fmt"
	"net/http"
	"time"

	"greenlight.kerseeehuang.com/internal/data"
)

// createMovieHandler create a movie and store into DB.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// showMovieHandler shows a movie information.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Get the params in the request context.
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	movie := data.Movie{
		ID:       id,
		CreateAt: time.Now(),
		Title:    "Jujutsu Kaisen 0",
		Runtime:  105,
		Genres:   []string{"anime", "horror", "romance"},
		Version:  1,
	}

	err = app.writeJSON(w, http.StatusOK, movie, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Server Internal Error!", http.StatusInternalServerError)
	}

}
