package jsonrepair

import (
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"
)

// parseArray parses a JSON array, handling missing brackets and commas
func (p *Parser) parseArray() ([]JSONValue, error) {
	if p.index >= len(p.input) {
		return nil, errors.New("unexpected end of input")
	}

	// Prevent infinite recursion
	p.parseDepth++
	if p.parseDepth > 1000 {
		return nil, errors.New("maximum parse depth exceeded")
	}
	defer func() { p.parseDepth-- }()

	// Check and skip opening bracket
	if ch, size := utf8.DecodeRuneInString(p.input[p.index:]); ch != '[' {
		if p.strict {
			return nil, p.unexpectedCharError("[")
		}
		// In repair mode, assume missing opening bracket
		p.log("Repairing missing opening bracket in array")
	} else {
		p.index += size
	}

	p.context.current = ContextArray
	var elements []JSONValue

	iterations := 0
	for p.index < len(p.input) {
		iterations++
		if iterations > 10000 {
			return nil, errors.New("too many array iterations, possible infinite loop")
		}
		// Skip whitespace and comments
		p.skipWhitespaces()
		if p.skipComment() {
			continue
		}

		// Check for closing bracket
		if ch, size := utf8.DecodeRuneInString(p.input[p.index:]); ch == ']' {
			p.index += size
			p.context.current = ContextRoot
			return elements, nil
		}

		// Parse array element
		element, err := p.parseValue()
		if err != nil {
			if p.strict {
				return nil, err
			}
			// Try to repair by skipping invalid element
			p.log(fmt.Sprintf("Skipping invalid array element: %v", err))

			// Skip until next comma or closing bracket
			for p.index < len(p.input) {
				ch, size := utf8.DecodeRuneInString(p.input[p.index:])
				if ch == ',' || ch == ']' {
					break
				}
				// Always advance by at least 1 to avoid infinite loops
				if size == 0 {
					p.index++
				} else {
					p.index += size
				}
			}
			continue
		}

		elements = append(elements, element)

		// Skip whitespace
		p.skipWhitespaces()

		// Check for comma or closing bracket
		ch, size := utf8.DecodeRuneInString(p.input[p.index:])
		if ch == ']' {
			p.index += size
			p.context.current = ContextRoot
			return elements, nil
		}

		if ch == ',' {
			p.index += size
			continue
		}

		// Missing comma (repair mode)
		if !p.strict && (isValueStart(ch) || ch == ']' || ch == '}') {
			p.log("Repairing missing comma in array")
			continue
		}

		// If we get here, we found something unexpected after a value
		if p.strict {
			return nil, p.unexpectedCharError(", or ]")
		}

		// In repair mode, break the loop if we can't make progress
		p.log("Unexpected character after array element, stopping")
		break
	}

	// Unclosed array (repair mode)
	if !p.strict {
		p.log("Repairing unclosed array")
		p.context.current = ContextRoot
		return elements, nil
	}

	return nil, errors.New("unterminated array")
}

// isValueStart checks if a character can start a JSON value
func isValueStart(ch rune) bool {
	return ch == '{' || ch == '[' || ch == '"' || ch == '\'' ||
		ch == 't' || ch == 'T' || ch == 'f' || ch == 'F' ||
		ch == 'n' || ch == 'N' || ch == '-' || ch == '+' ||
		unicode.IsDigit(ch)
}
