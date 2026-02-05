package jsonrepair

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// NewParser creates a new JSON parser with optional configuration
func NewParser(input string, opts ...ParserOption) *Parser {
	options := DefaultRepairOptions()
	for _, opt := range opts {
		opt(options)
	}

	parser := &Parser{
		input:      input,
		index:      0,
		context:    &Context{current: ContextRoot},
		options:    options,
		logging:    options.Logging,
		logger:     make([]LogEntry, 0),
		strict:     options.Strict,
		parseDepth: 0,
	}

	if options.StreamStable {
		parser.streamStable = true
	}

	if options.Schema != nil {
		parser.schemaRepairer = &SchemaRepairer{
			schema: options.Schema,
		}
	}

	return parser
}

// Parse parses and repairs the JSON input
func (p *Parser) Parse() (JSONValue, error) {
	p.log("Starting JSON repair parse")

	// Handle empty input
	if len(p.input) == 0 || strings.TrimSpace(p.input) == "" {
		return nil, errors.New("empty input")
	}

	// Skip BOM if present
	p.skipBOM()

	// Parse the JSON value
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	// Skip trailing whitespace
	p.skipWhitespaces()

	// Check if we consumed all input (unless in stream-stable mode)
	if !p.streamStable && p.index < len(p.input) {
		remaining := strings.TrimSpace(p.input[p.index:])
		if remaining != "" {
			return nil, fmt.Errorf("unexpected trailing content: %s", remaining[:min(20, len(remaining))])
		}
	}

	return value, nil
}

// ParseWithSchema parses JSON with schema-guided repair
func (p *Parser) ParseWithSchema(schema interface{}) (JSONValue, error) {
	p.schemaRepairer = &SchemaRepairer{schema: schema}
	p.options.Schema = schema
	return p.Parse()
}

// parseValue parses a JSON value (object, array, string, number, boolean, or null)
func (p *Parser) parseValue() (JSONValue, error) {
	startIndex := p.index
	p.skipWhitespaces()

	if p.index >= len(p.input) {
		return nil, errors.New("unexpected end of input")
	}

	ch, _ := p.getCharAt(0)

	// Handle comments before value - must check before safety check
	if p.skipComment() {
		return p.parseValue()
	}

	// If skipWhitespaces made no progress and we're not at a valid start, error out
	if p.index == startIndex {
		if ch != '{' && ch != '[' && ch != '"' && ch != '\'' && ch != '\u201c' && ch != '\u201d' &&
			ch != 't' && ch != 'T' && ch != 'f' && ch != 'F' && ch != 'n' && ch != 'N' &&
			ch != '-' && !unicode.IsDigit(ch) {
			return nil, fmt.Errorf("cannot parse value at position %d", p.index)
		}
	}

	switch ch {
	case '{':
		p.log("Parsing object")
		return p.parseObject()
	case '[':
		p.log("Parsing array")
		return p.parseArray()
	case '"', '\'', '\u201c', '\u201d':
		p.log("Parsing string")
		return p.parseString()
	case 'n':
		if p.index+4 <= len(p.input) && p.input[p.index:p.index+4] == "null" {
			p.index += 4
			return nil, nil //nolint:nilnil // Returning nil for JSON null value is valid
		}
		fallthrough
	case 't':
		if p.index+4 <= len(p.input) && p.input[p.index:p.index+4] == "true" {
			p.index += 4
			return true, nil
		}
		fallthrough
	case 'f':
		if p.index+5 <= len(p.input) && p.input[p.index:p.index+5] == "false" {
			p.index += 5
			return false, nil
		}
		fallthrough
	default:
		// Could be a number or a malformed value
		if unicode.IsDigit(ch) || ch == '-' || ch == '+' {
			return p.parseNumber()
		}

		// Try to repair malformed values
		return p.attemptValueRepair()
	}
}

// attemptValueRepair attempts to repair a malformed value
func (p *Parser) attemptValueRepair() (JSONValue, error) {
	p.log("Attempting value repair")

	// If we have schema repairer, try schema-guided repair
	if p.schemaRepairer != nil && !p.strict {
		return p.schemaRepairer.attemptSchemaRepair(p)
	}

	// Check if we're at the end of input
	if p.index >= len(p.input) {
		return nil, errors.New("unexpected end of input")
	}

	// Try to determine what type of value this might be
	remaining := p.input[p.index:]
	ch, _ := utf8.DecodeRuneInString(remaining)

	// Check if it looks like a string (starts with a quote)
	if isStringDelimiter(ch) {
		return p.parseString()
	}

	// Check if it looks like a number
	if unicode.IsDigit(ch) || ch == '-' || ch == '+' || ch == '.' {
		return p.parseNumber()
	}

	// Check for structural characters (don't try to parse these as strings)
	if ch == '{' || ch == '}' || ch == '[' || ch == ']' || ch == ':' || ch == ',' {
		return nil, fmt.Errorf("unexpected character %c", ch)
	}

	// For anything else, try unquoted string parsing (but with safeguard)
	startPos := p.index
	value, err := p.parseUnquotedString()
	if err != nil {
		return nil, err
	}

	// If we made no progress, something is wrong
	if p.index == startPos {
		return nil, fmt.Errorf("stuck at position %d", p.index)
	}

	return value, nil
}

// skipBOM skips the Byte Order Mark if present
func (p *Parser) skipBOM() {
	if strings.HasPrefix(p.input[p.index:], "\uFEFF") {
		p.index += 3
		p.log("Skipped BOM")
	}
}

