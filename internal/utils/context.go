package utils

import "strings"

// ContextTracker provides utilities for tracking parsing context
// to enable better error reporting and repair decisions

type ContextTracker struct {
	contextStack []string
	currentLine  int
	currentCol   int
}

// NewContextTracker creates a new context tracker
func NewContextTracker() *ContextTracker {
	return &ContextTracker{
		contextStack: make([]string, 0),
		currentLine:  1,
		currentCol:   1,
	}
}

// PushContext adds a new context to the stack
func (ct *ContextTracker) PushContext(context string) {
	ct.contextStack = append(ct.contextStack, context)
}

// PopContext removes the most recent context from the stack
func (ct *ContextTracker) PopContext() {
	if len(ct.contextStack) > 0 {
		ct.contextStack = ct.contextStack[:len(ct.contextStack)-1]
	}
}

// GetCurrentContext returns the current context path
func (ct *ContextTracker) GetCurrentContext() string {
	if len(ct.contextStack) == 0 {
		return "root"
	}
	return ct.contextStack[len(ct.contextStack)-1]
}

// GetFullContextPath returns the full context path as a string
func (ct *ContextTracker) GetFullContextPath() string {
	if len(ct.contextStack) == 0 {
		return "root"
	}
	return joinContextPath(ct.contextStack)
}

// UpdatePosition updates the current position based on character input
func (ct *ContextTracker) UpdatePosition(ch rune) {
	if ch == '\n' {
		ct.currentLine++
		ct.currentCol = 1
	} else {
		ct.currentCol++
	}
}

// GetPosition returns the current line and column
func (ct *ContextTracker) GetPosition() (line, col int) {
	return ct.currentLine, ct.currentCol
}

// UpdatePositionFromString updates position based on a string
func (ct *ContextTracker) UpdatePositionFromString(s string) {
	for _, ch := range s {
		ct.UpdatePosition(ch)
	}
}

// joinContextPath joins context stack into a path string
func joinContextPath(stack []string) string {
	// This is a simple implementation - could be enhanced
	path := ""
	var pathSb75 strings.Builder
	for i, ctx := range stack {
		if i > 0 {
			pathSb75.WriteString(".")
		}
		pathSb75.WriteString(ctx)
	}
	path += pathSb75.String()
	return path
}

// ContextProvider defines interface for types that provide context
// Useful for schema-aware context tracking
type ContextProvider interface {
	GetContext() string
	GetPath() []string
}

// PathContext provides context with JSON path information
type PathContext struct {
	path []string
}

// NewPathContext creates a new path context
func NewPathContext() *PathContext {
	return &PathContext{
		path: make([]string, 0),
	}
}

// Push adds a path element
func (pc *PathContext) Push(element string) {
	pc.path = append(pc.path, element)
}

// Pop removes the last path element
func (pc *PathContext) Pop() {
	if len(pc.path) > 0 {
		pc.path = pc.path[:len(pc.path)-1]
	}
}

// Get returns the current context path
func (pc *PathContext) Get() []string {
	return pc.path
}

// String returns the path as a string
func (pc *PathContext) String() string {
	return joinContextPath(pc.path)
}

// Copy creates a copy of the path context
func (pc *PathContext) Copy() *PathContext {
	newPath := make([]string, len(pc.path))
	copy(newPath, pc.path)
	return &PathContext{path: newPath}
}
