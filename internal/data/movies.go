package data

import (
	"time"

	"greenlight.kerseeehuang.com/internal/validator"
)

type Movie struct {
	ID       int64     `json:"id"`
	CreateAt time.Time `json:"-"` // Use `-` tag to unshow this field to users.
	Title    string    `json:"title,omitempty"`
	Year     int32     `json:"year,omitempty"`
	Runtime  Runtime   `json:"runtime,omitempty"` // Movie runtime in minutes
	Genres   []string  `json:"genres"`
	Version  int32     `json:"version"`
}

// ValidateMove validates movie and store validation into v.Errors
func ValidateMovie(v *validator.Validator, movie *Movie) {
	// Check the title of movie.
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not more than 500 bytes")

	// Check the year of movie.
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must before now")

	// Check the runtime of movie.
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	// Check the genres of movie.
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) <= 5, "genres", "must less than or equal to 5")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(validator.Unique(movie.Genres), "genres", "values in genres must be unique")

}
