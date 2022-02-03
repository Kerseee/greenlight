package data

import (
	"database/sql"
	"errors"
)

// ErrRecordNotFound is an error that shows "record not found".
var ErrRecordNotFound = errors.New("record not found")

// Models holds all data models used in the whole project.
type Models struct {
	Movies MovieModel
}

// NewModels return an instance of Models with given db.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
