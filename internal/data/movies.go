package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

const dbTimeOut = 3 * time.Second // 3s timeout for all CRUD

// Insert inserts a movie into DB.
func (m MovieModel) Insert(movie *Movie) error {
	// Define the sql query for inserting.
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, create_at, version`

	// Declare arguments array for values in the above query.
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Create a context with 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute query and return the returning error.
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreateAt, &movie.Version)
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

	// Create time-out context.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Retrieve the movie from movies table in DB.
	var movie Movie
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
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

// GetAll return a slice of movies based on given title, genres, and filters.
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	// Define the query of getting results.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, create_at, title, year, runtime, genres, version
		FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// Create a context with 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute the query.
	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	// Read the rows and store information into movies.
	var totalRecords int
	movies := []*Movie{}
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreateAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		movies = append(movies, &movie)
	}

	// Check if any error happens during row scan.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate the metadata based on totalRecords.
	metadata := calculateMatadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}

// Update updates the information of given movie in DB.
// Return data.ErrEditConflict if conflict happens.
func (m MovieModel) Update(movie *Movie) error {
	// Define the query of updating movie.
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	// Create the argument array for the above query.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// Create a context with 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute the query with arguments and scan the new version value into the movie struct.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
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

	// Create a context with 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeOut)
	defer cancel()

	// Execute the query and get the execution result from DB.
	result, err := m.DB.ExecContext(ctx, query, id)
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
