package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidCommand = errors.New("protocol: invalid command name")

func FormatOK(data string) (string, error) {

	if data == "" {
		return "OK\n", nil
	}

	s := fmt.Sprintf("OK %s", data)
	err := ValidateLine(s)
	if err != nil {
		return "", err
	}

	return s + "\n", nil
}

func FormatOKJSON(v any) (string, error) {

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(v)
	if err != nil {
		return "", err
	}

	s := "OK " + strings.TrimSuffix(buffer.String(), "\n")
	err = ValidateLine(s)
	if err != nil {
		return "", err
	}

	return s + "\n", nil
}

func FormatErr(e ErrorCode) string {

	return fmt.Sprintf("ERR %d %s\n", e.Code, e.Message)
}

func FormatEvent(cat, typ, payload string) (string, error) {

	slc := []string{cat, typ, payload}
	newSlc := []string{}

	if cat == "" || payload == "" {
		return "", ErrMalformedEvent
	}
	if cat == ScopeStats && typ != "" {
		return "", ErrMalformedEvent
	}
	if cat != ScopeStats && typ == "" {
		return "", ErrMalformedEvent
	}
	if strings.Contains(cat, " ") || strings.Contains(typ, " ") {
		return "", ErrMalformedEvent
	}

	for _, s := range slc {
		if s != "" {
			newSlc = append(newSlc, s)
		}
	}

	s := strings.Join(newSlc, " ")
	s = fmt.Sprintf("EVT %s", s)
	err := ValidateLine(s)
	if err != nil {
		return "", err
	}

	return s + "\n", nil
}

// FormatCommand builds the line a client sends: the command name in upper case,
// followed by args if any, and a trailing "\n". Surrounding spaces in args are trimmed.
// It returns ErrInvalidCommand if name is not made of ASCII letters only,
// or the error from ValidateLine.
func FormatCommand(name, args string) (string, error) {
	if name == "" {
		return "", ErrInvalidCommand
	}

	// RFC 42TAP: command-name = 1*ALPHA
	for _, r := range name {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return "", ErrInvalidCommand
		}
	}

	command := strings.ToUpper(name)
	args = strings.TrimSpace(args)
	if args != "" {
		command += " " + args
	}

	if err := ValidateLine(command); err != nil {
		return "", err
	}
	return command + "\n", nil
}
