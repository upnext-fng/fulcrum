package store

import (
	"reflect"
	"strings"
)

// hasField checks if a struct type has a field with the given name
func hasField(structType reflect.Type, fieldName string) bool {
	if structType.Kind() != reflect.Struct {
		return false
	}

	// Check direct fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Check field name
		if field.Name == fieldName {
			return true
		}

		// Check GORM column tag
		if gormTag := field.Tag.Get("gorm"); gormTag != "" {
			if strings.Contains(gormTag, "column:"+fieldName) {
				return true
			}
		}

		// Check json tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName == fieldName {
				return true
			}
		}

		// Check embedded structs
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if hasField(field.Type, fieldName) {
				return true
			}
		}
	}

	return false
}
