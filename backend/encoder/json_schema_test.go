package encoder

import (
	"reflect"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type inner struct {
	Name string `json:"name" prompt:"the name"`
	Age  int    `json:"age,omitempty"`
}

type sample struct {
	Title     string                    `json:"title" prompt:"the title;enum: alpha , beta ,gamma"`
	Count     int                       `json:"count"`
	Big       uint64                    `json:"big"`
	Ratio     float64                   `json:"ratio"`
	Active    bool                      `json:"active"`
	Tags      []string                  `json:"tags"`
	Nums      []int                     `json:"nums"`
	Items     []inner                   `json:"items"`
	PtrItems  []*inner                  `json:"ptr_items"`
	Labels    map[string]string         `json:"labels"`
	AnyMap    map[string]int            `json:"any_map"`
	Nested    inner                     `json:"nested"`
	Ptr       *inner                    `json:"ptr"`
	When      time.Time                 `json:"when"`
	Deleted   gorm.DeletedAt            `json:"deleted"`
	Raw       datatypes.JSONType[inner] `json:"raw" prompt:"raw payload"`
	Chan      chan int                  `json:"chan"`
	Skipped   string                    `json:"-"`
	NoTag     string
	Hidden    string `json:"hidden" prompt:"-"`
	EmptyName string `json:",omitempty"`
	ignored   string `json:"ignored"`
}

func TestJSONSchemaSample(t *testing.T) {
	schema := JSONSchema(sample{})
	if schema == nil || schema["type"] != "object" {
		t.Fatalf("schema = %v", schema)
	}
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("properties type %T", schema["properties"])
	}
	for _, key := range []string{"title", "count", "big", "ratio", "active", "tags", "nums", "items", "ptr_items", "labels", "any_map", "nested", "ptr", "raw"} {
		if _, ok := props[key]; !ok {
			t.Errorf("missing property %q in %v", key, props)
		}
	}
	for _, key := range []string{"when", "deleted", "chan", "hidden", "ignored"} {
		if _, ok := props[key]; ok {
			t.Errorf("property %q must be skipped", key)
		}
	}
	title := props["title"].(map[string]interface{})
	if title["type"] != "string" || title["description"] != "the title" {
		t.Errorf("title = %v", title)
	}
	if enum, ok := title["enum"].([]interface{}); !ok || len(enum) != 3 || enum[0] != "alpha" {
		t.Errorf("title enum = %v", title["enum"])
	}
	if props["tags"].(map[string]interface{})["items"].(map[string]interface{})["type"] != "string" {
		t.Errorf("tags = %v", props["tags"])
	}
	if props["labels"].(map[string]interface{})["additionalProperties"] == nil {
		t.Errorf("labels = %v", props["labels"])
	}
	required, ok := schema["required"].([]string)
	if !ok || len(required) == 0 {
		t.Fatalf("required = %v", schema["required"])
	}
	for _, r := range required {
		if r == "age" {
			t.Error("omitempty field must not be required")
		}
	}
}

func TestJSONSchemaPointerAndNonStruct(t *testing.T) {
	if got := JSONSchema(&sample{}); got == nil {
		t.Fatal("pointer schema is nil")
	}
	if got := JSONSchema(42); got != nil {
		t.Errorf("non-struct schema = %v, want nil", got)
	}
	if got := jsonSchemaForType(reflect.TypeOf("x")); got != nil {
		t.Errorf("string schema = %v, want nil", got)
	}
	type empty struct{}
	if got := JSONSchema(empty{}); got == nil {
		t.Fatal("empty struct schema is nil")
	} else if _, ok := got["required"]; ok {
		t.Errorf("empty struct must have no required key: %v", got)
	}
}

func TestParseJSONTagForSchema(t *testing.T) {
	cases := []struct {
		tag  string
		name string
		omit bool
	}{
		{"name", "name", false},
		{"name,omitempty", "name", true},
		{"name,string,omitempty", "name", true},
		{"-", "", false},
		{",omitempty", "", true},
	}
	for _, tc := range cases {
		name, omit := parseJSONTagForSchema(tc.tag)
		if name != tc.name || omit != tc.omit {
			t.Errorf("parseJSONTagForSchema(%q) = %q,%v want %q,%v", tc.tag, name, omit, tc.name, tc.omit)
		}
	}
}

func TestParsePromptTag(t *testing.T) {
	desc, enum := parsePromptTag("hello;enum:a,b;other:x")
	if desc != "hello" || len(enum) != 2 {
		t.Errorf("got %q %v", desc, enum)
	}
	desc, enum = parsePromptTag("")
	if desc != "" || enum != nil {
		t.Errorf("empty got %q %v", desc, enum)
	}
	desc, enum = parsePromptTag("+")
	if desc != "+" || enum != nil {
		t.Errorf("plus got %q %v", desc, enum)
	}
}

func TestBuildPropertySchemaKinds(t *testing.T) {
	if got := buildPropertySchema(reflect.TypeOf(complex64(0)), "d", nil); got != nil {
		t.Errorf("complex = %v, want nil", got)
	}
	ptr := buildPropertySchema(reflect.TypeOf((*string)(nil)), "desc", nil)
	if ptr["type"] != "string" || ptr["description"] != "desc" {
		t.Errorf("ptr string = %v", ptr)
	}
	arr := buildPropertySchema(reflect.TypeOf([2]int{}), "", nil)
	if arr["type"] != "array" {
		t.Errorf("array = %v", arr)
	}
	plus := buildPropertySchema(reflect.TypeOf(""), "+", []string{"x"})
	if _, ok := plus["description"]; ok {
		t.Errorf("plus description must be omitted: %v", plus)
	}
	floats := buildPropertySchema(reflect.TypeOf(float32(0)), "", nil)
	if floats["type"] != "number" {
		t.Errorf("float = %v", floats)
	}
}
