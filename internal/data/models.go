package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found") // record not found
	ErrEditConflict   = errors.New("edit conflict")    // edit conflict
)

// Models holds all data models used in the whole project.
type Models struct {
	Movies MovieModel
	Tokens TokenModel
	Users  UserModel
}

// NewModels return an instance of Models with given db.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
		Tokens: TokenModel{DB: db},
		Users:  UserModel{DB: db},
	}
}
