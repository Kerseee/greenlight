package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"greenlight.kerseeehuang.com/internal/validator"
)

// Movie stores all information of each movie.
type Movie struct {
	ID       int64     `json:"id"`
	CreateAt time.Time `json:"-"` // Use `-` tag to unshow this field to users.
	Title    string    `json:"title,omitempty"`
	Year     int32     `json:"year,omitempty"`
	Runtime  Runtime   `json:"runtime,omitempty"` // Movie runtime in minutes
	Genres   []string  `json:"genres"`
	Version  int32     `json:"version"`
}

// MovieModel is a wrapper of *sql.DB
type MovieModel struct {
	DB *sql.DB
}

// Insert inserts a movie into DB.
func (m MovieModel) Insert(movie *Movie) error {
	// Define the sql query for inserting.
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, create_at, version`

	// Declare arguments array for values in the above query.
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Execute query and return the returning error.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreateAt, &movie.Version)
}

// Get retrives a movie given movie id from DB.
// Return nil, data.ErrRecordNotFound if there is no matching result in DB.
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the sql query for getting.
	query := `
		SELECT id, create_at, title, year, runtime, genres, version
		FROM movies
		Where id = $1`

	// Retrieve the movie from movies table in DB.
	var movie Movie
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreateAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// Update updates the information of given movie in DB. #TODO
func (m MovieModel) Update(movie *Movie) error {
	// Define the query of updating movie.
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version`

	// Create the argument array for the above query.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	// Execute the query with arguments and scan the new version value into the movie struct.
	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

// Delete deletes the movie with given id from DB.
func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	// Define the query for deleting movie from DB.
	query := `
		DELETE FROM movies
		WHERE id = $1`

	// Execute the query and get the execution result from DB.
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	// Check if there is any error by confirming the number of effected rows.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
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
