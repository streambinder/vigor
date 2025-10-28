package encoder

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test struct with various field types
type TestStruct struct {
	Name       string    `json:"name" prompt:"User's full name"`
	Age        int       `json:"age" prompt:"User's age in years"`
	Email      string    `json:"email" prompt:"User's email address"`
	IsActive   bool      `json:"is_active" prompt:"Whether user is active"`
	CreatedAt  time.Time `json:"created_at" prompt:"Creation timestamp"`
	unexported string    // should be ignored
	SkipMe     string    `json:"-"`                      // should be ignored due to json:"-"
	NoPrompt   string    `json:"no_prompt"`              // has no prompt tag
	SkipPrompt string    `json:"skip_prompt" prompt:"-"` // should be ignored due to prompt:"-"
}

type NestedStruct struct {
	ID      uuid.UUID  `json:"id" prompt:"Unique identifier"`
	Details TestStruct `json:"details"`
}

type SliceStruct struct {
	Tags   []string     `json:"tags" prompt:"List of tags"`
	Items  []TestStruct `json:"items"`
	Counts []int        `json:"counts" prompt:"Array of counts"`
}

type PointerStruct struct {
	Name    string      `json:"name" prompt:"Name field"`
	Details *TestStruct `json:"details"`
}

type ComplexStruct struct {
	ID            uuid.UUID    `json:"id" prompt:"Unique identifier"`
	SimpleField   string       `json:"simple" prompt:"A simple field"`
	NestedStruct  NestedStruct `json:"nested"`
	SliceOfBasic  []string     `json:"basic_slice" prompt:"Basic slice"`
	SliceOfStruct []TestStruct `json:"struct_slice"`
	PointerField  *TestStruct  `json:"pointer_field"`
}

