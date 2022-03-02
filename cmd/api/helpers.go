package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"greenlight.kerseeehuang.com/internal/validator"
)

// envelope envelopes the struct that will be writen to responses in JSON.
type envelope map[string]interface{}

// readIDParam retrieves the id in http.Request, parse it to int64 and return.
// If the id cannot be parsed to int64 or id < 1 then return 0 and an error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// Get the params in the request context.
	params := httprouter.ParamsFromContext(r.Context())

	// Retrive the movie id in params, parse it, and validate it.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// writeJSON is a helper for sending responses in JSON.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline for better layout shown in terminal.
	js = append(js, '\n')

	// Add the headers into the response header.
	for k, v := range headers {
		w.Header()[k] = v
	}

	// Set the "Content-Type" header to JSON.
	w.Header().Set("Content-Type", "application/json")

	// Write the response header.
	w.WriteHeader(status)

	// Write the response body.
	w.Write(js)

	return nil
}

// readJSON reads the JSON-encoded request, decodes it, and store result into dst.
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Limit the size of body to 1MB.
	maxBytes := 1 << (10 * 2)
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Declare a new decoder that disallow unknown fields.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the JSON-encoded request and store the result it into dst.
	err := dec.Decode(dst)

	// Triage the error. Determine which type the err is.
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshaledError *json.InvalidUnmarshalError

		switch {
		// Check syntax error.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// Check unmarshal type error.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body containts incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// Check empty body error.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// Check unknown field error.
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// Check size error.
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		// Panic the internal error.
		case errors.As(err, &invalidUnmarshaledError):
			panic(err)

		default:
			return err
		}
	}

	// Check if the request body only contains single json.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// readString returns the value from query qs with given key.
// If the given key does not exist in qs, return defaultVal.
func (app *application) readString(qs url.Values, key, defaultVal string) string {
	// Extract the value with given key from qs.
	val := qs.Get(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// readCSV reads the string value from qs with given key, splits the string with ","
// and then return the list of splited string.
// If the given key does not exist in qs, return defaultVals.
func (app *application) readCSV(qs url.Values, key string, defaultVals []string) []string {
	// Extract values from qs.
	csv := qs.Get(key)
	if csv == "" {
		return defaultVals
	}
	return strings.Split(csv, ",")
}

// readInt reads the value from qs with given key, parse the value from string to integer and return.
// If parsing error happens, store error messages into v.
func (app *application) readInt(qs url.Values, key string, defaultVal int, v *validator.Validator) int {
	// Extract value from qs.
	val := qs.Get(key)
	if val == "" {
		return defaultVal
	}

	// Parse the value.
	intVal, err := strconv.Atoi(val)
	if err != nil {
		v.AddError(key, "must be integer")
		return defaultVal
	}
	return intVal
}

// background opens a goroutine to execute f with recover.
func (app *application) background(f func()) {
	app.wg.Add(1)
	go func() {
		defer app.wg.Done()

		// Recover from panic.
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err).Error(), nil)
			}
		}()

		// Execute function f.
		f()
	}()
}
