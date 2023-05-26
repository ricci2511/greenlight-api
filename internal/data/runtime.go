package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

// Implementation of json.Marshaler interface for the Runtime type.
//
// Encodes a Runtime of type int32 in the format of "<runtime> mins".
func (r Runtime) MarshalJSON() ([]byte, error) {
	jsonValue := fmt.Sprintf("%d mins", r)

	// Wrap string values in double quotes to make it a valid JSON.
	quotedJSONValue := strconv.Quote(jsonValue)

	return []byte(quotedJSONValue), nil
}
