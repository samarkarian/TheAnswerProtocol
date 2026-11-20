package protocol

import (
	"errors"
	"maps"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    Message
		wantErr error
	}{
		// OK
		{"ok with data", "OK connected", Message{Kind: KindOK, Data: "connected"}, nil},
		{"ok alone", "OK", Message{Kind: KindOK}, nil},
		{"ok key=value", "OK hello proto=1", Message{Kind: KindOK, Data: "hello proto=1"}, nil},
		{"ok json", `OK {"hp": 80, "max_hp": 100}`, Message{Kind: KindOK, Data: `{"hp": 80, "max_hp": 100}`}, nil},
		{"ok free text", "OK Welcome to my bakery!", Message{Kind: KindOK, Data: "Welcome to my bakery!"}, nil},

		// end line
		{"trailing LF", "OK connected\n", Message{Kind: KindOK, Data: "connected"}, nil},
		{"trailing CRLF", "OK connected\r\n", Message{Kind: KindOK, Data: "connected"}, nil},
		{"trailing CR", "OK connected\r", Message{Kind: KindOK, Data: "connected"}, nil},

		// ERR
		{"err", "ERR 404 ITEM_NOT_FOUND", Message{Kind: KindErr, Code: 404, ErrMsg: "ITEM_NOT_FOUND"}, nil},
		{"err unknown code", "ERR 400 BAD_REQUEST", Message{Kind: KindErr, Code: 400, ErrMsg: "BAD_REQUEST"}, nil},
		{"err alone", "ERR", Message{}, ErrMalformedError},
		{"err without message", "ERR 404", Message{}, ErrMalformedError},
		{"err code not a number", "ERR abc NOPE", Message{}, ErrMalformedError},
		{"err code too short", "ERR 12 NOPE", Message{}, ErrMalformedError},
		{"err code too long", "ERR 1234 NOPE", Message{}, ErrMalformedError},
		{"err code with sign", "ERR +12 NOPE", Message{}, ErrMalformedError},
		{"err negative code", "ERR -12 NOPE", Message{}, ErrMalformedError},
		// we tolerate spaces before msg
		{"err with space before msg", "ERR 404  X", Message{Kind: KindErr, Code: 404, ErrMsg: " X"}, nil},

		// EVT
		{"evt presence", "EVT ROOM PRESENCE ENTER alice",
			Message{Kind: KindEvent, Category: "ROOM", Type: "PRESENCE", Payload: "ENTER alice"}, nil},
		{"evt chat with accents", "EVT GLOBAL CHAT bob salut à tous, ça va ?",
			Message{Kind: KindEvent, Category: "GLOBAL", Type: "CHAT", Payload: "bob salut à tous, ça va ?"}, nil},
		{"evt group invite", "EVT GROUP INVITE alice",
			Message{Kind: KindEvent, Category: "GROUP", Type: "INVITE", Payload: "alice"}, nil},
		{"evt stats has no type", "EVT STATS players=5",
			Message{Kind: KindEvent, Category: "STATS", Payload: "players=5"}, nil},
		{"evt unknown category", "EVT COMBAT HIT goblin 12",
			Message{Kind: KindEvent, Category: "COMBAT", Type: "HIT", Payload: "goblin 12"}, nil},
		{"evt alone", "EVT", Message{}, ErrMalformedEvent},
		{"evt without type", "EVT ROOM", Message{}, ErrMalformedEvent},
		// as ERR, we accept spaces before payload
		{"evt stats multiple spaces before", "EVT STATS  players=5",
			Message{Kind: KindEvent, Category: "STATS", Payload: " players=5"}, nil},

		// unknown first word
		{"client command", "LOOK", Message{}, ErrUnknownMessage},
		{"lowercase ok", "ok connected", Message{}, ErrUnknownMessage},
		{"ok glued to data", "OKconnected", Message{}, ErrUnknownMessage},

		// ValidateLine errors
		{"empty line", "", Message{}, ErrEmptyLine},
		{"invalid utf8", "OK \xff", Message{}, ErrInvalidUTF8},
		{"control char", "OK \x1b[2J", Message{}, ErrControlChar},

		// blank fields
		{"ok trailing space", "OK ", Message{Kind: KindOK}, nil},
		{"err empty message", "ERR 404 ", Message{}, ErrMalformedError},
		{"evt empty category", "EVT  ROOM CHAT x", Message{}, ErrMalformedEvent},
		{"evt empty type", "EVT ROOM  CHAT x", Message{}, ErrMalformedEvent},
		{"evt empty payload", "EVT ROOM CHAT ", Message{}, ErrMalformedEvent},
		{"evt stats empty payload", "EVT STATS ", Message{}, ErrMalformedEvent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLine(tt.line)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseLine(%q) error = %v, want %v", tt.line, err, tt.wantErr)
				}
				if got != (Message{}) {
					t.Errorf("ParseLine(%q) = %+v, want an empty Message on error", tt.line, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseLine(%q) unexpected error: %v", tt.line, err)
			}
			if got != tt.want {
				t.Errorf("ParseLine(%q)\n got  %+v\n want %+v", tt.line, got, tt.want)
			}
		})
	}
}

