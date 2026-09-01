// Package json provides JSON parsing and stringification for NilLang.
package json

import (
	"encoding/json"
	"fmt"
)

// Stringify serializes any data structure to a JSON string.
func Stringify(v interface{}, pretty bool) (string, error) {
	if pretty {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("json.stringify error: %w", err)
		}
		return string(b), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("json.stringify error: %w", err)
	}
	return string(b), nil
}

// Parse decodes a JSON string into generic Go data structures.
func Parse(jsonStr string) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("json.parse error: %w", err)
	}
	return result, nil
}

// Valid checks whether a string contains valid JSON syntax.
func Valid(jsonStr string) bool {
	return json.Valid([]byte(jsonStr))
}
