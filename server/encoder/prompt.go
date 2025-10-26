package encoder

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// JSONWithPrompts generates JSON schema by replacing values with prompt tag descriptions.
type JSONWithPrompts struct {
	Value any
}

// MarshalJSON converts the struct to JSON using prompt tags as values.
func (j JSONWithPrompts) MarshalJSON() ([]byte, error) {
	result := buildPromptMap(reflect.ValueOf(j.Value))
	return json.Marshal(result)
}

// Recursively builds a map[string]any (or []any) representing the structure
// but filled with prompt tags instead of real values.
func buildPromptMap(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return buildStructPromptMap(v)
	case reflect.Slice:
		return buildSlicePromptMap(v)
	default:
		return nil
	}
}

func buildStructPromptMap(v reflect.Value) map[string]any {
	t := v.Type()
	result := make(map[string]any)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // skip unexported
			continue
		}

		jsonTag := parseJSONTag(field)
		if jsonTag == "" {
			continue
		}

		fv := v.Field(i)
		ft := fv.Type()

		value := buildFieldPromptValue(field, fv, ft)
		if value != nil {
			result[jsonTag] = value
		}
	}
	return result
}

func parseJSONTag(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return field.Name
	}
	if comma := indexComma(jsonTag); comma >= 0 {
		jsonTag = jsonTag[:comma]
	}
	if jsonTag == "-" {
		return ""
	}
	return jsonTag
}

func buildFieldPromptValue(field reflect.StructField, fv reflect.Value, ft reflect.Type) any {
	// Handle slice of structs
	if ft.Kind() == reflect.Slice {
		return buildSliceFieldPromptValue(field, ft)
	}

	// Handle nested struct
	if ft.Kind() == reflect.Struct && ft.String() != "time.Time" {
		return buildPromptMap(fv)
	}

	// Handle pointer to struct
	if ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct {
		return buildPointerFieldPromptValue(fv, ft)
	}

	// Base case: use prompt tag
	prompt := field.Tag.Get("prompt")
	if prompt != "-" {
		return fmt.Sprintf("(%s) %s", ft.String(), prompt)
	}
	return nil
}

func buildSliceFieldPromptValue(field reflect.StructField, ft reflect.Type) any {
	prompt := field.Tag.Get("prompt")
	switch {
	case prompt == "-":
		return []any{}
	case prompt != "":
		return []any{prompt}
	default:
		return []any{
			buildPromptMap(reflect.New(ft.Elem())),
		}
	}
}

func buildPointerFieldPromptValue(fv reflect.Value, ft reflect.Type) any {
	if fv.IsNil() {
		return buildPromptMap(reflect.New(ft.Elem()).Elem())
	}
	return buildPromptMap(fv)
}

func buildSlicePromptMap(v reflect.Value) any {
	if v.Len() == 0 {
		return []any{}
	}
	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() == reflect.Struct {
		elem := reflect.New(elemType).Elem()
		return []any{buildPromptMap(elem)}
	} else if elemType.Kind() == reflect.String {
		return []string{"<string>"}
	}
	return []any{}
}

func indexComma(s string) int {
	for i := range s {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}
