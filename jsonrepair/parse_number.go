package jsonrepair

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// parseNumber parses a JSON number, handling integers, floats, and scientific notation
func (p *Parser) parseNumber() (JSONValue, error) {
	p.log("Parsing number")

	if p.index >= len(p.input) {
		return nil, errors.New("unexpected end of input")
	}

	startPos := p.index

	// Check if it starts with a minus sign
	if p.input[p.index] == '-' {
		p.index++
	}

	// Parse the integer part
	if err := p.parseIntegerPart(); err != nil {
		p.index = startPos
		return nil, err
	}

	// Check for decimal point
	isFloat := false
	if p.index < len(p.input) && p.input[p.index] == '.' {
		isFloat = true
		p.index++

		// Parse the fractional part
		if err := p.parseFractionalPart(); err != nil {
			p.index = startPos
			return nil, err
		}
	}

	// Check for exponent
	if p.index < len(p.input) && (p.input[p.index] == 'e' || p.input[p.index] == 'E') {
		isFloat = true
		p.index++

		// Parse the exponent part
		if err := p.parseExponentPart(); err != nil {
			p.index = startPos
			return nil, err
		}
	}

	// Parse the number string
	numStr := p.input[startPos:p.index]

	// Clean up and validate the number string
	numStr = p.cleanNumberString(numStr)

	// Parse as integer or float
	if isFloat {
		return p.parseFloat(numStr, startPos)
	}
	return p.parseInteger(numStr, startPos)
}

// parseIntegerPart parses the integer part of a number
func (p *Parser) parseIntegerPart() error {
	if p.index >= len(p.input) {
		return errors.New("expected digits in number")
	}

	// Special case: single zero
	if p.input[p.index] == '0' {
		p.index++

		// Check if followed by more digits (invalid in strict mode)
		if p.index < len(p.input) && unicode.IsDigit(rune(p.input[p.index])) {
			if p.strict {
				return errors.New("leading zeros not allowed in strict mode")
			}
			// In repair mode, skip leading zeros
			p.log("Repairing leading zeros in number")
			for p.index < len(p.input) && unicode.IsDigit(rune(p.input[p.index])) {
				p.index++
			}
		}
		return nil
	}

	// Parse digits
	if !unicode.IsDigit(rune(p.input[p.index])) {
		return errors.New("expected digit in number")
	}

	for p.index < len(p.input) && unicode.IsDigit(rune(p.input[p.index])) {
		p.index++
	}

	return nil
}

// parseFractionalPart parses the fractional part after a decimal point
func (p *Parser) parseFractionalPart() error {
	if p.index >= len(p.input) {
		if p.strict {
			return errors.New("expected digits after decimal point")
		}
		p.log("Repairing missing fractional part")
		return nil
	}

	// Parse digits after decimal point
	digitCount := 0
	for p.index < len(p.input) && unicode.IsDigit(rune(p.input[p.index])) {
		p.index++
		digitCount++
	}

	if digitCount == 0 && p.strict {
		return errors.New("expected digits after decimal point")
	}

	return nil
}

// parseExponentPart parses the exponent part of scientific notation
func (p *Parser) parseExponentPart() error {
	if p.index >= len(p.input) {
		if p.strict {
			return errors.New("incomplete exponent part")
		}
		p.log("Repairing incomplete exponent part")
		return nil
	}

	// Optional sign
	if p.input[p.index] == '+' || p.input[p.index] == '-' {
		p.index++
	}

	// Must have at least one digit
	if p.index >= len(p.input) || !unicode.IsDigit(rune(p.input[p.index])) {
		if p.strict {
			return errors.New("expected digit in exponent")
		}
		p.log("Repairing missing exponent digits")
		return nil
	}

	// Parse exponent digits
	for p.index < len(p.input) && unicode.IsDigit(rune(p.input[p.index])) {
		p.index++
	}

	return nil
}

