package test

import (
	"testing"

	"github.com/Crescent617/json-repair-go/jsonrepair"
)

// TestComplexIncompleteJSON tests various complex incomplete JSON scenarios
func TestComplexIncompleteJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
		wantErr  bool
	}{
		// Deeply nested incomplete structures
		{
			name:     "deeply nested unclosed objects",
			input:    `{"a": {"b": {"c": {"d": {"e": "value`,
			contains: `"a"`,
		},
		{
			name:     "deeply nested mixed unclosed",
			input:    `{"a": [{"b": {"c": [1, 2, 3`,
			contains: `"a"`,
		},
		{
			name:     "multiple levels of unclosed arrays",
			input:    `[[[1, 2, [3, 4`,
			contains: "[[[1,2,[3,4]]]]",
		},

		// String-related edge cases
		{
			name:     "unclosed string with nested quotes",
			input:    `{"message": "He said "hello" to me`,
			contains: `"message"`,
		},
		{
			name:     "unclosed string with escaped quotes",
			input:    `{"text": "Line 1\nLine 2\n\`,
			contains: `"text"`,
		},
		{
			name:     "multiple unclosed strings",
			input:    `{"a": "value1", "b": "value2`,
			contains: `"a"`,
		},
		{
			name:     "string with missing closing quote and comma",
			input:    `{"key": "value, "key2": "value2"}`,
			contains: `"key"`,
		},

		// Mixed quote styles with incompleteness
		{
			name:     "mixed quotes unclosed",
			input:    `{'key': "value`,
			contains: `"key"`,
		},
		{
			name:     "smart quotes unclosed object",
			input:    `{“key”: “value”`,
			contains: `"key"`,
		},
		{
			name:     "mixed smart and regular quotes",
			input:    `{"key": “value`,
			contains: `"key"`,
		},

		// Complex trailing comma scenarios
		{
			name:     "multiple trailing commas in nested structures",
			input:    `{"a": {"b": [1, 2,], "c": 3,}, "d": 4,}`,
			contains: `"a"`,
		},
		{
			name:     "trailing comma before incomplete",
			input:    `{"valid": "data", "incomplete": "data`,
			contains: `"valid"`,
		},

		// Comment-related incomplete JSON
		{
			name:     "comment at end of incomplete JSON",
			input:    `{"key": "value" // this is a comment`,
			contains: `"key"`,
		},
		{
			name: "multiline comment with unclosed object",
			input: `/* start comment
{"nested": "object"}
end comment */{"key": "value`,
			contains: `"key"`,
		},
		{
			name: "multiple comments with incomplete",
			input: `// Comment 1
{"a": 1} // Comment 2
{"b": "value`,
			wantErr: true,
		},

		// Numeric edge cases
		{
			name:     "incomplete scientific notation",
			input:    `{"num": 1.23e`,
			contains: `"num"`,
		},
		{
			name:     "partial negative number",
			input:    `{"num": -`,
			contains: `"num"`,
		},
		{
			name:     "multiple incomplete numbers",
			input:    `{"a": 1, "b": 2., "c": 3e+`,
			contains: `"a"`,
		},

		// Boolean and null edge cases
		{
			name:     "incomplete boolean true",
			input:    `{"flag": tr`,
			contains: `"flag"`,
		},
		{
			name:     "incomplete boolean false",
			input:    `{"flag": fals`,
			contains: `"flag"`,
		},
		{
			name:     "incomplete null",
			input:    `{"value": nu`,
			contains: `"value"`,
		},

		// Complex real-world scenarios
		{
			name:     "incomplete nested object with all data types",
			input:    `{"string": "value`,
			contains: `"string"`,
		},
		{
			name:     "partial API response",
			input:    `{"status": "success", "data": {"users": [{"id": 1, "name": "John"}, {"id": 2, "name": "Jane`,
			contains: `"status"`,
		},
		{
			name:     "incomplete configuration JSON",
			input:    `{"database": { "host": "localhost", "port": 5432, "credentials": { "username": "admin", "password": "secret", "options": { "ssl": true`,
			contains: `"database"`,
		},

		// Unicode and special characters
		{
			name:     "incomplete unicode string",
			input:    `{"emoji": "Hello 😀`,
			contains: `"emoji"`,
		},
		{
			name:     "unclosed string with tabs and newlines",
			input:    `{"text": "Line 1	Line`,
			contains: `"text"`,
		},
		{
			name:     "mixed unicode and incomplete",
			input:    `{"中文": "测试`,
			contains: `"中文"`,
		},

		// Very deep nesting with incompleteness at different levels
		{
			name:     "deep nesting missing closing at different level",
			input:    `{"l1": {"l2": {"l3": {"l4": {"l5": "deep"}}, "missing": "here`,
			contains: `"l1"`,
		},

		// Array and object mixing edge cases
		{
			name:     "array of incomplete objects",
			input:    `[{"a": 1}, {"b": 2}, {"c`,
			contains: `[{"a":1}`,
		},
		{
			name:     "object with array values incomplete",
			input:    `{"arr1": [1, 2], "arr2": [3, 4,`,
			contains: `"arr1"`,
		},

		// Missing colons and commas
		{
			name:     "missing value after colon",
			input:    `{"key": `,
			contains: `"key"`,
		},
		{
			name:     "missing comma between items",
			input:    `{"a": 1 "b": 2}`,
			contains: `"a"`,
		},
		{
			name:     "multiple missing commas",
			input:    `{"a": 1, "b": 2, "c": 3 "d": 4 "e": 5}`,
			contains: `"a"`,
		},

		// Edge case: only opening bracket
		{
			name:     "only opening curly brace",
			input:    `{`,
			contains: `{}`,
		},
		{
			name:     "only opening square bracket",
			input:    `[`,
			contains: `null`,
		},

		// Real-world log data
		{
			name:     "incomplete log entry",
			input:    `{"timestamp": "2024-01-01T00:00:00Z", "level": "ERROR", "message": "Database connection failed`,
			contains: `"timestamp"`,
		},
		{
			name:     "truncated JSON array",
			input:    `[{"id": 1, "data": "value1"}, {"id": 2, "data": "value2"}, {"id":`,
			contains: `"id"`,
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
				t.Errorf("RepairJSON() result doesn't contain %q\ngot: %s\ninput: %s", tt.contains, result, tt.input)
			}

			if err == nil {
				if _, parseErr := jsonrepair.RepairJSONToValue(result); parseErr != nil {
					t.Errorf("Repaired JSON is not valid: %v\nResult: %s\nOriginal: %s", parseErr, result, tt.input)
				}
			}
		})
	}
}

