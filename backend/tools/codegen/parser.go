package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
)

// Struct represents a parsed Go struct
type Struct struct {
	Name   string
	Fields []Field
}

// Field represents a struct field
type Field struct {
	Name         string
	Type         string
	JsonTag      string
	JsonOmit     bool // json:"-"
	IsOptional   bool // pointer or omitempty
	IsCollection bool // slice/array
	CollectionOf string
	IsRequired   bool // flutter:"required"
}

// Parser parses Go source files
type Parser struct {
	modelDir    string
	modulePath  string
	typeAliases map[string]string // maps type name to underlying type (e.g., FactArea -> string)
}

// NewParser creates a new parser
func NewParser(modelDir, modulePath string) *Parser {
	return &Parser{
		modelDir:    modelDir,
		modulePath:  modulePath,
		typeAliases: make(map[string]string),
	}
}

// ParsePackage parses all Go files in the model directory
func (p *Parser) ParsePackage() ([]Struct, error) {
	fset := token.NewFileSet()

	// Parse all Go files in the directory
	pkgs, err := parser.ParseDir(fset, p.modelDir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	// First pass: collect type aliases (e.g., type FactArea string)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					// check if it's a simple type alias (not a struct)
					if ident, ok := typeSpec.Type.(*ast.Ident); ok {
						p.typeAliases[typeSpec.Name.Name] = ident.Name
					}
				}
			}
		}
	}

	// Second pass: parse structs
	var structs []Struct

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					s := p.parseStruct(typeSpec.Name.Name, structType)
					structs = append(structs, s)
				}
			}
		}
	}

	return structs, nil
}

// parseStruct parses a struct type
func (p *Parser) parseStruct(name string, structType *ast.StructType) Struct {
	s := Struct{
		Name:   name,
		Fields: []Field{},
	}

	for _, field := range structType.Fields.List {
		// Skip embedded fields without names
		if len(field.Names) == 0 {
			continue
		}

		f := p.parseField(field)
		if !f.JsonOmit {
			s.Fields = append(s.Fields, f)
		}
	}

	return s
}

// parseField parses a struct field
func (p *Parser) parseField(field *ast.Field) Field {
	f := Field{
		Name: field.Names[0].Name,
	}

	// Parse type
	f.Type, f.IsOptional, f.IsCollection, f.CollectionOf = p.parseType(field.Type)

	// Parse tags
	if field.Tag != nil {
		tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))

		// Parse json tag
		jsonTag := tag.Get("json")
		if jsonTag == "-" {
			f.JsonOmit = true
			return f
		}

		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			f.JsonTag = parts[0]

			// Check for omitempty
			for _, part := range parts[1:] {
				if part == "omitempty" {
					f.IsOptional = true
				}
			}
		}

		// Parse flutter tag
		if tag.Get("flutter") == "required" {
			f.IsRequired = true
		}
	}

	// Default json tag is lowercase field name
	if f.JsonTag == "" {
		f.JsonTag = toSnakeCase(f.Name)
	}

	return f
}

// parseType parses a field type expression
func (p *Parser) parseType(expr ast.Expr) (typeName string, isOptional bool, isCollection bool, collectionOf string) {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple type: string, int, or custom type alias
		name := t.Name
		// resolve type alias if exists
		if underlying, ok := p.typeAliases[name]; ok {
			return underlying, false, false, ""
		}
		return name, false, false, ""

	case *ast.StarExpr:
		// Pointer type: *Type
		innerType, _, innerIsCollection, innerCollectionOf := p.parseType(t.X)
		return innerType, true, innerIsCollection, innerCollectionOf

	case *ast.ArrayType:
		// Array/slice type: []Type
		elemType, _, _, _ := p.parseType(t.Elt)
		return "[]" + elemType, false, true, elemType

	case *ast.SelectorExpr:
		// Qualified type: pkg.Type
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name, false, false, ""
		}
		return t.Sel.Name, false, false, ""

	case *ast.MapType:
		// Map type: map[K]V
		// Check if it's map[string]string specifically
		if keyIdent, ok := t.Key.(*ast.Ident); ok && keyIdent.Name == "string" {
			if valIdent, ok := t.Value.(*ast.Ident); ok && valIdent.Name == "string" {
				return "MapStringString", false, false, ""
			}
		}
		return "Map", false, false, ""

	case *ast.IndexExpr:
		// Generic type: pkg.Type[T] (e.g., datatypes.JSONType[TrainingReasoning])
		// Extract the type argument and use it as the type
		if sel, ok := t.X.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok {
				if pkg.Name == "datatypes" && strings.HasPrefix(sel.Sel.Name, "JSONType") {
					// datatypes.JSONType[T] -> T
					innerType, isOpt, isColl, collOf := p.parseType(t.Index)
					return innerType, isOpt, isColl, collOf
				}
			}
		}
		return "unknown", false, false, ""

	case *ast.InterfaceType:
		// Interface type
		return "interface{}", false, false, ""

	default:
		return "unknown", false, false, ""
	}
}
