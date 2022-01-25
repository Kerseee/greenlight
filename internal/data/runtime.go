package data

import (
	"fmt"
	"strconv"
)

// Runtime is used for movie runtime in minutes.
type Runtime int32

// MarshalJSON return the JSON-encoded value of Rumtime r.
func (r Runtime) MarshalJSON() ([]byte, error) {
	// jsonValue is the string format for encoding r to JSON.
	jsonValue := fmt.Sprintf("%d mins", r)

	// Add double quotes to the JSON value to form a valid JSON string.
	// Otherwise a runtime error will be raised when calling json.Marshal().
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}
