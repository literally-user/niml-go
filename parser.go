package niml_go

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Parser represents a NIML file parser.
type Parser struct{}

// NewParser creates a new instance of the NIML parser.
func NewParser() Parser {
	return Parser{}
}

// Parse Reads a NIML file at the specified path and populates the provided struct.
// The config parameter must be a pointer to a struct.
// Struct fields can use the `niml:"key"` tag to specify the key name in the file.
// If no tag is specified, the field name is used.
//
// Example:
//
//	type Config struct {
//	    Host string `niml:"host"`
//	    Port int    `niml:"port"`
//	}
//
//	var cfg Config
//	err := parser.Parse("config.niml", &cfg)
func (Parser) Parse(path string, config interface{}) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to struct")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	configMap, err := parseFile(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	return fillStruct(configMap, config)
}

// parseFile parses the contents of a NIML file and returns a map with the data.
// NIML format:
//   - (topic) - topic/section declaration
//   - / key = "value" - key-value pair declaration
//   - ;; comment - single-line comment
//
// Comments can be either standalone lines or inline after values.
func parseFile(data string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	lines := strings.Split(data, "\n")

	var currentTopic string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, ";;") {
			continue
		}

		if idx := strings.Index(line, ";;"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
			currentTopic = strings.Trim(line, "()")
			if _, exists := result[currentTopic]; !exists {
				result[currentTopic] = make(map[string]interface{})
			}
			continue
		}

		if strings.HasPrefix(line, "/") {
			line = strings.TrimPrefix(line, "/")
			line = strings.TrimSpace(line)

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, `"`)

			if currentTopic != "" {
				if topicMap, ok := result[currentTopic].(map[string]interface{}); ok {
					topicMap[key] = value
				}
			} else {
				result[key] = value
			}
		}
	}

	return result, nil
}

// fillStruct recursively populates struct fields with values from the map.
// Automatically converts data types from strings to the corresponding field types.
// Supports nested structs.
func fillStruct(configMap map[string]interface{}, config interface{}) error {
	v := reflect.ValueOf(config).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		key := getFieldKey(field)

		mapValue, exists := configMap[key]
		if !exists {
			continue
		}

		if err := setFieldValue(fieldValue, mapValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

// getFieldKey extracts the key name for a struct field.
// First checks for the `niml` tag, if not present uses the field name.
func getFieldKey(field reflect.StructField) string {
	if tag := field.Tag.Get("niml"); tag != "" {
		return tag
	}
	return field.Name
}

// setFieldValue sets the value of a struct field, automatically converting the type.
// Supported types:
//   - string
//   - int, int8, int16, int32, int64
//   - uint, uint8, uint16, uint32, uint64
//   - float32, float64
//   - bool
//   - struct (recursively)
func setFieldValue(fieldValue reflect.Value, mapValue interface{}) error {
	switch fieldValue.Kind() {
	case reflect.String:
		if str, ok := mapValue.(string); ok {
			fieldValue.SetString(str)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := parseToInt(mapValue)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intVal, err := parseToInt(mapValue)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uint64(intVal))

	case reflect.Float32, reflect.Float64:
		floatVal, err := parseToFloat(mapValue)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := parseToBool(mapValue)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolVal)

	case reflect.Struct:
		if nestedMap, ok := mapValue.(map[string]interface{}); ok {
			return fillStruct(nestedMap, fieldValue.Addr().Interface())
		}
	}

	return nil
}

// parseToInt converts a value of various types to int64.
// Supports strings, integers, and floating-point numbers.
func parseToInt(value interface{}) (int64, error) {
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
func parseToFloat(value interface{}) (float64, error) {
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
func parseToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseBool(strings.TrimSpace(v))
	case bool:
		return v, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}
