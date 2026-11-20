package protocol

import (
	"errors"
	"strings"
)

var (
	// ErrMalformedCommand is returned when the arguments of a command do not match its format.
	ErrMalformedCommand = errors.New("protocol: malformed command")
)

type Command struct {
	Name string
	Args string
}

func ParseCommand(line string) (Command, error) {

	command := Command{}

	line = strings.TrimSuffix(line, "\r")

	if strings.TrimSpace(line) == "" {
		return command, ErrEmptyLine
	}
	err := ValidateLine(line)
	if err != nil {
		return command, err
	}

	cmd, args, _ := strings.Cut(strings.TrimSpace(line), " ")
	cmd = strings.ToUpper(cmd)
	args = strings.TrimSpace(args)

	command = Command{
		Name: cmd,
		Args: args,
	}

	return command, nil
}

// ParseChatArgs splits the arguments of a CHAT command, such as "global hello",
// into an upper-case scope and a message with surrounding spaces removed.
// It returns ErrMalformedCommand if the scope is not GLOBAL, ROOM or GROUP,
// or if the message is blank.
func ParseChatArgs(cmdArgs string) (string, string, error) {
	scope, message, _ := strings.Cut(cmdArgs, " ")
	message = strings.TrimSpace(message)
	if message == "" {
		return "", "", ErrMalformedCommand
	}
	scope = strings.ToUpper(scope)
	switch scope {
	case ScopeGlobal, ScopeRoom, ScopeGroup:
		return scope, message, nil
	default:
		return "", "", ErrMalformedCommand
	}
}
