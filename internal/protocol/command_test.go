package protocol

import (
	"errors"
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    Command
		wantErr error
	}{
		// valid lines
		{"command with args", "TAKE Healing Herbs", Command{"TAKE", "Healing Herbs"}, nil},
		{"lowercase command", "take Healing Herbs", Command{"TAKE", "Healing Herbs"}, nil},
		{"command alone", "look", Command{"LOOK", ""}, nil},
		{"args keep their case", "Connect Alice", Command{"CONNECT", "Alice"}, nil},
		{"unknown command is not an error", "foo bar", Command{"FOO", "bar"}, nil},
		{"accents in args", "CHAT GLOBAL salut à tous", Command{"CHAT", "GLOBAL salut à tous"}, nil},

		// spaces: trimmed around, kept inside
		{"spaces around the line", "  LOOK  ", Command{"LOOK", ""}, nil},
		{"spaces between command and args", "TAKE   Healing Herbs", Command{"TAKE", "Healing Herbs"}, nil},
		{"spaces inside args kept", "CHAT ROOM salut   toi", Command{"CHAT", "ROOM salut   toi"}, nil},

		// a final \r is removed before ValidateLine
		{"final carriage return", "LOOK\r", Command{"LOOK", ""}, nil},
		{"final carriage return after args", "TAKE Healing Herbs\r", Command{"TAKE", "Healing Herbs"}, nil},

		// errors: the Command must be empty
		{"empty line", "", Command{}, ErrEmptyLine},
		{"only spaces", "   ", Command{}, ErrEmptyLine},
		{"only carriage return", "\r", Command{}, ErrEmptyLine},
		{"carriage return inside", "LO\rOK", Command{}, ErrControlChar},
		{"tab", "LOOK\t", Command{}, ErrControlChar},
		{"terminal escape", "CHAT GLOBAL \x1b[2J", Command{}, ErrControlChar},
		{"invalid UTF-8", "LOOK \xff", Command{}, ErrInvalidUTF8},
		{"too long", "TAKE " + strings.Repeat("a", MaxLineLen), Command{}, ErrLineTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommand(tt.line)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseCommand(%q) error = %v, want %v", tt.line, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseCommand(%q) = %+v, want %+v", tt.line, got, tt.want)
			}
		})
	}
}
