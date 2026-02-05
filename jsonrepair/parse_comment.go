package jsonrepair

import (
	"strings"
)

// skipComments removes all comments from the input
func (p *Parser) skipComments() {
	for p.index < len(p.input) {
		skipped := false

		// Skip whitespace first
		p.skipWhitespaces()

		// Try to skip different types of comments
		if p.skipLineComment() {
			skipped = true
		} else if p.skipBlockComment() {
			skipped = true
		} else if p.skipHashComment() {
			skipped = true
		}

		if !skipped {
			break
		}
	}
}

// skipLineComment skips a line comment (//)
func (p *Parser) skipLineComment() bool {
	remaining := p.input[p.index:]

	if strings.HasPrefix(remaining, "//") {
		// Line comment - skip to end of line
		for p.index < len(p.input) && p.input[p.index] != '\n' {
			p.index++
		}
		// Skip the newline too
		if p.index < len(p.input) && p.input[p.index] == '\n' {
			p.index++
		}
		p.log("Skipped line comment")
		return true
	}

	return false
}

// skipHashComment skips a hash comment (#)
func (p *Parser) skipHashComment() bool {
	remaining := p.input[p.index:]

	if strings.HasPrefix(remaining, "#") {
		// Hash comment - skip to end of line
		for p.index < len(p.input) && p.input[p.index] != '\n' {
			p.index++
		}
		// Skip the newline too
		if p.index < len(p.input) && p.input[p.index] == '\n' {
			p.index++
		}
		p.log("Skipped hash comment")
		return true
	}

	return false
}

// skipBlockComment skips a block comment (/* ... */)
func (p *Parser) skipBlockComment() bool {
	remaining := p.input[p.index:]

	if strings.HasPrefix(remaining, "/*") {
		// Block comment - find the end
		end := strings.Index(remaining[2:], "*/")
		if end >= 0 {
			p.index += 2 + end + 2 // Skip /* ... */
			p.log("Skipped block comment")
			return true
		}
		// Unterminated block comment - in repair mode, skip to end
		if !p.strict {
			p.log("Skipped unterminated block comment to end of input")
			p.index = len(p.input)
			return true
		}
	}

	return false
}

// isComment checks if the current position is at the start of a comment
func (p *Parser) isComment() bool {
	remaining := p.input[p.index:]
	return strings.HasPrefix(remaining, "//") ||
		strings.HasPrefix(remaining, "#") ||
		strings.HasPrefix(remaining, "/*")
}

// removeAllComments is a static function to remove comments from a string
func removeAllComments(input string, strict bool) string {
	var result strings.Builder
	i := 0

	for i < len(input) {
		// Skip whitespace
		for i < len(input) && isWhitespace(rune(input[i])) {
			result.WriteByte(input[i])
			i++
		}

		if i >= len(input) {
			break
		}

		remaining := input[i:]

		// Check for line comment
		if strings.HasPrefix(remaining, "//") {
			// Skip to end of line
			for i < len(input) && input[i] != '\n' {
				i++
			}
			continue
		}

		// Check for hash comment
		if strings.HasPrefix(remaining, "#") {
			// Skip to end of line
			for i < len(input) && input[i] != '\n' {
				i++
			}
			continue
		}

		// Check for block comment
		if strings.HasPrefix(remaining, "/*") {
			// Skip to end of block comment
			j := i + 2
			for j+1 < len(input) && (input[j] != '*' || input[j+1] != '/') {
				j++
			}
			if j+1 < len(input) {
				i = j + 2 // Skip */
			} else {
				// Unterminated block comment
				if strict {
					// In strict mode, we'll include it
					result.WriteString(input[i:j])
				}
				i = len(input)
			}
			continue
		}

		// Check for string literals
		if isStringDelimiter(rune(input[i])) {
			delim := input[i]
			result.WriteByte(input[i])
			i++
			for i < len(input) {
				if input[i] == '\\' && i+1 < len(input) {
					result.WriteByte(input[i])
					i++
					result.WriteByte(input[i])
					i++
				} else if input[i] == delim {
					result.WriteByte(input[i])
					i++
					break
				} else {
					result.WriteByte(input[i])
					i++
				}
			}
			continue
		}

		// Regular character
		result.WriteByte(input[i])
		i++
	}

	return result.String()
}
