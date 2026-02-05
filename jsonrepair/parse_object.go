package jsonrepair

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// parseObject parses a JSON object, handling missing quotes and malformed syntax
func (p *Parser) parseObject() (map[string]JSONValue, error) {
	if p.index >= len(p.input) {
		return nil, errors.New("unexpected end of input")
	}

	// Prevent infinite recursion
	p.parseDepth++
	if p.parseDepth > 1000 {
		return nil, errors.New("maximum parse depth exceeded")
	}
	defer func() { p.parseDepth-- }()

	// Check and skip opening brace
	if ch, size := utf8.DecodeRuneInString(p.input[p.index:]); ch != '{' {
		if p.strict {
			return nil, p.unexpectedCharError("{")
		}
		// In repair mode, assume missing opening brace
		p.log("Repairing missing opening brace in object")
	} else {
		p.index += size
	}

	p.context.current = ContextObjectKey
	object := make(map[string]JSONValue)

	iterations := 0
	for p.index < len(p.input) {
		iterations++
		if iterations > 10000 {
			return nil, errors.New("too many object iterations, possible infinite loop")
		}
		// Skip whitespace and comments
		p.skipWhitespaces()
		if p.skipComment() {
			continue
		}

		// Check for closing brace
		if ch, size := utf8.DecodeRuneInString(p.input[p.index:]); ch == '}' {
			p.index += size
			p.context.current = ContextRoot
			return object, nil
		}

		// Parse key-value pair
		key, value, err := p.parseObjectPair()
		if err != nil {
			if p.strict {
				return nil, err
			}
			p.log(fmt.Sprintf("Error parsing object pair: %v", err))

			// Try to recover by skipping until next comma or closing brace
			for p.index < len(p.input) {
				ch, size := utf8.DecodeRuneInString(p.input[p.index:])
				if ch == ',' || ch == '}' {
					break
				}
				p.index += size
			}
			continue
		}

		// Store the key-value pair (overwrite duplicates, matching JSON behavior)
		object[key] = value

		// Skip whitespace
		p.skipWhitespaces()

		// Check for comma or closing brace
		ch, size := utf8.DecodeRuneInString(p.input[p.index:])
		if ch == '}' {
			p.index += size
			p.context.current = ContextRoot
			return object, nil
		}

		if ch == ',' {
			p.index += size
			p.context.current = ContextObjectKey
			continue
		}

		// Missing comma (repair mode)
		if !p.strict {
			p.log("Repairing missing comma in object")
			p.context.current = ContextObjectKey
			continue
		}

		// Unexpected character
		return nil, p.unexpectedCharError(", or }")
	}

	// Unclosed object (repair mode)
	if !p.strict {
		p.log("Repairing unclosed object")
		p.context.current = ContextRoot
		return object, nil
	}

	return nil, errors.New("unterminated object")
}

// parseObjectPair parses a key-value pair within an object
func (p *Parser) parseObjectPair() (string, JSONValue, error) {
	// Parse the key
	p.context.current = ContextObjectKey

	key, err := p.parseObjectKey()
	if err != nil {
		return "", nil, fmt.Errorf("error parsing object key: %w", err)
	}

	// Skip whitespace
	p.skipWhitespaces()

	// Check for colon
	if ch, size := utf8.DecodeRuneInString(p.input[p.index:]); ch == ':' {
		p.index += size
	} else if !p.strict {
		// Missing colon (repair mode)
		p.log(fmt.Sprintf("Repairing missing colon after key: %s", key))
	} else {
		return "", nil, p.unexpectedCharError(":")
	}

	// Skip whitespace
	p.skipWhitespaces()

	// Parse the value
	p.context.current = ContextObjectValue

	value, err := p.parseValue()
	if err != nil {
		if p.strict {
			return "", nil, fmt.Errorf("error parsing value for key '%s': %w", key, err)
		}

		// Try to repair by inserting null value
		p.log(fmt.Sprintf("Repairing missing value for key '%s' with null", key))
		value = nil
	}

	return key, value, nil
}

