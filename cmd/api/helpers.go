package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// readIDParam retrieves the id in http.Request, parse it to int64 and return.
// If the id cannot be parsed to int64 or id < 1 then return 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// Get the params in the request context.
	params := httprouter.ParamsFromContext(r.Context())

	// Retrive the movie id in params, parse it, and validate it.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1{
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}