package types

import (
	"strconv"
	"strings"
	"fmt"
)

// parseToInt converts a value of various types to int64.
// Supports strings, integers, and floating-point numbers.
func ParseToInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// parseToFloat converts a value of various types to float64.
// Supports strings and numeric types.
func ParseToFloat(value interface{}) (float64, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseFloat(strings.TrimSpace(v), 64)
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float", value)
	}
}

// parseToBool converts a value to boolean.
// Supports strings ("true", "false", "1", "0") and bool values.
func ParseToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseBool(strings.TrimSpace(v))
	case bool:
		return v, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}
