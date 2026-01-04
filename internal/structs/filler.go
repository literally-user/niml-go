package structs

import (
	"fmt"
	"reflect"
	
	"github.com/literally-user/niml-go/internal/types"
)

// fillStruct recursively populates struct fields with values from the map.
// Automatically converts data types from strings to the corresponding field types.
// Supports nested structs.
func FillStruct(configMap map[string]interface{}, config interface{}) error {
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
		intVal, err := types.ParseToInt(mapValue)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intVal, err := types.ParseToInt(mapValue)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uint64(intVal))

	case reflect.Float32, reflect.Float64:
		floatVal, err := types.ParseToFloat(mapValue)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatVal)

	case reflect.Bool:
		boolVal, err := types.ParseToBool(mapValue)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolVal)

	case reflect.Struct:
		if nestedMap, ok := mapValue.(map[string]interface{}); ok {
			return FillStruct(nestedMap, fieldValue.Addr().Interface())
		}
	}

	return nil
}
