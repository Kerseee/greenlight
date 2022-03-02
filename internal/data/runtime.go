package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Runtime is used for movie runtime in minutes.
type Runtime int32

var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// MarshalJSON return the JSON-encoded value of Rumtime r.
// It satisfies the json.Marshaler interface
func (r Runtime) MarshalJSON() ([]byte, error) {
	// jsonValue is the string format for encoding r to JSON.
	jsonValue := fmt.Sprintf("%d mins", r)

	// Add double quotes to the JSON value to form a valid JSON string.
	// Otherwise a runtime error will be raised when calling json.Marshal().
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

// UnmarshalJSON decodes the jsonValue and store the decoded value into r.
// It satisfies the json.Unmarshaledr interface
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// Remove the double-quote.
	unquoteJSONVal, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string and check if all parts are valid.
	parts := strings.Split(unquoteJSONVal, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Parse the minute in the parts into int32.
	mins, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Store the value of mins into r.
	*r = Runtime(mins)

	return nil
}
