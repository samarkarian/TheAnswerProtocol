package protocol

import (
	"strings"
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
