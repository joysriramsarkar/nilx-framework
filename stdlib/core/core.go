// Package core provides fundamental built-in helpers, types, and assertions for NilLang.
package core

import (
	"fmt"
	"strings"
)

// Assert checks a condition and panics with a formatted message if false.
func Assert(condition bool, message string, args ...interface{}) {
	if !condition {
		if len(args) > 0 {
			panic(fmt.Sprintf("assertion failed: "+message, args...))
		}
		panic(fmt.Sprintf("assertion failed: %s", message))
	}
}

// TypeOf returns the runtime type name of any value.
func TypeOf(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case bool:
		return "bool"
	case int, int32, int64:
		return "i32"
	case float32, float64:
		return "f64"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return fmt.Sprintf("%T", v)
	}
}

// ToString converts any value to its string representation.
func ToString(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		return val
	case []string:
		return "[" + strings.Join(val, ", ") + "]"
	default:
		return fmt.Sprintf("%v", val)
	}
}
