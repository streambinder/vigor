package encoder

import (
	"encoding/json"
	"reflect"
)

// FlutterMetadata generates metadata about Flutter-specific tags for a struct.
type FlutterMetadata struct {
	Value any
}

// FieldMetadata contains Flutter-specific metadata for a field.
type FieldMetadata struct {
	Name     string `json:"name"`
	JSONTag  string `json:"json_tag"`
	Required bool   `json:"required"`
	Type     string `json:"type"`
}

// MarshalJSON converts the struct to JSON with Flutter metadata.
func (f FlutterMetadata) MarshalJSON() ([]byte, error) {
	fields := extractFlutterMetadata(reflect.ValueOf(f.Value))
	return json.Marshal(fields)
}

// extractFlutterMetadata extracts Flutter tag metadata from a struct.
func extractFlutterMetadata(v reflect.Value) []FieldMetadata {
	if !v.IsValid() {
		return nil
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	return extractStructMetadata(v)
}

func extractStructMetadata(v reflect.Value) []FieldMetadata {
	t := v.Type()
	var metadata []FieldMetadata

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // skip unexported fields
			continue
		}

		jsonTag := jsonTag(field)
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		flutterTag := field.Tag.Get("flutter")
		if flutterTag == "skip" || flutterTag == "-" {
			continue
		}

		fieldMeta := FieldMetadata{
			Name:     field.Name,
			JSONTag:  jsonTag,
			Required: flutterTag == "required",
			Type:     field.Type.String(),
		}

		metadata = append(metadata, fieldMeta)
	}

	return metadata
}
