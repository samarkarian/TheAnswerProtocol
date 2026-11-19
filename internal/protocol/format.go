package protocol

import (
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

	if strings.Contains(data, "\r") || strings.Contains(data, "\n") {
		return "", ValidateLine(data)
	}

	return fmt.Sprintf("OK %s\n", data), nil
}

func FormatOKJSON(v any) (string, error) {

	jsonFmt, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	s := string(jsonFmt)

	return fmt.Sprintf("OK %s\n", s), nil
}

func FormatEvent(cat, typ, payload string) (string, error) {

	slc := []string{cat, typ, payload}
	newSlc := []string{}

	for _, s := range slc {
		if s != "" {
			newSlc = append(newSlc, s)
		}
	}

	for _, s := range slc {
		if strings.Contains(s, "\r") || strings.Contains(s, "\n") {
			return "", ValidateLine(s)
		}
	}

	s := strings.Join(newSlc, " ")
	return fmt.Sprintf("EVT %s\n", s), nil
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
