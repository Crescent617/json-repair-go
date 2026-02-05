package jsonrepair

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// parseString parses a JSON string, handling various delimiters and repair
func (p *Parser) parseString() (string, error) {
	if p.index >= len(p.input) {
		return "", errors.New("unexpected end of input")
	}

	// Get the opening delimiter
	openingDelim, size := utf8.DecodeRuneInString(p.input[p.index:])
	isQuoted := false

	// Check if it's a valid string delimiter
	for _, delim := range StringDelimiters {
		if openingDelim == delim {
			isQuoted = true
			break
		}
	}

	// If not quoted, we'll try to parse as an unquoted string (repair mode)
	if !isQuoted {
		if p.strict {
			return "", p.unexpectedCharError("string delimiter")
		}
		// In repair mode, try to parse as unquoted string
		return p.parseUnquotedString()
	}

	// Skip opening delimiter
	p.index += size

	// Parse the string content
	var result strings.Builder

	for p.index < len(p.input) {
		// Check for closing delimiter
		if ch, size := utf8.DecodeRuneInString(p.input[p.index:]); ch == openingDelim {
			// Check if this is truly the end of the string or an unescaped quote
			// Look ahead to see if this quote is followed by a delimiter that indicates string end
			if p.isAtStringEnd(p.index + size) {
				p.index += size
				return result.String(), nil
			}
			// Not at string end, this is an unescaped quote in repair mode
			if !p.strict {
				result.WriteRune(ch)
				p.index += size
				p.log("Repaired unescaped quote in string")
				continue
			}
		}

		// Also check for any valid string delimiter as closing (for mixed quotes)
		// This handles cases like {"key": "value"} where smart quote opens but double quote closes
		if ch, size := utf8.DecodeRuneInString(p.input[p.index:]); isStringDelimiter(ch) && ch != openingDelim {
			if p.isAtStringEnd(p.index + size) {
				p.index += size
				return result.String(), nil
			}
			// Not at string end, this is an unescaped quote in repair mode
			if !p.strict {
				result.WriteRune(ch)
				p.index += size
				p.log("Repaired unescaped quote in string")
				continue
			}
		}

		// Handle escape sequences
		if p.input[p.index] == '\\' {
			p.index++
			if p.index >= len(p.input) {
				break
			}

			escapedChar, size := utf8.DecodeRuneInString(p.input[p.index:])
			unescaped, err := p.unescapeChar(escapedChar)
			if err != nil {
				return "", err
			}
			result.WriteRune(unescaped)
			p.index += size
			continue
		}

		// Handle Unicode escape sequences
		if strings.HasPrefix(p.input[p.index:], "\\u") {
			unicodeChar, newIndex, err := p.parseUnicodeEscape()
			if err != nil {
				return "", err
			}
			result.WriteRune(unicodeChar)
			p.index = newIndex
			continue
		}

		// Regular character
		char, size := utf8.DecodeRuneInString(p.input[p.index:])
		result.WriteRune(char)
		p.index += size
	}

	// If we get here, the string wasn't properly closed
	if p.strict {
		return "", errors.New("unterminated string")
	}

	// In repair mode, return what we have
	p.log("Repaired unterminated string")
	return result.String(), nil
}

// parseUnquotedString parses a string without quotes (repair mode)
func (p *Parser) parseUnquotedString() (string, error) {
	p.log("Parsing unquoted string (repair mode)")

	var result strings.Builder

	for p.index < len(p.input) {
		ch, size := utf8.DecodeRuneInString(p.input[p.index:])

		// Stop at certain characters that indicate end of string
		if ch == ',' || ch == '}' || ch == ']' || ch == ':' || isWhitespace(ch) {
			break
		}

		// Skip escape sequences for now (shouldn't be in unquoted strings)
		if ch == '\\' {
			p.index += size // Skip the backslash
			if p.index < len(p.input) {
				nextChar, nextSize := utf8.DecodeRuneInString(p.input[p.index:])
				result.WriteRune(nextChar)
				p.index += nextSize
			}
			continue
		}

		result.WriteRune(ch)
		p.index += size
	}

	if result.Len() == 0 {
		return "", errors.New("expected string value")
	}

	return result.String(), nil
}

