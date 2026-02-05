// Package helpers provides utility functions for parsing JSON values.
package helpers

import (
	"errors"
	"strings"
)

// ParseBooleanOrNull attempts to parse a boolean or null value from the input string
// starting at the given position. Returns the parsed value, the new position,
// and whether parsing was successful.
func ParseBooleanOrNull(input string, pos int) (interface{}, int, bool) {
	if pos >= len(input) {
		return nil, pos, false
	}

	// Check for "true"
	if strings.HasPrefix(input[pos:], "true") {
		return true, pos + 4, true
	}

	// Check for "false"
	if strings.HasPrefix(input[pos:], "false") {
		return false, pos + 5, true
	}

	// Check for "null"
	if strings.HasPrefix(input[pos:], "null") {
		return nil, pos + 4, true
	}

	// Try case-insensitive matching for repair mode
	if pos+4 <= len(input) {
		substr := input[pos : pos+4]
		lower := strings.ToLower(substr)

		if lower == "true" {
			return true, pos + 4, true
		}

		// Check for "null" (4 chars)
		if lower == "null" {
			return nil, pos + 4, true
		}
	}

	if pos+5 <= len(input) {
		substr := input[pos : pos+5]
		lower := strings.ToLower(substr)

		if lower == "false" {
			return false, pos + 5, true
		}
	}

	return nil, pos, false
}

// IsBooleanOrNullPrefix checks if the input at the given position could be
// the start of a boolean or null literal
func IsBooleanOrNullPrefix(input string, pos int) bool {
	if pos >= len(input) {
		return false
	}

	ch := input[pos]
	switch ch {
	case 't', 'T', 'f', 'F', 'n', 'N':
		return true
	default:
		return false
	}
}

// RecognizeBooleanOrNullWithRepair attempts to recognize boolean/null values
// even when they are malformed (e.g., missing characters)
func RecognizeBooleanOrNullWithRepair(input string, pos int) (interface{}, int, error) {
	// First try exact matching
	if val, newPos, ok := ParseBooleanOrNull(input, pos); ok {
		return val, newPos, nil
	}

	// Try to repair common issues
	remaining := input[pos:]

	// Try to match partial "true"
	if len(remaining) >= 1 && (remaining[0] == 't' || remaining[0] == 'T') {
		if len(remaining) >= 4 {
			substr := remaining[:4]
			lower := strings.ToLower(substr)
			if lower == "true" {
				return true, pos + 4, nil
			}
		} else if strings.HasPrefix("true", strings.ToLower(remaining)) {
			// Partial match - consider it as true but need to check if more input follows
			return true, pos + len(remaining), errors.New("incomplete boolean literal")
		}
	}

	// Try to match partial "false"
	if len(remaining) >= 1 && (remaining[0] == 'f' || remaining[0] == 'F') {
		if len(remaining) >= 5 {
			substr := remaining[:5]
			lower := strings.ToLower(substr)
			if lower == "false" {
				return false, pos + 5, nil
			}
		} else if strings.HasPrefix("false", strings.ToLower(remaining)) {
			// Partial match
			return false, pos + len(remaining), errors.New("incomplete boolean literal")
		}
	}

	// Try to match partial "null"
	if len(remaining) >= 1 && (remaining[0] == 'n' || remaining[0] == 'N') {
		if len(remaining) >= 4 {
			substr := remaining[:4]
			lower := strings.ToLower(substr)
			if lower == "null" {
				return nil, pos + 4, nil
			}
		} else if strings.HasPrefix("null", strings.ToLower(remaining)) {
			// Partial match
			return nil, pos + len(remaining), errors.New("incomplete null literal")
		}
	}

	return nil, pos, errors.New("unrecognized boolean or null literal")
}

// BooleanToString converts a boolean to its JSON string representation
func BooleanToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
