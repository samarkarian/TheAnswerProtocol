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

func TestParseChatArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        string
		wantScope   string
		wantMessage string
		wantErr     error
	}{
		{"global", "GLOBAL hello", ScopeGlobal, "hello", nil},
		{"room lowercase", "room hi", ScopeRoom, "hi", nil},
		{"group mixed case", "Group yo", ScopeGroup, "yo", nil},

		{"accents and emoji", "GLOBAL ça va ? 🎉", ScopeGlobal, "ça va ? 🎉", nil},
		{"inner spaces kept, edges trimmed", "GLOBAL   salut   à tous  ", ScopeGlobal, "salut   à tous", nil},
		{"protocol words not interpreted", "ROOM EVT GLOBAL CHAT bob", ScopeRoom, "EVT GLOBAL CHAT bob", nil},

		{"stats is not a chat scope", "STATS hi", "", "", ErrMalformedCommand},
		{"unknown scope", "PARTOUT hi", "", "", ErrMalformedCommand},
		{"empty scope", " hi", "", "", ErrMalformedCommand},

		{"no message", "GLOBAL", "", "", ErrMalformedCommand},
		{"empty message", "GLOBAL ", "", "", ErrMalformedCommand},
		{"blank message", "GLOBAL    ", "", "", ErrMalformedCommand},
		{"empty args", "", "", "", ErrMalformedCommand},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, message, err := ParseChatArgs(tt.args)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ParseChatArgs(%q) error = %v, want %v", tt.args, err, tt.wantErr)
			}
			if scope != tt.wantScope || message != tt.wantMessage {
				t.Errorf("ParseChatArgs(%q) = (%q, %q), want (%q, %q)",
					tt.args, scope, message, tt.wantScope, tt.wantMessage)
			}
		})
	}
}

func TestParseGroupArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantSub  string
		wantUser string
		wantErrs []error // toutes les erreurs que errors.Is doit reconnaître
	}{
		{"create", "CREATE", GroupCreate, "", nil},
		{"leave lowercase", "leave", GroupLeave, "", nil},
		{"create trailing spaces", "CREATE   ", GroupCreate, "", nil},

		{"invite", "INVITE bob", GroupInvite, "bob", nil},
		{"join mixed case, name kept", "join Alice", GroupJoin, "Alice", nil},
		{"spaces around name trimmed", "INVITE   bob  ", GroupInvite, "bob", nil},
		{"accented name", "INVITE élodie", GroupInvite, "élodie", nil},

		{"invite without name", "INVITE", "", "", []error{ErrMalformedCommand, ErrInvalidUsername}},
		{"invite empty name", "INVITE ", "", "", []error{ErrMalformedCommand, ErrInvalidUsername}},
		{"join blank name", "JOIN    ", "", "", []error{ErrMalformedCommand, ErrInvalidUsername}},
		{"name with space", "JOIN alice bob", "", "", []error{ErrMalformedCommand, ErrInvalidUsername}},

		{"create with argument", "CREATE now", "", "", []error{ErrMalformedCommand}},
		{"leave with argument", "LEAVE alice", "", "", []error{ErrMalformedCommand}},

		{"unknown sub-command", "KICK bob", "", "", []error{ErrMalformedCommand}},
		{"empty args", "", "", "", []error{ErrMalformedCommand}},
		{"leading space", " INVITE bob", "", "", []error{ErrMalformedCommand}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, user, err := ParseGroupArgs(tt.args)

			if len(tt.wantErrs) == 0 && err != nil {
				t.Fatalf("ParseGroupArgs(%q) unexpected error: %v", tt.args, err)
			}
			if len(tt.wantErrs) > 0 && err == nil {
				t.Fatalf("ParseGroupArgs(%q) error = nil, want %v", tt.args, tt.wantErrs)
			}
			for _, want := range tt.wantErrs {
				if !errors.Is(err, want) {
					t.Errorf("ParseGroupArgs(%q) error = %v, want it to match %v", tt.args, err, want)
				}
			}

			if sub != tt.wantSub || user != tt.wantUser {
				t.Errorf("ParseGroupArgs(%q) = (%q, %q), want (%q, %q)",
					tt.args, sub, user, tt.wantSub, tt.wantUser)
			}
		})
	}
}