// unescapeChar converts an escape sequence to its actual character
func (p *Parser) unescapeChar(escaped rune) (rune, error) {
	switch escaped {
	case '"':
		return '"', nil
	case '\\':
		return '\\', nil
	case '/':
		return '/', nil
	case 'b':
		return '\b', nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	default:
		if p.strict {
			return 0, fmt.Errorf("invalid escape sequence: \\%c", escaped)
		}
		// In repair mode, return the character as-is
		p.log(fmt.Sprintf("Repaired invalid escape sequence: \\%c", escaped))
		return escaped, nil
	}
}

// parseUnicodeEscape parses a Unicode escape sequence (\uXXXX)
func (p *Parser) parseUnicodeEscape() (rune, int, error) {
	if !strings.HasPrefix(p.input[p.index:], "\\u") {
		return 0, p.index, errors.New("expected Unicode escape sequence")
	}

	pos := p.index + 2 // Skip \u

	if pos+4 > len(p.input) {
		if p.strict {
			return 0, pos, errors.New("incomplete Unicode escape sequence")
		}
		// In repair mode, try to get what we can
		hexDigits := p.input[pos:]
		if len(hexDigits) < 1 {
			return 0, pos, errors.New("invalid Unicode escape")
		}
		// Pad with zeros
		hexDigits += "0000"
		hexDigits = hexDigits[:4]

		codePoint, err := strconv.ParseInt(hexDigits, 16, 32)
		if err != nil {
			return 0, pos, fmt.Errorf("invalid Unicode escape: %s", hexDigits)
		}

		p.log(fmt.Sprintf("Repaired incomplete Unicode escape: \\u%s", hexDigits))
		return rune(codePoint), pos + len(hexDigits), nil
	}

	hexDigits := p.input[pos : pos+4]
	codePoint, err := strconv.ParseInt(hexDigits, 16, 32)
	if err != nil {
		if p.strict {
			return 0, pos, fmt.Errorf("invalid Unicode escape: \\u%s", hexDigits)
		}
		// Try to repair by checking each digit
		repairedHex := repairHexDigits(hexDigits)
		codePoint, err2 := strconv.ParseInt(repairedHex, 16, 32)
		if err2 != nil {
			// Return replacement character
			p.log("Repaired invalid Unicode escape with replacement character")
			return '\uFFFD', pos + 4, err2
		}
		p.log(fmt.Sprintf("Repaired Unicode escape: \\u%s -> \\u%s", hexDigits, repairedHex))
		return rune(codePoint), pos + 4, nil
	}

	// Handle surrogate pairs
	if isHighSurrogate(rune(codePoint)) && pos+10 <= len(p.input) {
		// Check for low surrogate
		if strings.HasPrefix(p.input[pos+4:], "\\u") {
			lowHex := p.input[pos+6 : pos+10]
			lowCodePoint, err := strconv.ParseInt(lowHex, 16, 32)
			if err == nil && isLowSurrogate(rune(lowCodePoint)) {
				// Combine surrogates
				unicodeChar := decodeSurrogatePair(rune(codePoint), rune(lowCodePoint))
				return unicodeChar, pos + 10, nil
			}
		}
	}

	return rune(codePoint), pos + 4, nil
}

// isHighSurrogate checks if a rune is a Unicode high surrogate
func isHighSurrogate(r rune) bool {
	return r >= 0xD800 && r <= 0xDBFF
}

// isLowSurrogate checks if a rune is a Unicode low surrogate
func isLowSurrogate(r rune) bool {
	return r >= 0xDC00 && r <= 0xDFFF
}

// decodeSurrogatePair combines a high and low surrogate into a single Unicode character
func decodeSurrogatePair(high, low rune) rune {
	// Formula from Unicode standard
	return 0x10000 + (high-0xD800)*0x400 + (low - 0xDC00)
}

// repairHexDigits attempts to repair invalid hex digits in Unicode escapes
func repairHexDigits(hex string) string {
	var result []rune
	for _, ch := range hex {
		if unicode.In(ch, unicode.ASCII_Hex_Digit) {
			result = append(result, ch)
		} else {
			// Replace invalid hex digit with '0'
			result = append(result, '0')
		}
	}
	return string(result)
}
