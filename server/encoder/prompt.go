package encoder

import (
	"encoding/json"
	"reflect"
)

// JSONWithPrompts replaces struct field values with their `prompt` tag values.
// Works recursively on nested structs and slices of structs.
type JSONWithPrompts struct {
	Value any
}

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
		t := v.Type()
		result := make(map[string]any)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" { // skip unexported
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = field.Name
			} else {
				if comma := indexComma(jsonTag); comma >= 0 {
					jsonTag = jsonTag[:comma]
				}
			}
			if jsonTag == "-" {
				continue
			}

			fv := v.Field(i)
			ft := fv.Type()

			// Handle slice of structs
			if ft.Kind() == reflect.Slice {
				prompt := field.Tag.Get("prompt")
				if prompt == "-" {
					result[jsonTag] = []any{}
				} else if prompt != "" {
					result[jsonTag] = []any{prompt}
				} else {
					result[jsonTag] = []any{
						buildPromptMap(reflect.New(ft.Elem())),
					}
				}
				continue
			}

			// Handle nested struct
			if ft.Kind() == reflect.Struct && ft.String() != "time.Time" {
				result[jsonTag] = buildPromptMap(fv)
				continue
			}

			// Handle pointer to struct
			if ft.Kind() == reflect.Pointer && ft.Elem().Kind() == reflect.Struct {
				if fv.IsNil() {
					result[jsonTag] = buildPromptMap(reflect.New(ft.Elem()).Elem())
				} else {
					result[jsonTag] = buildPromptMap(fv)
				}
				continue
			}

			// Base case: use prompt tag
			prompt := field.Tag.Get("prompt")
			if prompt != "-" {
				result[jsonTag] = prompt
			}
		}
		return result
	case reflect.Slice:
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
	default:
		return nil
	}
}

func indexComma(s string) int {
	for i := range s {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}
