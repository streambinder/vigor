package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Generator generates Dart code from parsed structs
type Generator struct {
	outputDir       string
	modelImportPath string
	localTypes      map[string]bool
	tmpl            *template.Template
}

// NewGenerator creates a new generator
func NewGenerator(outputDir, modelImportPath string, localTypes map[string]bool) *Generator {
	g := &Generator{
		outputDir:       outputDir,
		modelImportPath: modelImportPath,
		localTypes:      localTypes,
	}

	tmpl := template.Must(template.New("dart").Funcs(template.FuncMap{
		"toDartType":          toDartType,
		"toSnakeCase":         toSnakeCase,
		"toCamelCase":         toCamelCase,
		"toClassName":         toClassName,
		"needsImport":         needsImport,
		"getImports":          g.getImports,
		"isNullable":          isNullable,
		"defaultValue":        defaultValue,
		"getRequiredFields":   getRequiredFields,
		"needsListDefault":    needsListDefault,
		"needsStringDefault":  needsStringDefault,
		"isDateTime":          isDateTime,
		"hasDateTime":         hasDateTime,
		"hasNullableDateTime": hasNullableDateTime,
	}).Parse(dartTemplate))

	g.tmpl = tmpl
	return g
}

// Generate generates a Dart file for a struct
func (g *Generator) Generate(s Struct) error {
	// Ensure output directory exists
	if err := os.MkdirAll(g.outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate file path
	fileName := toSnakeCase(s.Name) + ".dart"
	filePath := filepath.Join(g.outputDir, fileName)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := g.tmpl.Execute(file, s); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// toDartType converts a Go type to a Dart type
func toDartType(goType string, isOptional bool, isCollection bool, collectionOf string) string {
	var dartType string

	if isCollection {
		elemType := mapGoTypeToDart(collectionOf)
		dartType = fmt.Sprintf("List<%s>", elemType)
	} else {
		dartType = mapGoTypeToDart(goType)
	}

	if isOptional {
		return dartType + "?"
	}
	return dartType
}

// mapGoTypeToDart maps Go types to Dart types
func mapGoTypeToDart(goType string) string {
	// Handle qualified types (pkg.Type)
	if strings.Contains(goType, ".") {
		parts := strings.Split(goType, ".")
		pkg := parts[0]
		typeName := parts[1]

		switch pkg {
		case "uuid":
			return "String" // uuid.UUID -> String
		case "time":
			return "DateTime" // time.Time -> DateTime
		case "pq":
			if typeName == "StringArray" {
				return "List<String>"
			}
		case "datatypes":
			if typeName == "JSON" {
				return "Map<String, dynamic>"
			}
		case "gorm":
			if typeName == "DeletedAt" {
				return "DateTime" // gorm.DeletedAt -> DateTime
			}
		}

		// Default: assume it's a custom type, use the type name
		return toClassName(typeName)
	}

	// Handle simple types
	switch goType {
	case "string":
		return "String"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "int"
	case "float32", "float64":
		return "double"
	case "bool":
		return "bool"
	case "interface{}", "any":
		return "dynamic"
	case "Map":
		return "Map<String, dynamic>"
	default:
		// Assume it's a custom struct type
		return toClassName(goType)
	}
}

// needsImport checks if a field needs an import
func needsImport(field Field) bool {
	typeName := field.CollectionOf
	if typeName == "" {
		typeName = field.Type
	}

	// Check if it's a custom type (not a primitive)
	dartType := mapGoTypeToDart(typeName)
	primitives := []string{"String", "int", "double", "bool", "DateTime", "dynamic", "List", "Map"}

	for _, prim := range primitives {
		if strings.HasPrefix(dartType, prim) {
			return false
		}
	}

	return true
}

// getImports returns all necessary imports for a struct
func (g *Generator) getImports(s Struct) []string {
	var imports []string
	seen := make(map[string]bool)

	for _, field := range s.Fields {
		if needsImport(field) {
			typeName := field.CollectionOf
			if typeName == "" {
				typeName = field.Type
			}

			dartType := mapGoTypeToDart(typeName)
			fileName := toSnakeCase(dartType) + ".dart"

			// check if this type is local or needs external import
			if g.modelImportPath != "" && !g.localTypes[dartType] {
				fileName = g.modelImportPath + fileName
			}

			if !seen[fileName] {
				seen[fileName] = true
				imports = append(imports, fileName)
			}
		}
	}
	return imports
}

// isNullable checks if a field is nullable
func isNullable(field Field) bool {
	return field.IsOptional
}

// defaultValue returns the default value for a field in constructor
func defaultValue(field Field) string {
	if field.IsOptional {
		return ""
	}
	return "required "
}

// getRequiredFields returns a list of JSON field names that are marked as required
func getRequiredFields(s Struct) []string {
	var required []string
	for _, field := range s.Fields {
		if field.IsRequired {
			required = append(required, field.JsonTag)
		}
	}
	return required
}

// needsListDefault checks if a field needs an empty list default value
// Returns true for non-nullable collection fields to prevent null errors
func needsListDefault(field Field) bool {
	if field.IsOptional {
		return false
	}
	// Check if it's a direct collection ([]Type)
	if field.IsCollection {
		return true
	}
	// Check if the Dart type is a List type (e.g., pq.StringArray -> List<String>)
	dartType := toDartType(field.Type, false, field.IsCollection, field.CollectionOf)
	return strings.HasPrefix(dartType, "List<")
}

// needsStringDefault checks if a non-required String field needs a default empty value
func needsStringDefault(field Field) bool {
	if field.IsOptional || field.IsRequired || field.IsCollection {
		return false
	}
	dartType := toDartType(field.Type, false, field.IsCollection, field.CollectionOf)
	return dartType == "String"
}

// isDateTime checks if a field type maps to DateTime
func isDateTime(goType string) bool {
	return goType == "time.Time" || goType == "gorm.DeletedAt"
}

// hasDateTime checks if any field in the struct is a DateTime
func hasDateTime(s Struct) bool {
	for _, field := range s.Fields {
		if isDateTime(field.Type) {
			return true
		}
	}
	return false
}

// hasNullableDateTime checks if any field in the struct is a nullable DateTime
func hasNullableDateTime(s Struct) bool {
	for _, field := range s.Fields {
		if isDateTime(field.Type) && field.IsOptional {
			return true
		}
	}
	return false
}

// Dart template
const dartTemplate = `// GENERATED CODE - DO NOT MODIFY BY HAND
// Generated by vigor/codegen

import 'package:json_annotation/json_annotation.dart';
{{- range $import := getImports . }}
import '{{ $import }}';
{{- end }}

part '{{ toSnakeCase .Name }}.g.dart';

@JsonSerializable(explicitToJson: true)
class {{ toClassName .Name }} {
{{- $requiredFields := getRequiredFields . }}
{{- if $requiredFields }}
  // Fields marked as required in the backend model
  static const List<String> requiredFields = [
{{- range $i, $field := $requiredFields }}
{{- if $i }},{{ end }}
    '{{ $field }}'
{{- end }}
  ];
{{- end }}

{{- range .Fields }}
  @JsonKey(name: '{{ .JsonTag }}'{{- if needsListDefault . }}, defaultValue: []{{- else if needsStringDefault . }}, defaultValue: ''{{- end }}{{- if isDateTime .Type }}{{- if .IsOptional }}, toJson: _nullableDateTimeToJson{{- else }}, toJson: _dateTimeToJson{{- end }}{{- end }})
  final {{ toDartType .Type .IsOptional .IsCollection .CollectionOf }} {{ toCamelCase .Name }};
{{- end }}

  {{ toClassName .Name }}({
{{- range .Fields }}
    {{ defaultValue . }}this.{{ toCamelCase .Name }},
{{- end }}
  });

  factory {{ toClassName .Name }}.fromJson(Map<String, dynamic> json) => _${{ toClassName .Name }}FromJson(json);

  Map<String, dynamic> toJson() => _${{ toClassName .Name }}ToJson(this);
{{- if hasDateTime . }}

  static String _dateTimeToJson(DateTime dt) => dt.toUtc().toIso8601String();
{{- end }}
{{- if hasNullableDateTime . }}

  static String? _nullableDateTimeToJson(DateTime? dt) => dt?.toUtc().toIso8601String();
{{- end }}
}
`
