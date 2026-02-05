# json-repair-go

> Fully authored by Kimi K2.5

[![Go Reference](https://pkg.go.dev/badge/github.com/Crescent617/json-repair-go.svg)](https://pkg.go.dev/github.com/Crescent617/json-repair-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Crescent617/json-repair-go)](https://goreportcard.com/report/github.com/Crescent617/json-repair-go)

A Go port of the Python json_repair library that fixes malformed JSON and returns valid, parseable JSON.

## Features

- **Repair Malformed JSON**: Automatically fixes common JSON syntax errors
- **Multiple Quote Support**: Handles single quotes, double quotes, and Unicode smart quotes
- **Comment Removal**: Strips JavaScript-style comments (`//`, `/* */`, `#`)
- **Trailing Comma Fix**: Removes trailing commas in objects and arrays
- **Missing Quote Repair**: Adds missing quotes around object keys and string values
- **Scientific Notation**: Properly handles numbers in scientific notation
- **Unicode Support**: Full Unicode escape sequence support including surrogate pairs
- **Unclosed Structure Repair**: Handles unclosed arrays, objects, and strings
- **Type-Strict**: Uses Go's type system with `interface{}` for flexible JSON values
- **Pure Go**: No external dependencies, uses only the Go standard library
- **High Performance**: Optimized for speed and memory efficiency

## Installation

```bash
go get github.com/Crescent617/json-repair-go
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/Crescent617/json-repair-go/pkg/jsonrepair"
)

func main() {
    // Repair malformed JSON
    input := `{name: "John", age: 30,}` // Missing quotes and trailing comma

    result, err := jsonrepair.RepairJSON(input)
    if err != nil {
        panic(err)
    }

    fmt.Println(result)
    // Output: {"age":30,"name":"John"}
}
```

### Parse to Go Values

```go
// Repair and get Go value
value, err := jsonrepair.RepairJSONToValue(`{active: true, scores: [95, 87, 92]}`)
if err != nil {
    panic(err)
}

// Access values
data := value.(map[string]interface{})
fmt.Println(data["active"])  // true
fmt.Println(data["scores"])  // [95, 87, 92]
```

### Working with Files

```go
// Read and repair from file
value, err := jsonrepair.FromFile("data.json")
if err != nil {
    panic(err)
}

// Repair and save to new file
err = jsonrepair.RepairFile("input.json", "output.json")
if err != nil {
    panic(err)
}
```

### Advanced Options

```go
// Use options for configuration
result, err := jsonrepair.RepairJSON(
    input,
    jsonrepair.WithIndent(2),           // Pretty print with 2 spaces
    jsonrepair.WithLogging(),           // Enable debug logging
    jsonrepair.WithStrict(),            // Strict mode (less aggressive repair)
    jsonrepair.WithEnsureASCII(),       // Escape non-ASCII characters
)
```

## API Reference

### Core Functions

#### `RepairJSON(input string, opts ...ParserOption) (string, error)`
Repairs malformed JSON and returns a valid JSON string.

```go
result, err := jsonrepair.RepairJSON(`{name: "John"}`)
// Returns: `{"name":"John"}`
```

#### `RepairJSONToValue(input string, opts ...ParserOption) (JSONValue, error)`
Repairs JSON and returns it as a Go value (map, slice, string, number, bool, or nil).

```go
value, err := jsonrepair.RepairJSONToValue(`[1, 2, 3]`)
// Returns: []interface{}{1.0, 2.0, 3.0}
```

#### `Loads(input string, opts ...ParserOption) (JSONValue, error)`
Alias for `RepairJSONToValue`.

#### `Load(reader io.Reader, opts ...ParserOption) (JSONValue, error)`
Repairs JSON from an `io.Reader`.

#### `FromFile(filename string, opts ...ParserOption) (JSONValue, error)`
Repairs JSON from a file.

#### `RepairFile(inputPath, outputPath string, opts ...ParserOption) error`
Repairs JSON from a file and writes to another file.

### Configuration Options

#### `WithLogging()`
Enables debug logging (stored in parser logger).

#### `WithStrict()`
Enables strict mode (less aggressive repair).

#### `WithStreamStable()`
Enables stream stability mode for incremental parsing.

#### `WithIndent(n int)`
Pretty-prints output with n spaces of indentation.

#### `WithEnsureASCII()`
Escapes all non-ASCII characters in output.

#### `WithMaxDepth(depth int)`
Sets maximum nesting depth (default: 1000).

#### `WithSchema(schema interface{})`
Sets JSON schema for schema-guided repair (experimental).

## Repair Capabilities

### What Gets Repaired

1. **Missing Quotes Around Keys**: `{key: "value"}` → `{"key": "value"}`
2. **Single Quotes**: `{'key': 'value'}` → `{"key": "value"}`
3. **Smart Quotes**: `{“key”: “value”}` → `{"key": "value"}`
4. **Trailing Commas**: `[1, 2,]` → `[1, 2]`
5. **Unclosed Structures**: `{"key": "value"` → `{"key": "value"}`
6. **Comments**: `// comment\n{"a": 1}` → `{"a": 1}`
7. **Hash Comments**: `# comment\n{"a": 1}` → `{"a": 1}`
8. **Block Comments**: `/* comment */{"a": 1}` → `{"a": 1}`
9. **Missing Commas**: `{a: 1 b: 2}` → `{"a": 1, "b": 2}`
10. **Unquoted Strings**: `[hello, world]` → `["hello", "world"]`

### Examples

```go
// Input with multiple issues
input := `
// User data
{
    name: "Alice",        // unquoted key
    age: 30,              // trailing comma
    active: true,         // boolean
    scores: [95, 87, 92], // array
    settings: {           // nested object
        theme: 'dark',    // single quotes
        notifications: false
    },                    // trailing comma
}
`

result, _ := jsonrepair.RepairJSON(input, jsonrepair.WithIndent(2))
fmt.Println(result)
// Output:
// {
//   "age": 30,
//   "name": "Alice",
//   "active": true,
//   "scores": [95, 87, 92],
//   "settings": {
//     "theme": "dark",
//     "notifications": false
//   }
// }
```

## Error Handling

The library provides detailed error messages for parsing failures:

```go
input := `{"key": undefined}`
result, err := jsonrepair.RepairJSON(input)
if err != nil {
    // Check if it's a parsing error
    if parseErr, ok := err.(*jsonrepair.ParseError); ok {
        fmt.Printf("Parse error at position %d in context %s: %s\n",
            parseErr.Position, parseErr.Context, parseErr.Message)
    }
}
```

## Architecture

### Project Structure

```
json-repair-go/
├── pkg/
│   └── jsonrepair/
│       ├── types.go          # Core types and configuration
│       ├── constants.go      # Constants and delimiters
│       ├── parser.go         # Main parser logic
│       ├── parse_string.go   # String parsing
│       ├── parse_number.go   # Number parsing
│       ├── parse_array.go    # Array parsing
│       ├── parse_object.go   # Object parsing
│       ├── parse_comment.go  # Comment removal
│       ├── schema_repair.go  # Schema-guided repair (experimental)
│       └── jsonrepair.go     # Public API
├── test/                     # Test files
├── examples/                 # Example programs
└── cmd/
    └── jsonrepair/           # CLI tool (WIP)
```

### Design Principles

1. **Pure Functionality**: No side effects or hidden state
2. **Immutable**: Original input is never modified
3. **Type-Safe**: Leverages Go's type system
4. **Composable**: Easy to integrate into larger systems
5. **Efficient**: Streaming parser with minimal allocations
6. **Repair-First**: Always attempts repair before giving up

## Limitations

- Schema validation is currently experimental (Phase 4 incomplete)
- Some deeply nested or complex malformed JSON may not be repairable
- Stream stability mode is not yet fully optimized
- Performance may degrade with very large files (>100MB)

## Development

### Running Tests

```bash
go test ./... -v
```

### Running Lint

The project uses [golangci-lint](https://golangci-lint.run/) for comprehensive linting.

```bash
# Install golangci-lint
make install-lint

# Run lint
make lint

# Run lint with auto-fix
make lint-fix

# Run all checks (fmt, vet, lint, test)
make check
```

### Using Make

Common development commands:

```bash
make help      # Show all available commands
make lint      # Run linter
make test      # Run tests
make build     # Build the project
make fmt       # Format code
make vet       # Run go vet
make clean     # Clean build artifacts
make tidy      # Tidy go modules
```

### Benchmarks

```bash
go test -bench=. ./test
```

### Code Coverage

```bash
go test -cover ./...

# Or with make (generates HTML report)
make test-coverage
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

Areas for contribution:
- Complete schema validation implementation
- Performance optimizations
- Additional test cases
- CLI tool completion
- Documentation improvements

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by the [Python json_repair library](https://github.com/mangiucugna/json_repair)
- Built with Go's excellent standard library
- Thanks to all contributors

## Version History

### v0.1.0 (Current)
- Initial release
- Core JSON repair functionality
- Support for common JSON issues
- Basic test suite
- Public API implementation

### Planned Features (v0.2.0)
- Complete schema validation
- Enhanced CLI tool
- Performance optimizations
- Additional repair strategies
- More comprehensive test coverage (80%+)