// skipWhitespaces skips whitespace characters
func (p *Parser) skipWhitespaces() {
	for p.index < len(p.input) {
		ch, size := utf8.DecodeRuneInString(p.input[p.index:])
		if !isWhitespace(ch) {
			break
		}
		p.index += size
	}
}

// skipComment skips comments (# and //)
func (p *Parser) skipComment() bool {
	if p.index >= len(p.input) {
		return false
	}

	// To prevent infinite recursion, mark that we've attempted to skip a comment
	// and check if we're making progress
	currentIndex := p.index

	remaining := p.input[p.index:]

	// Single line comments (# or //)
	if strings.HasPrefix(remaining, "#") || strings.HasPrefix(remaining, "//") {
		p.index = p.skipToEndOfLine()
		if p.index <= currentIndex {
			// We made no progress, something is wrong
			p.index++ // Force progress
		}
		p.log("Skipped comment")
		return true
	}

	// Block comments (/* */) - optional
	if strings.HasPrefix(remaining, "/*") {
		end := strings.Index(remaining[2:], "*/")
		if end >= 0 {
			p.index += 2 + end + 2
			if p.index <= currentIndex {
				// We should always make progress with a valid block comment
				p.index = currentIndex + 1
			}
			p.log("Skipped block comment")
			return true
		}
	}

	return false
}

// skipToEndOfLine skips to the end of the current line
func (p *Parser) skipToEndOfLine() int {
	start := p.index
	for p.index < len(p.input) && p.input[p.index] != '\n' {
		p.index++
	}
	// Skip the newline itself
	if p.index < len(p.input) && p.input[p.index] == '\n' {
		p.index++
	}
	return p.index - start
}

// skipToCharacter skips until the target character is found
func (p *Parser) skipToCharacter(target rune, startOffset int) int {
	pos := startOffset
	for pos < len(p.input) {
		ch, _ := p.getCharAt(pos)
		if ch == target {
			return pos
		}
		pos++
	}
	return -1
}

// getCharAt returns the character at the given offset from current position
func (p *Parser) getCharAt(offset int) (rune, bool) {
	pos := p.index + offset
	if pos >= len(p.input) {
		return 0, false
	}
	ch, _ := utf8.DecodeRuneInString(p.input[pos:])
	return ch, true
}

// peek returns the next character without advancing
func (p *Parser) peek() (rune, bool) {
	return p.getCharAt(0)
}

// peekN returns the character n positions ahead without advancing
func (p *Parser) peekN(n int) (rune, bool) {
	return p.getCharAt(n)
}

// log adds a log entry if logging is enabled
func (p *Parser) log(text string) {
	if p.logging {
		entry := LogEntry{
			Text:    text,
			Context: p.context.current.String(),
		}
		p.logger = append(p.logger, entry)
	}
}

// getLogger returns the parser's log entries
func (p *Parser) getLogger() []LogEntry {
	return p.logger
}

// isWhitespace checks if a rune is considered whitespace
func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f' ||
		(ch >= '\u00A0' && ch <= '\u200A' && string(ch) != "\u200B") ||
		ch == '\u202F' || ch == '\u205F' || ch == '\u3000'
}

// unexpectedCharError creates an error for unexpected characters
func (p *Parser) unexpectedCharError(expected string) error {
	ch, _ := p.peek()
	pos := p.index

	if pos >= len(p.input) {
		return fmt.Errorf("unexpected end of input, expected %s", expected)
	}

	// Show some context around the error
	contextStart := max(0, pos-10)
	contextEnd := min(len(p.input), pos+10)
	context := p.input[contextStart:contextEnd]

	return fmt.Errorf("unexpected character '%c' at position %d (context: ...%s...), expected %s",
		ch, pos, context, expected)
}

// ParseString is a utility function for parsing a string with options
func ParseString(input string, opts ...ParserOption) (JSONValue, error) {
	parser := NewParser(input, opts...)
	return parser.Parse()
}

// ParseBytes is a utility function for parsing bytes with options
func ParseBytes(data []byte, opts ...ParserOption) (JSONValue, error) {
	return ParseString(string(data), opts...)
}

// isAtStringEnd checks if a quote at the given position is a string terminator
func (p *Parser) isAtStringEnd(pos int) bool {
	// If we're in strict mode, any quote match terminates the string
	if p.strict {
		return true
	}

	// Look ahead to see what comes after this quote
	if pos >= len(p.input) {
		return true // End of input, must be string terminator
	}

	// Skip whitespace after the quote
	nextPos := pos
	for nextPos < len(p.input) {
		ch, size := utf8.DecodeRuneInString(p.input[nextPos:])
		if !isWhitespace(ch) {
			break
		}
		nextPos += size
	}

	// If we're at end of input, it's a string terminator
	if nextPos >= len(p.input) {
		return true
	}

	// Look at the next non-whitespace character
	nextChar, _ := utf8.DecodeRuneInString(p.input[nextPos:])

	// If followed by structural characters, it's terminator
	// : - next key-value pair
	// , - next element
	// } - end of object
	// ] - end of array
	if nextChar == ':' || nextChar == ',' || nextChar == '}' || nextChar == ']' {
		return true
	}

	// Check if the context indicates we're after a key name
	if p.context.current == ContextObjectKey && nextChar == '"' {
		// This could be a key-value separator
		return true
	}

	// If next position looks like it could be a key or value, this quote is unescaped
	return false
}
