package encoder

import "reflect"

func indexComma(s string) int {
	for i := range s {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}

func jsonTag(field reflect.StructField) string {
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
