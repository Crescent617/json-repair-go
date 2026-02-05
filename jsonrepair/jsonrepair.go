package jsonrepair

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RepairJSON repairs and returns a valid JSON string from potentially malformed input
func RepairJSON(input string, opts ...ParserOption) (string, error) {
	value, err := RepairJSONToValue(input, opts...)
	if err != nil {
		return "", err
	}
	return marshalJSON(value, opts...)
}

// RepairJSONToValue repairs JSON and returns it as a Go value (map, slice, string, number, bool, or nil)
func RepairJSONToValue(input string, opts ...ParserOption) (JSONValue, error) {
	// Check for empty input
	if strings.TrimSpace(input) == "" {
		return nil, errors.New("empty input")
	}

	// First try to parse with standard JSON
	options := mergeOptions(opts...)
	if !options.SkipJSONLoads {
		var result JSONValue
		if err := json.Unmarshal([]byte(input), &result); err == nil {
			// Input is already valid JSON
			return result, nil
		}
	}

	// Use our repair parser
	parser := NewParser(input, opts...)
	return parser.Parse()
}

// Loads repairs JSON from a string and returns it as a Go value
// Alias for RepairJSONToValue for API compatibility
func Loads(input string, opts ...ParserOption) (JSONValue, error) {
	return RepairJSONToValue(input, opts...)
}

// Load repairs JSON from an io.Reader and returns it as a Go value
func Load(reader io.Reader, opts ...ParserOption) (JSONValue, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}
	return RepairJSONToValue(string(data), opts...)
}

// FromFile repairs JSON from a file and returns it as a Go value
func FromFile(filename string, opts ...ParserOption) (JSONValue, error) {
	// Check if file exists
	if _, err := os.Stat(filename); err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}

	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return RepairJSONToValue(string(data), opts...)
}

// RepairFile reads a JSON file, repairs it, and writes it back (or to a new file)
func RepairFile(inputPath, outputPath string, opts ...ParserOption) error {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Repair JSON
	repaired, err := RepairJSON(string(data), opts...)
	if err != nil {
		return fmt.Errorf("failed to repair JSON: %w", err)
	}

	// Determine output path
	if outputPath == "" {
		// Overwrite input file
		outputPath = inputPath
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil { //nolint:gosec // Directory permissions are appropriate
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write repaired JSON
	if err := os.WriteFile(outputPath, []byte(repaired), 0o600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// ValidJSON checks if a string is valid JSON
func ValidJSON(input string) bool {
	var result interface{}
	return json.Unmarshal([]byte(input), &result) == nil
}

// ValidJSONReader checks if the content from a reader is valid JSON
func ValidJSONReader(reader io.Reader) bool {
	data, err := io.ReadAll(reader)
	if err != nil {
		return false
	}
	return ValidJSON(string(data))
}

// mergeOptions combines multiple ParserOption into a single RepairOptions
func mergeOptions(opts ...ParserOption) *RepairOptions {
	options := DefaultRepairOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// marshalJSON marshals a value to JSON string based on options
func marshalJSON(value JSONValue, opts ...ParserOption) (string, error) {
	options := mergeOptions(opts...)

	// If value is already a string and we're in repair mode, return it
	if str, ok := value.(string); ok {
		if options.SkipJSONLoads {
			return str, nil
		}
	}

	// Marshal to JSON
	var data []byte
	var err error

	if options.Indent > 0 {
		data, err = json.MarshalIndent(value, "", strings.Repeat(" ", options.Indent))
	} else {
		data, err = json.Marshal(value)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	result := string(data)

	// Handle ASCII-only output
	if options.EnsureASCII {
		// JSON marshaling already ensures ASCII by escaping non-ASCII
		return result, nil
	}

	return result, nil
}

// EscapeString escapes a string for JSON
func EscapeString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		// Fall back to simple escaping if marshaling fails
		return `"` + s + `"`
	}
	return string(b)
}

// Unmarshal is a convenience function for unmarshaling JSON with repair
func Unmarshal(data []byte, v interface{}, opts ...ParserOption) error {
	str, err := RepairJSON(string(data), opts...)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(str), v)
}

// MustRepairJSON is like RepairJSON but panics if repair fails
func MustRepairJSON(input string, opts ...ParserOption) string {
	result, err := RepairJSON(input, opts...)
	if err != nil {
		panic(fmt.Sprintf("Failed to repair JSON: %v", err))
	}
	return result
}

// MustRepairJSONToValue is like RepairJSONToValue but panics if repair fails
func MustRepairJSONToValue(input string, opts ...ParserOption) JSONValue {
	result, err := RepairJSONToValue(input, opts...)
	if err != nil {
		panic(fmt.Sprintf("Failed to repair JSON: %v", err))
	}
	return result
}
