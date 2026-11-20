package protocol

import (
	"errors"
	"unicode"
	"unicode/utf8"
)

// Errors handled in ValidateLine and ValidateUsername
var (
	ErrEmptyLine       = errors.New("protocol: line is empty")
	ErrLineTooLong     = errors.New("protocol: line too long")
	ErrInvalidUTF8     = errors.New("protocol: invalid UTF-8")
	ErrControlChar     = errors.New("protocol: control character found")
	ErrInvalidUsername = errors.New("protocol: invalid username")
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

// ValidateUsername returns ErrInvalidUsername if name is empty or contains a space.
func ValidateUsername(name string) error {
	if name == "" {
		return ErrInvalidUsername
	}
	for _, r := range name {
		if unicode.IsSpace(r) {
			return ErrInvalidUsername
		}
	}
	return nil
}
