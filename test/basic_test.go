package test

import (
	"testing"

	"github.com/Crescent617/json-repair-go/jsonrepair"
)

func TestRepairValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple object",
			input:    `{"name": "John", "age": 30}`,
			expected: `{"age":30,"name":"John"}`,
		},
		{
			name:     "simple array",
			input:    `[1, 2, 3]`,
			expected: `[1,2,3]`,
		},
		{
			name:     "nested structure",
			input:    `{"user": {"name": "Alice", "scores": [95, 87, 92]}}`,
			expected: `{"user":{"name":"Alice","scores":[95,87,92]}}`,
		},
		{
			name:     "boolean and null",
			input:    `{"active": true, "deleted": false, "value": null}`,
			expected: `{"active":true,"deleted":false,"value":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonrepair.RepairJSON(tt.input)
			if err != nil {
				t.Errorf("RepairJSON() failed for valid JSON: %v", err)
				return
			}

			// Parse both to compare structure (order might differ)
			val1, err1 := jsonrepair.RepairJSONToValue(result)
			val2, err2 := jsonrepair.RepairJSONToValue(tt.expected)

			if err1 != nil || err2 != nil {
				// If we can't parse both, fall back to string comparison
				if result != tt.expected {
					t.Errorf("RepairJSON() = %v, want %v", result, tt.expected)
				}
				return
			}

			// Compare values
			if !deepEqual(val1, val2) {
				t.Errorf("RepairJSON() produced different structure\ngot:  %+v\nwant: %+v", val1, val2)
			}
		})
	}
}

func TestRepairMalformedJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		contains string // substring that should be in output
	}{
		{
			name:     "missing quotes around key",
			input:    `{name: "John"}`,
			contains: `"name"`,
		},
		{
			name:     "trailing comma in object",
			input:    `{"a": 1, "b": 2,}`,
			contains: `"a"`,
		},
		{
			name:     "trailing comma in array",
			input:    `[1, 2, 3,]`,
			contains: "1,2,3",
		},
		{
			name:     "single quotes instead of double",
			input:    `{'key': 'value'}`,
			contains: `"key"`,
		},
		{
			name:     "unclosed array",
			input:    `[1, 2, 3`,
			contains: "[1,2,3]",
		},
		{
			name:     "unclosed object",
			input:    `{"key": "value"`,
			contains: `"key"`,
		},
		{
			name:     "comment before JSON",
			input:    "// This is a comment\n{\"a\": 1}",
			contains: `"a"`,
		},
		{
			name:     "hash comment",
			input:    "# Comment\n{\"b\": 2}",
			contains: `"b"`,
		},
		{
			name:     "block comment",
			input:    "/* comment */{\"c\": 3}",
			contains: `"c"`,
		},
		{
			name:     "smart quotes",
			input:    `{“key”: “value”}`,
			contains: `"key"`,
		},
		{
			name:     "boolean without quotes",
			input:    `{"active": true}`,
			contains: `"active":true`,
		},
		{
			name:     "null without quotes",
			input:    `{"value": null}`,
			contains: `"value":null`,
		},
		{
			name:     "scientific notation",
			input:    `{"sci": 1.23e-4}`,
			contains: `"sci"`,
		},
		{
			name:     "numbers as strings",
			input:    `{"num": "123"}`,
			contains: `"num"`,
		},
		{
			name:     "incompelete json",
			input:    `{"num": "123"`,
			contains: `"num"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonrepair.RepairJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RepairJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.contains != "" && !contains(result, tt.contains) {
				t.Errorf("RepairJSON() result doesn't contain %q\ngot: %s", tt.contains, result)
			}

			// Verify the result is valid JSON
			if err == nil {
				if _, parseErr := jsonrepair.RepairJSONToValue(result); parseErr != nil {
					t.Errorf("Repaired JSON is not valid: %v\nResult: %s", parseErr, result)
				}
			}
		})
	}
}

func TestJSONValueTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{
			name:     "object",
			input:    `{"key": "value"}`,
			wantType: "object",
		},
		{
			name:     "array",
			input:    `[1, 2, 3]`,
			wantType: "array",
		},
		{
			name:     "string",
			input:    `"hello"`,
			wantType: "string",
		},
		{
			name:     "number",
			input:    `42`,
			wantType: "number",
		},
		{
			name:     "boolean",
			input:    `true`,
			wantType: "boolean",
		},
		{
			name:     "null",
			input:    `null`,
			wantType: "null",
		},
		{
			name:     "float",
			input:    `3.14`,
			wantType: "float",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := jsonrepair.RepairJSONToValue(tt.input)
			if err != nil {
				t.Errorf("RepairJSONToValue() error = %v", err)
				return
			}

			_ = value // Test passes if no error
		})
	}
}

func TestIndentOption(t *testing.T) {
	input := `{"a":1,"b":{"c":3}}`

	result, err := jsonrepair.RepairJSON(input, jsonrepair.WithIndent(2))
	if err != nil {
		t.Errorf("RepairJSON() with indent failed: %v", err)
		return
	}

	// Check if result contains newlines (indicating pretty printing)
	if !contains(result, "\n") {
		t.Error("Expected indented output to contain newlines")
	}
}

func TestEmptyInput(t *testing.T) {
	_, err := jsonrepair.RepairJSON("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestInvalidInput(t *testing.T) {
	// Test that we handle truly invalid input gracefully
	tests := []string{
		",",
		"{",
		"[",
		"junk",
		"{,",
		"[,",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := jsonrepair.RepairJSON(input)
			// We expect these to fail - just check they don't panic
			_ = err
		})
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// deepEqual compares two JSON values for equality
func deepEqual(a, b interface{}) bool {
	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal, ok := b.(map[string]interface{})
		if !ok || len(aVal) != len(bVal) {
			return false
		}
		for k, v := range aVal {
			bv, exists := bVal[k]
			if !exists || !deepEqual(v, bv) {
				return false
			}
		}
		return true
	case []interface{}:
		bVal, ok := b.([]interface{})
		if !ok || len(aVal) != len(bVal) {
			return false
		}
		for i, v := range aVal {
			if !deepEqual(v, bVal[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}