// parseObjectKey parses an object key, handling quoted and unquoted keys
func (p *Parser) parseObjectKey() (string, error) {
	p.skipWhitespaces()

	if p.index >= len(p.input) {
		return "", errors.New("unexpected end of input")
	}

	ch, _ := utf8.DecodeRuneInString(p.input[p.index:])

	// Check if it's a quoted string (normal case)
	if isStringDelimiter(ch) {
		// parseString expects to be at the opening delimiter, so don't advance here
		return p.parseString()
	}

	// Not quoted - try to parse as unquoted key (repair mode)
	if p.strict {
		return "", p.unexpectedCharError("string")
	}

	p.log("Parsing unquoted object key (repair mode)")
	return p.parseUnquotedObjectKey()
}

// parseUnquotedObjectKey parses an object key without quotes (repair mode)
func (p *Parser) parseUnquotedObjectKey() (string, error) {
	var result strings.Builder
	startPos := p.index

	for p.index < len(p.input) {
		ch, size := utf8.DecodeRuneInString(p.input[p.index:])

		// Stop at characters that indicate end of key
		if ch == ':' || ch == ',' || ch == '}' || isWhitespace(ch) {
			break
		}

		result.WriteRune(ch)
		p.index += size
	}

	key := result.String()

	if key == "" {
		// Try to generate a key from the next value
		p.index = startPos
		return p.generateKeyFromValue()
	}

	p.log(fmt.Sprintf("Repaired unquoted key: %s", key))
	return key, nil
}

// generateKeyFromValue tries to generate a key based on the upcoming value
func (p *Parser) generateKeyFromValue() (string, error) {
	// Look ahead to see what type of value is coming
	originalPos := p.index
	defer func() { p.index = originalPos }()

	p.skipWhitespaces()

	if p.index >= len(p.input) {
		return "", errors.New("cannot generate key from end of input")
	}

	ch, _ := utf8.DecodeRuneInString(p.input[p.index:])

	// Generate key based on value type
	switch ch {
	case '{':
		return "object", nil
	case '[':
		return "array", nil
	case '"', '\'', '\u201c', '\u201d':
		return "string", nil
	default:
		if unicode.IsDigit(ch) || ch == '-' || ch == '+' {
			return "number", nil
		}
		return "value", nil
	}
}

// isStringDelimiter checks if a rune is a valid string delimiter
func isStringDelimiter(ch rune) bool {
	for _, delim := range StringDelimiters {
		if ch == delim {
			return true
		}
	}
	return false
}

// attemptObjectRepair attempts to repair a malformed object
func (p *Parser) attemptObjectRepair() (map[string]JSONValue, error) {
	p.log("Attempting object repair")

	result := make(map[string]JSONValue)

	// Try to interpret as much as we can as key-value pairs
	for p.index < len(p.input) {
		p.skipWhitespaces()

		// Try to find a key (look for string or identifier)
		keyStart := p.index
		for p.index < len(p.input) {
			ch, size := utf8.DecodeRuneInString(p.input[p.index:])
			if ch == ':' || ch == ',' || ch == '}' || isWhitespace(ch) {
				break
			}
			p.index += size
		}

		key := strings.TrimSpace(p.input[keyStart:p.index])
		if key == "" {
			break
		}

		// Try to find a colon
		p.skipWhitespaces()
		if ch, _ := utf8.DecodeRuneInString(p.input[p.index:]); ch == ':' {
			p.index++
			p.skipWhitespaces()
		}

		// Try to parse a value (up to next comma or closing brace)
		valueStart := p.index
		for p.index < len(p.input) {
			ch, _ := utf8.DecodeRuneInString(p.input[p.index:])
			if ch == ',' || ch == '}' {
				break
			}
			p.index++
		}

		if p.index > valueStart {
			valueStr := p.input[valueStart:p.index]
			// Try to parse the value string
			parser := NewParser(valueStr, WithLogging())
			if value, err := parser.Parse(); err == nil {
				result[key] = value
			}
		}

		// Skip comma if present
		if ch, _ := utf8.DecodeRuneInString(p.input[p.index:]); ch == ',' {
			p.index++
		}
	}

	p.log(fmt.Sprintf("Repaired object with %d keys", len(result)))
	return result, nil
}
