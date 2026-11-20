package protocol

import (
	"errors"
	"fmt"
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

// ParseGroupArgs splits the arguments of a GROUP command, such as "invite alice",
// into an upper-case sub-command and a username (empty for CREATE and LEAVE).
// It returns ErrMalformedCommand if the sub-command is not CREATE, INVITE, JOIN
// or LEAVE, if CREATE or LEAVE has an argument, or if INVITE or JOIN has no valid
// username; in that last case the error also matches ErrInvalidUsername.
func ParseGroupArgs(cmdArgs string) (string, string, error) {
	subCmd, name, _ := strings.Cut(cmdArgs, " ")
	subCmd = strings.ToUpper(subCmd)
	name = strings.TrimSpace(name)

	switch subCmd {
	case GroupCreate, GroupLeave:
		if name != "" {
			return "", "", ErrMalformedCommand
		}
		return subCmd, "", nil
	case GroupInvite, GroupJoin:
		if err := ValidateUsername(name); err != nil {
			return "", "", fmt.Errorf("%w: %w", ErrMalformedCommand, err)
		}
		return subCmd, name, nil
	default:
		return "", "", ErrMalformedCommand
	}
}
