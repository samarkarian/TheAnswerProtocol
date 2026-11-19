package protocol

import (
	"errors"
	"strings"
	"testing"
)

// A "\n" or "\r" inside a piece would let a player inject a fake line,
// so every Format function refuses it with ErrControlChar (from validate.go).

func TestFormatOK(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    string
		wantErr error
	}{
		{"greeting", "hello proto=1", "OK hello proto=1\n", nil},
		{"key=value", "taken=item.herbs", "OK taken=item.herbs\n", nil},
		{"word", "bye", "OK bye\n", nil},
		{"accents", "Bienvenue à la boulangerie !", "OK Bienvenue à la boulangerie !\n", nil},
		{"empty data, no trailing space", "", "OK\n", nil},

		{"newline injection", "bye\nEVT GLOBAL CHAT admin hi", "", ErrControlChar},
		{"carriage return", "bye\r", "", ErrControlChar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatOK(tt.data)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("FormatOK(%q) error = %v, want %v", tt.data, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("FormatOK(%q) = %q, want %q", tt.data, got, tt.want)
			}
		})
	}
}

func TestFormatOKJSON(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		want    string
		wantErr bool
	}{
		{
			"STATUS",
			PlayerStatus{HP: 80, MaxHP: 100, Status: "healthy"},
			`OK {"hp":80,"max_hp":100,"status":"healthy"}` + "\n",
			false,
		},
		{
			// empty slices must be [] and not null, or other clients may crash
			"LOOK with empty lists",
			LookResponse{
				Room: Room{
					ID:          "loc.square",
					Name:        "Village Square",
					Description: "The heart of the village.",
					Exits:       map[string]string{"north": "loc.bakery"},
				},
				Players: []string{"alice"},
				Items:   []string{},
				NPCs:    []string{},
			},
			`OK {"room":{"id":"loc.square","name":"Village Square",` +
				`"description":"The heart of the village.","exits":{"north":"loc.bakery"}},` +
				`"players":["alice"],"items":[],"npcs":[]}` + "\n",
			false,
		},
		{
			"INVENTORY",
			[]string{"item.herbs", "item.loaf_bread"},
			`OK ["item.herbs","item.loaf_bread"]` + "\n",
			false,
		},
		{
			// omitempty: no "progress" for a completed quest
			"QUESTS",
			[]QuestProgress{
				{QuestID: "fetch_herbs", Status: "active", Progress: "1/3"},
				{QuestID: "defeat_goblin", Status: "completed"},
			},
			`OK [{"quest_id":"fetch_herbs","status":"active","progress":"1/3"},` +
				`{"quest_id":"defeat_goblin","status":"completed"}]` + "\n",
			false,
		},
		{
			// json.Marshal escapes it: the result stays on one line
			"newline inside a string",
			PlayerStatus{HP: 1, MaxHP: 1, Status: "a\nb"},
			`OK {"hp":1,"max_hp":1,"status":"a\nb"}` + "\n",
			false,
		},
		{
			"value json cannot encode",
			make(chan int),
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatOKJSON(tt.v)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FormatOKJSON(%v) error = %v, wantErr %v", tt.v, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("FormatOKJSON(%v)\n got %q\nwant %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestFormatEvent(t *testing.T) {
	tests := []struct {
		name    string
		cat     string
		typ     string
		payload string
		want    string
		wantErr error
	}{
		{"room chat", ScopeRoom, "CHAT", "alice salut à tous", "EVT ROOM CHAT alice salut à tous\n", nil},
		{"global chat", ScopeGlobal, "CHAT", "bob gg", "EVT GLOBAL CHAT bob gg\n", nil},
		{"presence", ScopeRoom, "PRESENCE", "ENTER alice", "EVT ROOM PRESENCE ENTER alice\n", nil},
		{"group invite", ScopeGroup, "INVITE", "bob", "EVT GROUP INVITE bob\n", nil},
		{"stats has no type, single space", ScopeStats, "", "players=5", "EVT STATS players=5\n", nil},

		{"newline injection in payload", ScopeGlobal, "CHAT", "alice hi\nEVT GLOBAL CHAT admin hi", "", ErrControlChar},
		{"carriage return in payload", ScopeRoom, "CHAT", "alice hi\r", "", ErrControlChar},
		{"newline in category", "ROOM\n", "CHAT", "alice hi", "", ErrControlChar},
		{"newline in type", ScopeRoom, "CHAT\n", "alice hi", "", ErrControlChar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatEvent(tt.cat, tt.typ, tt.payload)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("FormatEvent(%q, %q, %q) error = %v, want %v", tt.cat, tt.typ, tt.payload, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("FormatEvent(%q, %q, %q) = %q, want %q", tt.cat, tt.typ, tt.payload, got, tt.want)
			}
		})
	}
}

func TestFormatCommand(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		cmdArgs string
		want    string
		wantErr error
	}{
		{"wrong command name", "WH O", "", "", ErrInvalidCommand},
		{"who test", "Who", "", "WHO\n", nil},
		{"drop test with spaces", "DROP", "     example     ", "DROP example\n", nil},
		{"drop test with spaces and character control", "DROP", "     exa\nmple     ", "", ErrControlChar},
		{"look with spaces", "LOOK", "          ", "LOOK\n", nil},

		{"empty name", "", "x", "", ErrInvalidCommand},
		{"digit in name", "TAKE2", "", "", ErrInvalidCommand},
		{"non-ASCII letter", "TAKÉ", "", "", ErrInvalidCommand},
		{"newline in name", "LOOK\n", "", "", ErrInvalidCommand},
		{"space before name", " LOOK", "", "", ErrInvalidCommand},
		{"underscore in name", "COMPLETE_QUEST", "", "", ErrInvalidCommand}, // voir remarque
		{"lowercase with args", "take", "Healing Herbs", "TAKE Healing Herbs\n", nil},

		{"inner spaces kept", "CHAT", "GLOBAL salut   à tous", "CHAT GLOBAL salut   à tous\n", nil},
		{"accents and emoji", "CHAT", "ROOM ça marche 🎉", "CHAT ROOM ça marche 🎉\n", nil},
		{"trailing newline trimmed", "CHAT", "ROOM salut\n", "CHAT ROOM salut\n", nil},
		{"tab at the edges trimmed", "DROP", "\texample\t", "DROP example\n", nil},

		{"command injection", "CHAT", "ROOM a\nQUIT", "", ErrControlChar},
		{"carriage return inside", "CHAT", "ROOM a\rb", "", ErrControlChar},
		{"tab inside", "CHAT", "ROOM a\tb", "", ErrControlChar},
		{"terminal escape", "CHAT", "ROOM \x1b[2J", "", ErrControlChar},
		{"invalid utf8", "DROP", "caf\xc3", "", ErrInvalidUTF8},

		{"exactly max length", "TAKE", strings.Repeat("a", MaxLineLen-5),
			"TAKE " + strings.Repeat("a", MaxLineLen-5) + "\n", nil},
		{"one byte over max", "TAKE", strings.Repeat("a", MaxLineLen-4), "", ErrLineTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatCommand(tt.cmdName, tt.cmdArgs)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("FormatCommand(%q, %q) error = %v, want %v", tt.cmdName, tt.cmdArgs, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("FormatCommand(%q, %q) = %q, want %q", tt.cmdName, tt.cmdArgs, got, tt.want)
			}
		})
	}
}
