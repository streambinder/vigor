package main

import (
	"strings"
	"unicode"
)

// toSnakeCase converts PascalCase or camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// Insert underscore before uppercase letters (except at start)
			if i > 0 && (unicode.IsLower(rune(s[i-1])) || (i < len(s)-1 && unicode.IsLower(rune(s[i+1])))) {
				result.WriteRune('_')
			}
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// toCamelCase converts snake_case or PascalCase to camelCase
func toCamelCase(s string) string {
	// If already in camelCase, return as is
	if len(s) == 0 {
		return s
	}

	// Handle snake_case
	if strings.Contains(s, "_") {
		parts := strings.Split(s, "_")
		result := strings.ToLower(parts[0])
		for i := 1; i < len(parts); i++ {
			if len(parts[i]) > 0 {
				result += strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
			}
		}
		return result
	}

	// Handle common acronyms (ID, API, URL, etc.)
	if s == "ID" {
		return "id"
	}
	if s == "URL" {
		return "url"
	}
	if strings.HasSuffix(s, "ID") {
		// UserID -> userId
		base := s[:len(s)-2]
		return toCamelCase(base) + "Id"
	}
	if strings.HasSuffix(s, "URL") {
		base := s[:len(s)-3]
		return toCamelCase(base) + "Url"
	}

	// Convert PascalCase to camelCase
	return strings.ToLower(s[:1]) + s[1:]
}

// toClassName converts a string to PascalCase (ClassName)
func toClassName(s string) string {
	// If already PascalCase, return as is
	if len(s) > 0 && unicode.IsUpper(rune(s[0])) {
		return s
	}

	// Handle snake_case
	if strings.Contains(s, "_") {
		parts := strings.Split(s, "_")
		result := ""
		for _, part := range parts {
			if len(part) > 0 {
				result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
			}
		}
		return result
	}

	// Convert camelCase to PascalCase
	if len(s) > 0 {
		return strings.ToUpper(s[:1]) + s[1:]
	}

	return s
}