func TestMarshalJSON_SimpleStruct(t *testing.T) {
	testObj := TestStruct{
		Name:       "John Doe",
		Age:        30,
		Email:      "john@example.com",
		IsActive:   true,
		CreatedAt:  time.Now(),
		unexported: "should not appear",
		SkipMe:     "should not appear",
		NoPrompt:   "value",
		SkipPrompt: "should not appear",
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check that prompt tags are used as values
	if val, ok := decoded["name"].(string); !ok || val != "string: User's full name" {
		t.Errorf("Expected name to be 'string: User's full name', got: %v", decoded["name"])
	}

	if val, ok := decoded["age"].(string); !ok || val != "int: User's age in years" {
		t.Errorf("Expected age to be 'int: User's age in years', got: %v", decoded["age"])
	}

	// Check that unexported fields are not included
	if _, exists := decoded["unexported"]; exists {
		t.Error("Unexported field should not be included")
	}

	// Check that json:"-" fields are not included
	if _, exists := decoded["SkipMe"]; exists {
		t.Error("Field with json:\"-\" should not be included")
	}

	// Check that fields without prompt tags still appear but with empty prompt
	if val, ok := decoded["no_prompt"].(string); !ok || val != "string: " {
		t.Errorf("Expected no_prompt to be 'string: ', got: %v", decoded["no_prompt"])
	}

	// Check that prompt:"-" fields are not included
	if _, exists := decoded["skip_prompt"]; exists {
		t.Error("Field with prompt:\"-\" should not be included")
	}
}

func TestMarshalJSON_NestedStruct(t *testing.T) {
	testObj := NestedStruct{
		ID: uuid.New(),
		Details: TestStruct{
			Name:     "Jane Doe",
			Age:      25,
			Email:    "jane@example.com",
			IsActive: false,
		},
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check ID field
	if val, ok := decoded["id"].(string); !ok || val != "uuid.UUID: Unique identifier" {
		t.Errorf("Expected id to be 'uuid.UUID: Unique identifier', got: %v", decoded["id"])
	}

	// Check nested struct
	if details, ok := decoded["details"].(map[string]any); ok {
		if val, ok := details["name"].(string); !ok || val != "string: User's full name" {
			t.Errorf("Expected nested name to be 'string: User's full name', got: %v", details["name"])
		}
	} else {
		t.Error("Expected details to be a nested object")
	}
}

func TestMarshalJSON_SliceFields(t *testing.T) {
	testObj := SliceStruct{
		Tags:   []string{"tag1", "tag2"},
		Items:  []TestStruct{{Name: "Test"}},
		Counts: []int{1, 2, 3},
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check tags with prompt
	if tags, ok := decoded["tags"].([]any); ok {
		if len(tags) != 1 || tags[0] != "[]string" {
			t.Errorf("Expected tags to be ['[]string'], got: %v", tags)
		}
	} else {
		t.Errorf("Expected tags to be an array, got: %v", decoded["tags"])
	}

	// Check items (slice of structs)
	if items, ok := decoded["items"].([]any); ok {
		if len(items) != 1 {
			t.Errorf("Expected items to have 1 element, got: %d", len(items))
		}
	} else {
		t.Error("Expected items to be an array")
	}

	// Check counts with prompt
	if counts, ok := decoded["counts"].([]any); ok {
		if len(counts) != 1 || counts[0] != "[]int" {
			t.Errorf("Expected counts to be ['[]int'], got: %v", counts)
		}
	} else {
		t.Error("Expected counts to be an array")
	}
}

func TestMarshalJSON_PointerField(t *testing.T) {
	details := &TestStruct{
		Name:  "Pointer Test",
		Age:   40,
		Email: "test@example.com",
	}

	testObj := PointerStruct{
		Name:    "Parent",
		Details: details,
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check that pointer field is processed
	if details, ok := decoded["details"].(map[string]any); ok {
		if val, ok := details["name"].(string); !ok || val != "string: User's full name" {
			t.Errorf("Expected pointer details name to be 'string: User's full name', got: %v", details["name"])
		}
	} else {
		t.Error("Expected details to be a nested object")
	}
}

func TestMarshalJSON_NilPointerField(t *testing.T) {
	testObj := PointerStruct{
		Name:    "Parent",
		Details: nil,
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Check that nil pointer field still generates structure
	if details, ok := decoded["details"].(map[string]any); ok {
		if val, ok := details["name"].(string); !ok || val != "string: User's full name" {
			t.Errorf("Expected nil pointer details name to be 'string: User's full name', got: %v", details["name"])
		}
	} else {
		t.Error("Expected details to be a nested object even when nil")
	}
}

func TestMarshalJSON_ComplexStruct(t *testing.T) {
	testObj := ComplexStruct{
		ID:          uuid.New(),
		SimpleField: "simple value",
		NestedStruct: NestedStruct{
			ID: uuid.New(),
			Details: TestStruct{
				Name: "Nested",
			},
		},
		SliceOfBasic:  []string{"a", "b"},
		SliceOfStruct: []TestStruct{{Name: "Item1"}},
		PointerField:  &TestStruct{Name: "Pointer"},
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify structure is deeply nested
	if _, ok := decoded["nested"].(map[string]any); !ok {
		t.Error("Expected nested to be an object")
	}

	if slice, ok := decoded["struct_slice"].([]any); !ok || len(slice) != 1 {
		t.Error("Expected struct_slice to be an array with 1 element")
	}
}

func TestBuildPromptMap_InvalidValue(t *testing.T) {
	result := buildPromptMap(reflect.Value{})
	if result != nil {
		t.Errorf("Expected nil for invalid value, got: %v", result)
	}
}

func TestBuildPromptMap_NonStructValue(t *testing.T) {
	// Test with a string value
	strValue := reflect.ValueOf("test string")
	result := buildPromptMap(strValue)
	if result != nil {
		t.Errorf("Expected nil for non-struct value, got: %v", result)
	}

	// Test with an int value
	intValue := reflect.ValueOf(42)
	result = buildPromptMap(intValue)
	if result != nil {
		t.Errorf("Expected nil for int value, got: %v", result)
	}
}

func TestBuildPromptMap_TimeField(t *testing.T) {
	type TimeStruct struct {
		CreatedAt time.Time `json:"created_at" prompt:"Creation time"`
	}

	testObj := TimeStruct{
		CreatedAt: time.Now(),
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// time.Time should be treated as a basic type
	if val, ok := decoded["created_at"].(string); !ok || val != "time.Time: Creation time" {
		t.Errorf("Expected created_at to be 'time.Time: Creation time', got: %v", decoded["created_at"])
	}
}

func TestBuildPromptMap_EmptySlice(t *testing.T) {
	type EmptySliceStruct struct {
		EmptyStrings []string     `json:"empty_strings"`
		EmptyStructs []TestStruct `json:"empty_structs"`
	}

	testObj := EmptySliceStruct{
		EmptyStrings: []string{},
		EmptyStructs: []TestStruct{},
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Empty slices should still have a structure
	if emptyStrings, ok := decoded["empty_strings"].([]any); !ok {
		t.Error("Expected empty_strings to be an array")
	} else if len(emptyStrings) != 1 {
		t.Errorf("Expected empty_strings to have 1 element, got: %d", len(emptyStrings))
	}

	if emptyStructs, ok := decoded["empty_structs"].([]any); !ok {
		t.Error("Expected empty_structs to be an array")
	} else if len(emptyStructs) != 1 {
		t.Errorf("Expected empty_structs to have 1 template element, got: %d", len(emptyStructs))
	}
}

func TestIndexComma(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"json,omitempty", 4},
		{"json", -1},
		{"", -1},
		{"a,b,c", 1},
		{",start", 0},
	}

	for _, test := range tests {
		result := indexComma(test.input)
		if result != test.expected {
			t.Errorf("indexComma(%q) = %d, expected %d", test.input, result, test.expected)
		}
	}
}

func TestParseJSONTag(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{
			name: "simple tag",
			field: reflect.StructField{
				Name: "Field",
				Tag:  `json:"field_name"`,
			},
			expected: "field_name",
		},
		{
			name: "tag with options",
			field: reflect.StructField{
				Name: "Field",
				Tag:  `json:"field_name,omitempty"`,
			},
			expected: "field_name",
		},
		{
			name: "dash tag",
			field: reflect.StructField{
				Name: "Field",
				Tag:  `json:"-"`,
			},
			expected: "",
		},
		{
			name: "no tag",
			field: reflect.StructField{
				Name: "Field",
				Tag:  "",
			},
			expected: "Field",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseJSONTag(test.field)
			if result != test.expected {
				t.Errorf("parseJSONTag() = %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestBuildSliceFieldPromptValue_PromptDash(t *testing.T) {
	field := reflect.StructField{
		Name: "Field",
		Tag:  `json:"field" prompt:"-"`,
		Type: reflect.TypeOf([]string{}),
	}

	result := buildSliceFieldPromptValue(field, field.Type)
	if slice, ok := result.([]any); !ok || len(slice) != 0 {
		t.Errorf("Expected empty slice for prompt:\"-\", got: %v", result)
	}
}

func TestBuildSliceFieldPromptValue_WithPrompt(t *testing.T) {
	field := reflect.StructField{
		Name: "Field",
		Tag:  `json:"field" prompt:"Custom prompt"`,
		Type: reflect.TypeOf([]string{}),
	}

	result := buildSliceFieldPromptValue(field, field.Type)
	if slice, ok := result.([]any); !ok || len(slice) != 1 || slice[0] != "[]string" {
		t.Errorf("Expected ['[]string'], got: %v", result)
	}
}

func TestBuildSliceFieldPromptValue_NoPrompt(t *testing.T) {
	field := reflect.StructField{
		Name: "Field",
		Tag:  `json:"field"`,
		Type: reflect.TypeOf([]TestStruct{}),
	}

	result := buildSliceFieldPromptValue(field, field.Type)
	if slice, ok := result.([]any); !ok || len(slice) != 1 {
		t.Errorf("Expected slice with 1 template element, got: %v", result)
	}
}

func TestBuildPromptMap_Pointer(t *testing.T) {
	testObj := &TestStruct{
		Name:  "Test",
		Age:   30,
		Email: "test@example.com",
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify pointer is dereferenced and processed
	if val, ok := decoded["name"].(string); !ok || val != "string: User's full name" {
		t.Errorf("Expected name to be 'string: User's full name', got: %v", decoded["name"])
	}
}

func TestBuildPromptMap_NilPointer(t *testing.T) {
	var testObj *TestStruct

	result := buildPromptMap(reflect.ValueOf(testObj))
	if result != nil {
		t.Errorf("Expected nil for nil pointer, got: %v", result)
	}
}

func TestBuildPromptMap_SliceOfPointers(t *testing.T) {
	type SliceOfPointers struct {
		Items []*TestStruct `json:"items"`
	}

	testObj := SliceOfPointers{
		Items: []*TestStruct{
			{Name: "Item1"},
		},
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Verify slice of pointers is processed
	if items, ok := decoded["items"].([]any); !ok || len(items) != 1 {
		t.Error("Expected items to be an array with 1 element")
	}
}

func TestBuildFieldPromptValue_PointerToNonStruct(t *testing.T) {
	type PointerToInt struct {
		Value *int `json:"value" prompt:"Integer value"`
	}

	val := 42
	testObj := PointerToInt{
		Value: &val,
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Pointer to non-struct types should use prompt
	if val, ok := decoded["value"].(string); !ok || val != "*int: Integer value" {
		t.Errorf("Expected value to be '*int: Integer value', got: %v", decoded["value"])
	}
}

func TestBuildSlicePromptMap_NonStructElements(t *testing.T) {
	type SliceOfInts struct {
		Numbers []int `json:"numbers"`
	}

	testObj := SliceOfInts{
		Numbers: []int{1, 2, 3},
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Non-string, non-struct slices should return array with nil
	if numbers, ok := decoded["numbers"].([]any); !ok || len(numbers) != 1 {
		t.Errorf("Expected numbers to have 1 nil element, got: %v", numbers)
	}
}

func TestBuildFieldPromptValue_SliceOfStructsWithPrompt(t *testing.T) {
	type SliceWithPrompt struct {
		Items []TestStruct `json:"items" prompt:"List of items"`
	}

	testObj := SliceWithPrompt{
		Items: []TestStruct{{Name: "Test"}},
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Slice with prompt tag should use the type
	if items, ok := decoded["items"].([]any); !ok {
		t.Error("Expected items to be an array")
	} else if len(items) != 1 || items[0] != "[]encoder.TestStruct" {
		t.Errorf("Expected items to be ['[]encoder.TestStruct'], got: %v", items)
	}
}

func TestMarshalJSON_NilValue(t *testing.T) {
	encoder := JSONWithPrompts{Value: nil}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if string(result) != "null" {
		t.Errorf("Expected 'null', got: %s", string(result))
	}
}

func TestJSONTag_EmptyString(t *testing.T) {
	field := reflect.StructField{
		Name: "Field",
		Tag:  `json:""`,
	}

	result := parseJSONTag(field)
	// Empty string should return the field name
	if result != "Field" {
		t.Errorf("Expected 'Field', got: %q", result)
	}
}

func TestBuildPromptMap_PointerToPointer(t *testing.T) {
	type PointerToPointer struct {
		Value **string `json:"value" prompt:"String pointer"`
	}

	str := "test"
	ptr := &str
	testObj := PointerToPointer{
		Value: &ptr,
	}

	encoder := JSONWithPrompts{Value: testObj}
	result, err := json.Marshal(encoder)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// Should use prompt for pointer types
	if val, ok := decoded["value"].(string); !ok || val != "**string: String pointer" {
		t.Errorf("Expected value to be '**string: String pointer', got: %v", decoded["value"])
	}
}

// Direct tests for buildSlicePromptMap via buildPromptMap with slice values
func TestBuildPromptMap_DirectSlice_Empty(t *testing.T) {
	emptySlice := []TestStruct{}
	result := buildPromptMap(reflect.ValueOf(emptySlice))

	if slice, ok := result.([]any); !ok || len(slice) != 0 {
		t.Errorf("Expected empty slice, got: %v", result)
	}
}

func TestBuildPromptMap_DirectSlice_Structs(t *testing.T) {
	structSlice := []TestStruct{
		{Name: "Test1"},
		{Name: "Test2"},
	}
	result := buildPromptMap(reflect.ValueOf(structSlice))

	if slice, ok := result.([]any); !ok || len(slice) != 1 {
		t.Errorf("Expected slice with 1 template element, got: %v", result)
	} else {
		if _, ok := slice[0].(map[string]any); !ok {
			t.Error("Expected slice element to be a map[string]any (struct template)")
		}
	}
}

func TestBuildPromptMap_DirectSlice_Strings(t *testing.T) {
	stringSlice := []string{"a", "b", "c"}
	result := buildPromptMap(reflect.ValueOf(stringSlice))

	if slice, ok := result.([]string); !ok || len(slice) != 1 || slice[0] != "<string>" {
		t.Errorf("Expected [\"<string>\"], got: %v", result)
	}
}

func TestBuildPromptMap_DirectSlice_Pointers(t *testing.T) {
	item1 := TestStruct{Name: "Test1"}
	item2 := TestStruct{Name: "Test2"}
	pointerSlice := []*TestStruct{&item1, &item2}
	result := buildPromptMap(reflect.ValueOf(pointerSlice))

	if slice, ok := result.([]any); !ok || len(slice) != 1 {
		t.Errorf("Expected slice with 1 template element, got: %v", result)
	} else {
		if _, ok := slice[0].(map[string]any); !ok {
			t.Error("Expected slice element to be a map[string]any (struct template)")
		}
	}
}

func TestBuildPromptMap_DirectSlice_Ints(t *testing.T) {
	intSlice := []int{1, 2, 3}
	result := buildPromptMap(reflect.ValueOf(intSlice))

	if slice, ok := result.([]any); !ok || len(slice) != 0 {
		t.Errorf("Expected empty slice for non-string, non-struct elements, got: %v", result)
	}
}

func TestBuildPromptMap_DirectSlice_EmptyStrings(t *testing.T) {
	stringSlice := []string{}
	result := buildPromptMap(reflect.ValueOf(stringSlice))

	if slice, ok := result.([]any); !ok || len(slice) != 0 {
		t.Errorf("Expected empty slice, got: %v", result)
	}
}
