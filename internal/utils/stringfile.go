package utils

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// StringFileReader provides a unified interface for reading from strings or files
type StringFileReader struct {
	reader io.Reader
	closer io.Closer
	buffer *bufio.Reader
}

// NewStringReader creates a reader from a string
func NewStringReader(s string) *StringFileReader {
	reader := strings.NewReader(s)
	return &StringFileReader{
		reader: reader,
		buffer: bufio.NewReader(reader),
	}
}

// NewFileReader creates a reader from a file
func NewFileReader(filename string) (*StringFileReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return &StringFileReader{
		reader: file,
		closer: file,
		buffer: bufio.NewReader(file),
	}, nil
}

// ReadString reads until the first occurrence of delim in the input
func (sfr *StringFileReader) ReadString(delim byte) (string, error) {
	return sfr.buffer.ReadString(delim)
}

// ReadBytes reads until the first occurrence of delim in the input
func (sfr *StringFileReader) ReadBytes(delim byte) ([]byte, error) {
	return sfr.buffer.ReadBytes(delim)
}

// Read reads data into p
func (sfr *StringFileReader) Read(p []byte) (n int, err error) {
	return sfr.buffer.Read(p)
}

// Close closes the underlying file if it exists
func (sfr *StringFileReader) Close() error {
	if sfr.closer != nil {
		return sfr.closer.Close()
	}
	return nil
}

// Reset resets the buffer state
func (sfr *StringFileReader) Reset(r io.Reader) {
	sfr.buffer.Reset(r)
	sfr.reader = r
}

// ReadAll reads all remaining data
func (sfr *StringFileReader) ReadAll() ([]byte, error) {
	return io.ReadAll(sfr.buffer)
}

// ReadRune reads a single UTF-8 encoded Unicode character
func (sfr *StringFileReader) ReadRune() (rune, int, error) {
	return sfr.buffer.ReadRune()
}

// Buffered returns the number of bytes that can be read from the current buffer
func (sfr *StringFileReader) Buffered() int {
	return sfr.buffer.Buffered()
}