// cleanNumberString cleans up common issues in number strings
func (p *Parser) cleanNumberString(numStr string) string {
	if p.strict {
		return numStr
	}

	// Remove any trailing non-numeric characters (in repair mode)
	repaired := numStr
	for i := len(repaired) - 1; i >= 0; i-- {
		ch := repaired[i]
		if ch == '-' || ch == '+' || ch == 'e' || ch == 'E' || ch == '.' ||
			(ch >= '0' && ch <= '9') {
			break
		}
		repaired = repaired[:i]
	}

	if repaired != numStr {
		p.log(fmt.Sprintf("Repaired number: %s -> %s", numStr, repaired))
	}

	return repaired
}

// parseFloat parses a string as a float64
func (p *Parser) parseFloat(numStr string, _ int) (JSONValue, error) {
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		if p.strict {
			return nil, fmt.Errorf("invalid number: %s", numStr)
		}

		// Try to repair the number
		repaired, ok := p.attemptNumberRepair(numStr)
		if !ok {
			return nil, fmt.Errorf("invalid number: %s", numStr)
		}

		val, err = strconv.ParseFloat(repaired, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", numStr)
		}

		p.log(fmt.Sprintf("Repaired number: %s -> %s", numStr, repaired))
	}

	return val, nil
}

// parseInteger parses a string as an integer
func (p *Parser) parseInteger(numStr string, _ int) (JSONValue, error) {
	// Special case for negative zero
	if numStr == "-0" {
		return 0, nil
	}

	val, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		// Try parsing as float and see if it's actually an integer
		floatVal, floatErr := strconv.ParseFloat(numStr, 64)
		if floatErr == nil && floatVal == float64(int64(floatVal)) {
			return int64(floatVal), nil
		}

		if p.strict {
			return nil, fmt.Errorf("invalid integer: %s", numStr)
		}

		// Try to repair
		repaired, ok := p.attemptNumberRepair(numStr)
		if !ok {
			return nil, fmt.Errorf("invalid integer: %s", numStr)
		}

		val, err = strconv.ParseInt(repaired, 10, 64)
		if err != nil {
			// If still can't parse as int, try as float
			floatVal, err2 := strconv.ParseFloat(repaired, 64)
			if err2 != nil {
				return nil, fmt.Errorf("invalid number: %s", numStr)
			}
			p.log(fmt.Sprintf("Parsed as float instead of int: %s", numStr))
			return floatVal, nil
		}

		p.log(fmt.Sprintf("Repaired integer: %s -> %s", numStr, repaired))
	}

	return val, nil
}

// attemptNumberRepair attempts to repair a malformed number string
func (p *Parser) attemptNumberRepair(numStr string) (string, bool) {
	var repaired strings.Builder
	dotCount := 0
	eCount := 0
	minusCount := 0
	plusCount := 0

	for i, ch := range numStr {
		switch ch {
		case '.':
			dotCount++
			if dotCount > 1 {
				// Skip extra dots
				continue
			}
			repaired.WriteRune(ch)
		case 'e', 'E':
			eCount++
			if eCount > 1 {
				// Skip extra exponents
				continue
			}
			repaired.WriteRune(ch)
		case '-':
			minusCount++
			// Only allow minus at start or after 'e'/'E'
			if i == 0 || (i > 0 && (numStr[i-1] == 'e' || numStr[i-1] == 'E')) {
				repaired.WriteRune(ch)
			}
		case '+':
			plusCount++
			// Only allow plus after 'e'/'E'
			if i > 0 && (numStr[i-1] == 'e' || numStr[i-1] == 'E') {
				repaired.WriteRune(ch)
			}
		default:
			if unicode.IsDigit(ch) || unicode.IsLetter(ch) {
				repaired.WriteRune(ch)
			}
			// Skip other invalid characters
		}
	}

	return repaired.String(), repaired.Len() > 0
}

// containsAny checks if a string contains any of the given characters
func containsAny(s string, chars string) bool {
	for _, ch := range chars {
		if strings.ContainsRune(s, ch) {
			return true
		}
	}
	return false
}

// isNumberChar checks if a character can appear in a JSON number
func isNumberChar(ch rune) bool {
	return unicode.IsDigit(ch) || ch == '-' || ch == '+' || ch == 'e' || ch == 'E' || ch == '.'
}
