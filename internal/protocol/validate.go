package protocol

import (
	"errors"
	"unicode"
	"unicode/utf8"
)

// Errors handled in ValidateLine
var (
	ErrEmptyLine   = errors.New("protocol: line is empty")
	ErrLineTooLong = errors.New("protocol: line too long")
	ErrInvalidUTF8 = errors.New("protocol: invalid UTF-8")
	ErrControlChar = errors.New("protocol: control character found")
)

// ValidateLine checks that line can be parsed safely:
// it must be non-empty, at most MaxLineLen bytes, valid UTF-8
// and without control characters.
// Trailing "\n" or "\r\n" must be removed before calling
func ValidateLine(line string) error {
	if len(line) == 0 {
		return ErrEmptyLine
	}
	if len(line) > MaxLineLen {
		return ErrLineTooLong
	}
	if !utf8.ValidString(line) {
		return ErrInvalidUTF8
	}
	for _, r := range line {
		if unicode.IsControl(r) {
			return ErrControlChar
		}
	}
	return nil
}
