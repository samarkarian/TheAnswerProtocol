package protocol

import (
	"errors"
	"strings"
	"testing"

	"github.com/samarkarian/TheAnswerProtocol/internal/direction"
)

// Every Format function checks the whole line with ValidateLine: a "\n" or
// "\r" would let a player inject a fake line, and an escape sequence would
// reach the other players' terminals. A line built here must also be one that
// ParseLine accepts, which the tests check for every valid case.

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
		{"terminal escape", "\x1b[2J", "", ErrControlChar},
		{"tab inside", "a\tb", "", ErrControlChar},
		{"invalid UTF-8", "caf\xc3", "", ErrInvalidUTF8},
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
			if tt.wantErr == nil {
				m, err := ParseLine(got)
				if err != nil || m.Kind != KindOK || m.Data != tt.data {
					t.Errorf("ParseLine(%q) = %+v, %v; want OK with the same data", got, m, err)
				}
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
					Exits:       map[direction.Direction]string{"north": "loc.bakery"},
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
			// json.Marshal would write \u0026, \u003c and \u003e: unreadable in nc
			"HTML characters are not escaped",
			QuestOffer{
				QuestID:     "fetch_herbs",
				Description: "Bring me 3 herbs & 1 <rare> flower",
				Reward:      "gold_coin",
				Status:      "available",
			},
			`OK {"quest_id":"fetch_herbs","description":"Bring me 3 herbs & 1 <rare> flower",` +
				`"reward":"gold_coin","status":"available"}` + "\n",
			false,
		},
		{
			// json escapes \n and the other C0 characters, but not DEL (0x7f)
			// nor U+0080-U+009F: only ValidateLine catches them
			"control character json lets through",
			[]string{"a\x7fb"},
			"",
			true,
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
			if !tt.wantErr {
				if m, err := ParseLine(got); err != nil || m.Kind != KindOK {
					t.Errorf("ParseLine(%q) = %+v, %v; want an OK line", got, m, err)
				}
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
		{"terminal escape in payload", ScopeRoom, "CHAT", "alice \x1b[2J", "", ErrControlChar},
		{"tab in payload", ScopeRoom, "CHAT", "alice a\tb", "", ErrControlChar},
		{"invalid UTF-8 in payload", ScopeRoom, "CHAT", "alice caf\xc3", "", ErrInvalidUTF8},

		// never build a line that ParseLine refuses or splits differently
		{"empty payload", ScopeRoom, "CHAT", "", "", ErrMalformedEvent},
		{"empty payload for stats", ScopeStats, "", "", "", ErrMalformedEvent},
		{"empty category", "", "CHAT", "alice hi", "", ErrMalformedEvent},
		{"empty type outside stats", ScopeRoom, "", "alice hi", "", ErrMalformedEvent},
		{"space in category", "ROOM X", "CHAT", "alice hi", "", ErrMalformedEvent},
		{"space in type", ScopeRoom, "CHAT X", "alice hi", "", ErrMalformedEvent},
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
			if tt.wantErr == nil {
				m, err := ParseLine(got)
				if err != nil || m.Category != tt.cat || m.Type != tt.typ || m.Payload != tt.payload {
					t.Errorf("ParseLine(%q) = %+v, %v; want the same category, type and payload", got, m, err)
				}
			}
		})
	}
}

// RFC 42TAP 9.4 recommends 1024 bytes per line. The subject sets no other
// limit, so it applies to what the server sends too: a longer line would be
// refused by ParseLine (and probably by other groups' clients).
func TestFormatLineTooLong(t *testing.T) {
	long := strings.Repeat("a", MaxLineLen)
	tests := []struct {
		name    string
		format  func() (string, error)
		wantErr error
	}{
		{"OK at exactly the limit", func() (string, error) { return FormatOK(strings.Repeat("a", MaxLineLen-3)) }, nil},
		{"OK one byte over", func() (string, error) { return FormatOK(strings.Repeat("a", MaxLineLen-2)) }, ErrLineTooLong},
		{"event over the limit", func() (string, error) { return FormatEvent(ScopeRoom, "CHAT", "alice "+long) }, ErrLineTooLong},
		// "EVT ROOM CHAT " is 14 bytes: each piece is short, the whole line is not
		{"event at exactly the limit", func() (string, error) { return FormatEvent(ScopeRoom, "CHAT", strings.Repeat("a", MaxLineLen-14)) }, nil},
		{"event one byte over", func() (string, error) { return FormatEvent(ScopeRoom, "CHAT", strings.Repeat("a", MaxLineLen-13)) }, ErrLineTooLong},
		{"JSON over the limit", func() (string, error) { return FormatOKJSON([]string{long}) }, ErrLineTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.format()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if n := len(strings.TrimSuffix(got, "\n")); n > MaxLineLen {
				t.Errorf("line is %d bytes, want at most %d", n, MaxLineLen)
			}
		})
	}
}

// FormatErr takes no free text: every code and message is a constant of
// protocol.go. Writing each line by hand also checks those constants against
// RFC 42TAP 8.2 (400 and 403 are not in the RFC table: a team decision).
func TestFormatErr(t *testing.T) {
	tests := []struct {
		name string
		e    ErrorCode
		want string
	}{
		{"name in use", ErrNameInUse, "ERR 201 NAME_IN_USE\n"},
		{"no exit", ErrNoExit, "ERR 301 NO_EXIT\n"},
		{"bad request", ErrBadRequest, "ERR 400 BAD_REQUEST\n"},
		{"not in group", ErrNotInGroup, "ERR 401 NOT_IN_GROUP\n"},
		{"already in group", ErrAlreadyInGroup, "ERR 402 ALREADY_IN_GROUP\n"},
		{"not authenticated", ErrNotAuthenticated, "ERR 403 NOT_AUTHENTICATED\n"},
		{"item not found", ErrItemNotFound, "ERR 404 ITEM_NOT_FOUND\n"},
		{"item not in inventory", ErrItemNotInInventory, "ERR 404 ITEM_NOT_IN_INVENTORY\n"},
		{"npc not found", ErrNPCNotFound, "ERR 404 NPC_NOT_FOUND\n"},
		{"npc not hostile", ErrNPCNotHostile, "ERR 405 NPC_NOT_HOSTILE\n"},
		{"no quest available", ErrNoQuestAvailable, "ERR 406 NO_QUEST_AVAILABLE\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatErr(tt.e)
			if got != tt.want {
				t.Errorf("FormatErr(%+v) = %q, want %q", tt.e, got, tt.want)
			}
			m, err := ParseLine(got)
			if err != nil || m.Kind != KindErr || m.Code != tt.e.Code || m.ErrMsg != tt.e.Message {
				t.Errorf("ParseLine(%q) = %+v, %v; want the same code and message", got, m, err)
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
