package jsonrepair

// Package constants for JSON parsing and repair

// StringDelimiters defines the valid quote characters for strings
// Includes standard double quote, single quote, and Unicode smart quotes
var StringDelimiters = []rune{'"', '\'', '\u201c', '\u201d'}

// WhitespaceChars defines characters considered as whitespace
var WhitespaceChars = []rune{
	' ',      // space
	'\t',     // tab
	'\n',     // newline
	'\r',     // carriage return
	'\f',     // form feed
	'\u00A0', // non-breaking space
	'\u2000', // en quad
	'\u2001', // em quad
	'\u2002', // en space
	'\u2003', // em space
	'\u2004', // three-per-em space
	'\u2005', // four-per-em space
	'\u2006', // six-per-em space
	'\u2007', // figure space
	'\u2008', // punctuation space
	'\u2009', // thin space
	'\u200A', // hair space
	'\u200B', // zero width space
	'\u202F', // narrow no-break space
	'\u205F', // medium mathematical space
	'\u3000', // ideographic space
}

// MissingValue is a special sentinel value used to represent missing values in JSON
var MissingValue = struct{}{}

// NumericChars are characters that can appear in valid JSON numbers
var NumericChars = []rune{
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'.', '+', '-', 'e', 'E',
}

// Boolean and null literals
const (
	TrueLiteral  = "true"
	FalseLiteral = "false"
	NullLiteral  = "null"
)

// JSON structural characters
const (
	ObjectStart    = '{'
	ObjectEnd      = '}'
	ArrayStart     = '['
	ArrayEnd       = ']'
	StringQuote    = '"'
	PairSeparator  = ':'
	ValueSeparator = ','
)

// Comment markers
const (
	LineCommentStart  = "//"
	HashCommentStart  = "#"
	BlockCommentStart = "/*"
	BlockCommentEnd   = "*/"
)

// String escape sequences
var EscapeSequences = map[rune]rune{
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

// Control characters that need escaping in JSON strings
var ControlChars = []rune{
	'\u0000', '\u0001', '\u0002', '\u0003', '\u0004', '\u0005', '\u0006',
	'\u0007', '\u0008', '\u0009', '\u000A', '\u000B', '\u000C', '\u000D',
	'\u000E', '\u000F', '\u0010', '\u0011', '\u0012', '\u0013', '\u0014',
	'\u0015', '\u0016', '\u0017', '\u0018', '\u0019', '\u001A', '\u001B',
	'\u001C', '\u001D', '\u001E', '\u001F',
}

// Default values for various options
const (
	DefaultMaxDepth   = 1000
	DefaultIndentSize = 2
)

// Schema validation keywords
const (
	SchemaType        = "type"
	SchemaProperties  = "properties"
	SchemaRequired    = "required"
	SchemaDefault     = "default"
	SchemaItems       = "items"
	SchemaPattern     = "pattern"
	SchemaEnum        = "enum"
	SchemaFormat      = "format"
	SchemaMinimum     = "minimum"
	SchemaMaximum     = "maximum"
	SchemaMinLength   = "minLength"
	SchemaMaxLength   = "maxLength"
	SchemaMinItems    = "minItems"
	SchemaMaxItems    = "maxItems"
	SchemaUniqueItems = "uniqueItems"
)

// JSON data types as defined in JSON Schema
const (
	TypeObject  = "object"
	TypeArray   = "array"
	TypeString  = "string"
	TypeNumber  = "number"
	TypeInteger = "integer"
	TypeBoolean = "boolean"
	TypeNull    = "null"
)

// Common formats for JSON Schema format validation
const (
	FormatDateTime = "date-time"
	FormatDate     = "date"
	FormatTime     = "time"
	FormatEmail    = "email"
	FormatHostname = "hostname"
	FormatIPv4     = "ipv4"
	FormatIPv6     = "ipv6"
	FormatURI      = "uri"
	FormatUUID     = "uuid"
)