func TestParseKV(t *testing.T) {
	tests := []struct {
		name string
		data string
		want map[string]string
	}{
		{"greeting", "hello proto=1", map[string]string{"proto": "1"}},
		{"move", "room=loc.bakery", map[string]string{"room": "loc.bakery"}},
		{"take", "taken=item.herbs", map[string]string{"taken": "item.herbs"}},
		{"who / stats", "players=5", map[string]string{"players": "5"}},
		{"group", "group=g1", map[string]string{"group": "g1"}},

		{"two pairs", "a=1 b=2", map[string]string{"a": "1", "b": "2"}},
		{"surrounding spaces", "  room=loc.bakery   ", map[string]string{"room": "loc.bakery"}},
		{"multiple spaces between", "a=1    b=2", map[string]string{"a": "1", "b": "2"}},

		{"empty data", "", map[string]string{}},
		{"single word", "connected", map[string]string{}},
		{"words mixed with pairs", "hello world proto=1", map[string]string{"proto": "1"}},
		{"json is not key=value", `{"hp": 80}`, map[string]string{}},

		{"value containing =", "a=b=c", map[string]string{"a": "b=c"}},
		{"case preserved", "Room=X", map[string]string{"Room": "X"}},
		{"accents in value", "name=élodie", map[string]string{"name": "élodie"}},

		{"empty key ignored", "=x", map[string]string{}},
		{"lone equal sign", "=", map[string]string{}},
		{"empty value kept", "room=", map[string]string{"room": ""}},
		{"duplicate key, last wins", "a=1 a=2", map[string]string{"a": "2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseKV(tt.data)
			if !maps.Equal(got, tt.want) {
				t.Errorf("ParseKV(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestParsePresence(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		wantAction string
		wantName   string
		wantErrs   []error
	}{
		{"enter", "ENTER alice", PresenceEnter, "alice", nil},
		{"leave", "LEAVE bob", PresenceLeave, "bob", nil},
		{"accented name", "ENTER élodie", PresenceEnter, "élodie", nil},

		{"no name", "ENTER", "", "", []error{ErrMalformedEvent, ErrInvalidUsername}},
		{"empty name", "ENTER ", "", "", []error{ErrMalformedEvent, ErrInvalidUsername}},
		{"empty payload", "", "", "", []error{ErrMalformedEvent, ErrInvalidUsername}},
		{"name with space", "ENTER alice bob", "", "", []error{ErrMalformedEvent, ErrInvalidUsername}},
		{"double space", "ENTER  alice", "", "", []error{ErrMalformedEvent, ErrInvalidUsername}},

		{"unknown action", "JUMP alice", "", "", []error{ErrMalformedEvent}},
		{"lowercase action", "enter alice", "", "", []error{ErrMalformedEvent}},
		{"empty action", " alice", "", "", []error{ErrMalformedEvent}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, name, err := ParsePresence(tt.payload)

			if len(tt.wantErrs) == 0 && err != nil {
				t.Fatalf("ParsePresence(%q) unexpected error: %v", tt.payload, err)
			}
			if len(tt.wantErrs) > 0 && err == nil {
				t.Fatalf("ParsePresence(%q) error = nil, want %v", tt.payload, tt.wantErrs)
			}
			for _, want := range tt.wantErrs {
				if !errors.Is(err, want) {
					t.Errorf("ParsePresence(%q) error = %v, want it to match %v", tt.payload, err, want)
				}
			}

			if action != tt.wantAction || name != tt.wantName {
				t.Errorf("ParsePresence(%q) = (%q, %q), want (%q, %q)",
					tt.payload, action, name, tt.wantAction, tt.wantName)
			}
		})
	}
}

func TestParseChat(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		wantUser    string
		wantMessage string
		wantErrs    []error
	}{
		{"simple", "alice hey guys", "alice", "hey guys", nil},
		{"one word message", "bob hi", "bob", "hi", nil},
		{"accents and emoji", "élodie ça va ? 🎉", "élodie", "ça va ? 🎉", nil},
		{"inner spaces kept", "alice a   b    c", "alice", "a   b    c", nil},
		{"leading spaces kept", "alice   ascii art", "alice", "  ascii art", nil},
		{"trailing spaces kept", "alice hi  ", "alice", "hi  ", nil},
		{"protocol words not interpreted", "alice ENTER room=loc.x EVT", "alice", "ENTER room=loc.x EVT", nil},

		{"empty payload", "", "", "", []error{ErrMalformedEvent, ErrInvalidUsername}},
		{"empty username", " alice hi", "", "", []error{ErrMalformedEvent, ErrInvalidUsername}},

		{"no message", "alice", "", "", []error{ErrMalformedEvent}},
		{"empty message", "alice ", "", "", []error{ErrMalformedEvent}},
		{"only spaces", "alice    ", "", "", []error{ErrMalformedEvent}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, message, err := ParseChat(tt.payload)

			if len(tt.wantErrs) == 0 && err != nil {
				t.Fatalf("ParseChat(%q) unexpected error: %v", tt.payload, err)
			}
			if len(tt.wantErrs) > 0 && err == nil {
				t.Fatalf("ParseChat(%q) error = nil, want %v", tt.payload, tt.wantErrs)
			}
			for _, want := range tt.wantErrs {
				if !errors.Is(err, want) {
					t.Errorf("ParseChat(%q) error = %v, want it to match %v", tt.payload, err, want)
				}
			}

			if user != tt.wantUser || message != tt.wantMessage {
				t.Errorf("ParseChat(%q) = (%q, %q), want (%q, %q)",
					tt.payload, user, message, tt.wantUser, tt.wantMessage)
			}
		})
	}
}
