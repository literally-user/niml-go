package niml

import (
	"fmt"
	"os"
	"reflect"

	"github.com/literally-user/niml-go/internal/structs"
	"github.com/literally-user/niml-go/internal/file"
)

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
//	err := niml.Parse("config.niml", &cfg)
func Parse(path string, config interface{}) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to struct")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	configMap, err := file.ParseFile(string(data))
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	return structs.FillStruct(configMap, config)
}