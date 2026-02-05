package jsonrepair

// JSONValue represents any valid JSON value (object, array, string, number, boolean, or null)
// This is an alias for interface{} to provide flexibility in handling dynamic JSON values
type JSONValue any

// Parser is the main parser struct that handles JSON repair
// It maintains the parsing state and configuration options
type Parser struct {
	input          string
	index          int
	context        *Context
	logging        bool
	logger         []LogEntry
	streamStable   bool
	strict         bool
	schemaRepairer *SchemaRepairer
	options        *RepairOptions
	parseDepth     int // Track recursion depth to prevent infinite loops
}

// Context tracks the current parsing context for proper error handling and validation
type Context struct {
	current ContextValue
}

// ContextValue represents the different contexts a parser can be in
type ContextValue int

const (
	// ContextObjectKey indicates we're parsing an object key
	ContextObjectKey ContextValue = iota
	// ContextObjectValue indicates we're parsing an object value
	ContextObjectValue
	// ContextArray indicates we're parsing an array
	ContextArray
	// ContextRoot indicates we're at the root level
	ContextRoot
)

// String representation of ContextValue for debugging
func (c ContextValue) String() string {
	switch c {
	case ContextObjectKey:
		return "object_key"
	case ContextObjectValue:
		return "object_value"
	case ContextArray:
		return "array"
	case ContextRoot:
		return "root"
	default:
		return "unknown"
	}
}

// LogEntry represents a single log entry for debugging
// Used when logging is enabled in the parser
type LogEntry struct {
	// Text contains the log message
	Text string
	// Context contains the current parsing context when the log was created
	Context string
}

// RepairOptions contains configuration options for JSON repair
// Uses functional options pattern for flexible configuration
type RepairOptions struct {
	// ReturnObjects determines whether to return objects or keep as maps
	ReturnObjects bool
	// SkipJSONLoads skips standard JSON unmarshaling fallback
	SkipJSONLoads bool
	// Logging enables debug logging
	Logging bool
	// StreamStable enables stream stability mode
	StreamStable bool
	// Strict enables strict mode (less aggressive repair)
	Strict bool
	// Schema contains optional JSON schema for guided repair
	Schema interface{}
	// Indent specifies indentation for output formatting
	Indent int
	// EnsureASCII ensures all non-ASCII characters are escaped
	EnsureASCII bool
	// RecordPositions tracks position information (for error reporting)
	RecordPositions bool
	// MaxDepth limits the maximum nesting depth
	MaxDepth int
}

// DefaultRepairOptions returns a RepairOptions with sensible defaults
func DefaultRepairOptions() *RepairOptions {
	return &RepairOptions{
		ReturnObjects:   false,
		SkipJSONLoads:   false,
		Logging:         false,
		StreamStable:    false,
		Strict:          false,
		Indent:          0,
		EnsureASCII:     false,
		RecordPositions: false,
		MaxDepth:        1000,
	}
}

// ParserOption is a functional option for configuring the Parser
type ParserOption func(*RepairOptions)

// WithLogging enables debug logging
func WithLogging() ParserOption {
	return func(o *RepairOptions) {
		o.Logging = true
	}
}

// WithStreamStable enables stream stability mode
func WithStreamStable() ParserOption {
	return func(o *RepairOptions) {
		o.StreamStable = true
	}
}

// WithStrict enables strict mode
func WithStrict() ParserOption {
	return func(o *RepairOptions) {
		o.Strict = true
	}
}

// WithSchema sets the JSON schema for guided repair
func WithSchema(schema interface{}) ParserOption {
	return func(o *RepairOptions) {
		o.Schema = schema
	}
}

// WithIndent sets the indentation level for pretty printing
func WithIndent(indent int) ParserOption {
	return func(o *RepairOptions) {
		o.Indent = indent
	}
}

// WithEnsureASCII enables ASCII-only output
func WithEnsureASCII() ParserOption {
	return func(o *RepairOptions) {
		o.EnsureASCII = true
	}
}

// WithMaxDepth sets the maximum nesting depth
func WithMaxDepth(depth int) ParserOption {
	return func(o *RepairOptions) {
		o.MaxDepth = depth
	}
}

// ParseError represents a parsing error with context
type ParseError struct {
	// Message describes the error
	Message string
	// Position is the character position where the error occurred
	Position int
	// Context is the parsing context when the error occurred
	Context string
}

// Error implements the error interface
func (e *ParseError) Error() string {
	if e.Context != "" {
		return e.Message + " at position " + string(rune(e.Position)) + " in " + e.Context
	}
	return e.Message + " at position " + string(rune(e.Position))
}