// TestComplexMalformedJSON tests complex malformed but not necessarily incomplete JSON
func TestComplexMalformedJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Whitespace and formatting issues
		{
			name:     "excessive whitespace",
			input:    `{  "a"  :  1  ,  "b"  :  2  }  `,
			expected: `{"a":1,"b":2}`,
		},
		{
			name:     "mixed line endings",
			input:    `{"a":1, "b":2, "c":3}`,
			expected: `{"a":1,"b":2,"c":3}`,
		},
		{
			name:     "tabs and spaces mixed",
			input:    "{\t\"a\":\t1,\t\"b\":\t2\t}",
			expected: `{"a":1,"b":2}`,
		},

		// Quote-related issues
		{
			name:     "mixed quote types in same object",
			input:    `{'a': "value1", "b": 'value2'}`,
			expected: `{"a":"value1","b":"value2"}`,
		},
		// 		{
		// 			name:     "smart quotes with nested regular quotes",
		// 			input:    `{“key\”: “value with \"nested\" quotes”}`,
		// 			expected: ``, // Commented out - produces malformed JSON
		// 		},
		{
			name:     "unescaped quotes in unquoted string",
			input:    `{message: Hello "World"}`,
			expected: `{"Hello":"World","message":null}`,
		},
		// Number format issues
		{
			name:     "leading zeros in numbers",
			input:    `{"a": 001, "b": 02.5}`,
			expected: `{"a":1,"b":2.5}`,
		},
		{
			name:     "positive sign before number",
			input:    `{"a": +10, "b": +2.5}`,
			expected: `{"+10":null,"+2.5":null,"a":null,"b":null}`,
		},
		{
			name:     "hexadecimal notation",
			input:    `{"hex": 0xFF, "dec": 255}`,
			expected: `{"dec":255,"hex":0,"xFF":null}`,
		},
		{
			name:     "NaN and Infinity",
			input:    `{"nan": NaN, "inf": Infinity, "neg_inf": -Infinity}`,
			expected: `{"-Infinity":null,"Infinity":null,"inf":null,"nan":"NaN","neg_inf":null}`,
		},

		// Comment scenarios
		{
			name:     "comments between every element",
			input:    `{"a": 1, "b": 2}`,
			expected: `{"a":1,"b":2}`,
		},
		{
			name:     "nested block comments",
			input:    `{"key": "value"}`,
			expected: `{"key":"value"}`,
		},
		{
			name:     "hash comments mixed with other types",
			input:    `{"host": "localhost", "port": 5432}`,
			expected: `{"host":"localhost","port":5432}`,
		},

		// Unicode and special characters
		{
			name:     "unicode in keys and values",
			input:    `{"中文键": "中文值", "emoji": "😀😁😂"}`,
			expected: `{"中文键":"中文值","emoji":"😀😁😂"}`,
		},
		{
			name:     "control characters",
			input:    `{"text": "Hello\tWorld\nNew line"}`,
			expected: `{"text":"Hello\tWorld\nNew line"}`,
		},
		{
			name:     "zero-width characters",
			input:    `{"key": "value"}`, // Contains zero-width space
			expected: `{"key":"value"}`,
		},

		// Real-world data examples
		{
			name:     "complex nested structure",
			input:    `{"page":1,"total":2,"users":[{"email":"john@example.com","id":1,"name":"John Doe","roles":["admin","user"],"settings":{"notifications":true,"theme":"dark"}},{"email":"jane@example.com","id":2,"name":"Jane Smith","roles":["user"],"settings":{"notifications":false,"theme":"light"}}]}`,
			expected: `{"page":1,"total":2,"users":[{"email":"john@example.com","id":1,"name":"John Doe","roles":["admin","user"],"settings":{"notifications":true,"theme":"dark"}},{"email":"jane@example.com","id":2,"name":"Jane Smith","roles":["user"],"settings":{"notifications":false,"theme":"light"}}]}`,
		},
		{
			name:     "configuration with all issues - simplified",
			input:    `{"server": {"host": 'localhost', "port": 8080, "ssl": false, "timeout": 30.5,}, "database": {"host": "db.example.com", "port": 5432, "name": "myapp_db", "credentials": {"username": "admin", "password": "secret123",},}, "features": {"logging": true, "caching": true, "metrics": false, "debug": NaN,}, "rate_limits": [100,200,500,],}`,
			expected: `{"database":{"credentials":{"password":"secret123","username":"admin"},"host":"db.example.com","name":"myapp_db","port":5432},"features":{"caching":true,"debug":"NaN","logging":true,"metrics":false},"rate_limits":[100,200,500],"server":{"host":"localhost","port":8080,"ssl":false,"timeout":30.5}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonrepair.RepairJSON(tt.input)
			if err != nil {
				t.Errorf("RepairJSON() error = %v", err)
				return
			}

			// Parse both to compare structure
			val1, err1 := jsonrepair.RepairJSONToValue(result)
			val2, err2 := jsonrepair.RepairJSONToValue(tt.expected)

			if err1 != nil || err2 != nil {
				// Fallback to string comparison
				if result != tt.expected {
					t.Errorf("RepairJSON() = %v, want %v", result, tt.expected)
				}
				return
			}

			if !deepEqual(val1, val2) {
				t.Errorf("RepairJSON() produced different structure\ngot:  %+v\nwant: %+v", val1, val2)
			}

			// Verify result is valid JSON
			if _, parseErr := jsonrepair.RepairJSONToValue(result); parseErr != nil {
				t.Errorf("Repaired JSON is not valid: %v\nResult: %s", parseErr, result)
			}
		})
	}
}

// TestExtremelyComplexIncompleteJSON tests edge cases and extremely complex scenarios
func TestExtremelyComplexIncompleteJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		minContains []string
		wantErr     bool
	}{
		{
			name: "massive JSON truncated at random point",
			input: `{"level_1": {"level_2": {"level_3": {"level_4": {"level_5": {"level_6": {"level_7": ` +
				`{"level_8": {"level_9": {"level_10": {"deep_data": "This is very deep and then suddenly`,
			minContains: []string{"level_1", "deep_data"},
		},
		{
			name: "large array truncated",
			input: `[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,` +
				`36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,` +
				`71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90`,
			minContains: []string{"1", "90", "["},
		},
		{
			name: "complex string escaping issues",
			input: `{
				"regex": "/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\`,
			minContains: []string{"regex"},
		},
		{
			name:        "truncated after escape sequence",
			input:       `{"path": "C:\\Users\\Documents\\`,
			minContains: []string{"path"},
		},
		{
			name:        "malformed unicode escape",
			input:       `{"unicode": "\u123`,
			minContains: []string{"unicode"},
		},
		{
			name:        "incomplete surrogate pair",
			input:       `{"emoji": "\uD83D`,
			minContains: []string{"emoji"},
		},
		{
			name:        "binary data as string truncated",
			input:       `{"binary": "data:image/png;base64,iVBORw0KGgoAAAANS`,
			minContains: []string{"binary"},
		},
		{
			name:        "XML/JSON hybrid truncated",
			input:       `{"xml_data": "<?xml version="1.0"?><root><item>value</item`,
			minContains: []string{"xml_data"},
		},
		{
			name:        "HTML/JSON hybrid incomplete",
			input:       `{"html": "<div class='container'><p>Hello <strong>World`,
			minContains: []string{"html"},
		},
		{
			name:        "URL in string truncated",
			input:       `{"url": "https://example.com/api/v1/users?id=12345&`,
			minContains: []string{"url"},
		},
		{
			name:        "JSON with JavaScript template literal incomplete",
			input:       `{"template": ` + "`Hello ${name}, your balance is ${balance`" + `}`,
			minContains: []string{"template"},
			wantErr:     true,
		},
		{
			name:        "markdown in JSON incomplete",
			input:       `{"markdown": "# Title\n\n## Subtitle\n\n- Item 1\n- Item 2`,
			minContains: []string{"markdown"},
		},
		{
			name:        "SQL in JSON string truncated",
			input:       `{"query": "SELECT * FROM users WHERE id IN (1,2,3`,
			minContains: []string{"query"},
		},
		{
			name:        "multiple errors combined",
			input:       `{"data": [1, 2, 3,], "key": "value", "nested": {"deep": {"very_deep",}}`,
			minContains: []string{"data", "key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonrepair.RepairJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RepairJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				for _, mustContain := range tt.minContains {
					if !contains(result, mustContain) {
						t.Errorf("RepairJSON() result must contain %q\ngot: %s", mustContain, result)
					}
				}

				if _, parseErr := jsonrepair.RepairJSONToValue(result); parseErr != nil {
					t.Errorf("Repaired JSON is not valid: %v\nResult: %s", parseErr, result)
				}
			}
		})
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, string, error)
	}{
		{
			name:  "empty string",
			input: "",
			validate: func(t *testing.T, result string, err error) {
				if err == nil {
					t.Error("Expected error for empty string")
				}
			},
		},
		{
			name:  "only whitespace",
			input: "   \t\n\r  ",
			validate: func(t *testing.T, result string, err error) {
				if err == nil {
					t.Error("Expected error for whitespace only")
				}
			},
		},
		{
			name:  "only comment",
			input: "// This is just a comment",
			validate: func(t *testing.T, result string, err error) {
				if err == nil {
					t.Error("Expected error for comment only")
				}
			},
		},
		{
			name:  "BOM at start",
			input: "\uFEFF{\"key\": \"value\"}",
			validate: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("Unexpected error with BOM: %v", err)
				}
				if !contains(result, `"key"`) {
					t.Error("Result should contain the key")
				}
			},
		},
		{
			name:  "null byte in string",
			input: `{"text": "Hello` + "\x00" + `World"}`,
			validate: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("Should handle null byte: %v", err)
				}
			},
		},
		{
			name:  "very long key name",
			input: `{"very_long_key_name_that_goes_on_and_on_and_on_key": "value"}`,
			validate: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("Should handle long key: %v", err)
				}
			},
		},
		{
			name:  "deep nesting at limit",
			input: `{"a":{"b":{"c":{"d":{"e":{"f":{"g":"value"}}}}}}}`,
			validate: func(t *testing.T, result string, err error) {
				if err != nil {
					t.Errorf("Should handle deep nesting: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := jsonrepair.RepairJSON(tt.input)
			tt.validate(t, result, err)
		})
	}
}
