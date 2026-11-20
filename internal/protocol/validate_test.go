package protocol

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want error
	}{
		// valid lines
		{"simple command", "LOOK", nil},
		{"command with args", "TAKE Healing Herbs", nil},
		{"accents", "CHAT GLOBAL salut à tous, ça va ?", nil},
		{"emoji", "CHAT ROOM gg 🎉", nil},
		{"empty line", "", ErrEmptyLine},

		// length test, check that length is counted as byte
		{"exactly max length", strings.Repeat("a", MaxLineLen), nil},
		{"one byte over max", strings.Repeat("a", MaxLineLen+1), ErrLineTooLong},
		{"accents at max (2 bytes each)", strings.Repeat("é", MaxLineLen/2), nil},
		{"accents one over max", strings.Repeat("é", MaxLineLen/2+1), ErrLineTooLong},

		// invalid UTF-8
		{"invalid byte", "LOOK \xff", ErrInvalidUTF8},
		{"truncated sequence", "CHAT ROOM caf\xc3", ErrInvalidUTF8},

		// control chars
		{"terminal escape", "CHAT GLOBAL \x1b[2J", ErrControlChar},
		{"null byte", "LOOK\x00", ErrControlChar},
		{"newline inside", "CHAT ROOM a\nb", ErrControlChar},
		{"carriage return left by caller", "LOOK\r", ErrControlChar},
		{"tab", "CHAT ROOM a\tb", ErrControlChar},
		{"delete char", "LOOK\x7f", ErrControlChar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLine(tt.line)
			if tt.want == nil {
				if err != nil {
					t.Errorf("got error %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Errorf("got error %v, want %v", err, tt.want)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{"simple", "alice", nil},
		{"with digits", "bob42", nil},
		{"with symbols", "x_x-x.x", nil},
		{"accents", "élodie", nil},
		{"emoji", "gg🎉", nil},

		{"empty", "", ErrInvalidUsername},
		{"space inside", "alice bob", ErrInvalidUsername},
		{"leading space", " alice", ErrInvalidUsername},
		{"trailing space", "alice ", ErrInvalidUsername},
		{"tab", "alice\tbob", ErrInvalidUsername},
		{"newline", "alice\nbob", ErrInvalidUsername},
		{"non-breaking space", "alice\u00a0bob", ErrInvalidUsername},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateUsername(%q) = %v, want %v", tt.username, err, tt.wantErr)
			}
		})
	}
}
