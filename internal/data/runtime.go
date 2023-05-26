package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidRuntimeValue = errors.New("invalid runtime value")

type Runtime int32

// Implementation of json.Marshaler interface.
//
// Encodes a Runtime of type int32 in the format of "<runtime> mins".
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	// Wrap string values in double quotes to make it a valid JSON.
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}

// Implementation of json.Unmarshaler interface.
//
// Decodes a Runtime of format "<runtime> mins" into an int32 Runtime value.
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeValue
	}

	// Split and sanity check to make sure it satisfies the expected format.
	parts := strings.Split(unquotedJSONValue, " ")
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeValue
	}

	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeValue
	}

	// Assign the parsed value to the dereferenced runtime pointer.
	*r = Runtime(i)

	return nil
}
