package encoder

import (
	"reflect"
	"strings"
)

// JSONSchemaProperty represents a property in a JSON Schema
type JSONSchemaProperty map[string]interface{}

// BuildJSONSchema generates a strict JSON Schema from a struct type using prompt tags for descriptions
func JSONSchema(v interface{}) map[string]interface{} {
	return jsonSchemaForType(reflect.TypeOf(v))
}

func jsonSchemaForType(t reflect.Type) map[string]interface{} {
	// Handle pointer types
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
	}

	properties := make(map[string]interface{})
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Parse JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Extract field name and omitempty flag
		fieldName, omitempty := parseJSONTagForSchema(jsonTag)
		if fieldName == "" {
			continue
		}

		// Get prompt tag for description (may include enum)
		promptTag := field.Tag.Get("prompt")
		if promptTag == "-" {
			continue
		}

		// Parse prompt tag: "description;enum:val1,val2,val3"
		description, enumValues := parsePromptTag(promptTag)

		// Build property schema
		prop := buildPropertySchema(field.Type, description, enumValues)
		if prop != nil {
			properties[fieldName] = prop

			// Add to required if not omitempty
			if !omitempty {
				required = append(required, fieldName)
			}
		}
	}

	schema["properties"] = properties
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func parseJSONTagForSchema(tag string) (string, bool) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 || parts[0] == "-" {
		return "", false
	}

	fieldName := parts[0]
	omitempty := false

	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitempty = true
			break
		}
	}

	return fieldName, omitempty
}

// parsePromptTag parses "description;enum:val1,val2" into description and enum values
func parsePromptTag(tag string) (string, []string) {
	parts := strings.Split(tag, ";")
	description := parts[0]

	var enumValues []string
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "enum:") {
			values := strings.TrimPrefix(part, "enum:")
			for _, v := range strings.Split(values, ",") {
				enumValues = append(enumValues, strings.TrimSpace(v))
			}
		}
	}

	return description, enumValues
}

func buildPropertySchema(t reflect.Type, description string, enumValues []string) map[string]interface{} {
	// Handle pointer types
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	prop := make(map[string]interface{})

	switch t.Kind() {
	case reflect.String:
		prop["type"] = "string"
		if len(enumValues) > 0 {
			enumIface := make([]interface{}, len(enumValues))
			for i, v := range enumValues {
				enumIface[i] = v
			}
			prop["enum"] = enumIface
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		prop["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		prop["type"] = "number"
	case reflect.Bool:
		prop["type"] = "boolean"
	case reflect.Slice, reflect.Array:
		prop["type"] = "array"
		// Handle slice element type
		elemType := t.Elem()
		if elemType.Kind() == reflect.String {
			prop["items"] = map[string]interface{}{"type": "string"}
		} else if elemType.Kind() == reflect.Struct || (elemType.Kind() == reflect.Pointer && elemType.Elem().Kind() == reflect.Struct) {
			prop["items"] = jsonSchemaForType(elemType)
		}
	case reflect.Struct:
		// Check if it's a special type (time.Time, etc.)
		if t.String() == "time.Time" || t.String() == "gorm.DeletedAt" {
			return nil // Skip time fields
		}
		// Handle nested structs
		return jsonSchemaForType(t)
	default:
		return nil
	}

	// Add description from prompt tag if available
	if description != "" && description != "+" {
		prop["description"] = description
	}

	return prop
}
