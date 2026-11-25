package world

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// writeWorld writes content to a temporary world file and returns its path.
func writeWorld(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "world.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// Minimal valid entries, used to build world files that differ in one section only.
const (
	roomJSON   = `{"id": "loc.hall", "name": "Hall", "description": "x", "exits": {}, "items": [], "npcs": []}`
	itemJSON   = `{"id": "item.key", "name": "Key", "description": "x", "obtainable": true}`
	npcJSON    = `{"id": "npc.guard", "name": "Guard", "description": "x", "dialogue": [], "hostile": false}`
	questJSON  = `{"id": "quest.q", "giver": "npc.guard", "type": "fetch", "description": "x", "target": "item.key", "count": 1, "reward": "item.key"}`
	playerJSON = `{"hp": 100, "attack": 10, "defense": 5, "respawn_hp": 50}`
)

// worldWith returns a minimal world file in which one section is replaced by value.
// An empty section name returns the minimal world unchanged.
func worldWith(section, value string) string {
	s := map[string]string{
		"player": playerJSON,
		"rooms":  "[" + roomJSON + "]",
		"items":  "[" + itemJSON + "]",
		"npcs":   "[" + npcJSON + "]",
		"quests": "[" + questJSON + "]",
	}
	if section != "" {
		s[section] = value
	}
	return fmt.Sprintf(`{"spawn": "loc.hall", "player": %s, "rooms": %s, "items": %s, "npcs": %s, "quests": %s}`,
		s["player"], s["rooms"], s["items"], s["npcs"], s["quests"])
}

func TestLoadValid(t *testing.T) {
	w, err := Load(filepath.Join("testdata", "valid.json"))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	// World and player
	if w.Spawn != "loc.hall" {
		t.Errorf("Spawn = %q, want %q", w.Spawn, "loc.hall")
	}
	if w.Player.HP != 100 || w.Player.Attack != 10 || w.Player.Defense != 5 || w.Player.RespawnHP != 50 {
		t.Errorf("Player = %+v, want hp 100, attack 10, defense 5, respawn_hp 50", w.Player)
	}

	// Sizes of the lists and of the indexes
	sizes := []struct {
		name        string
		list, index int
		want        int
	}{
		{"rooms", len(w.Rooms), len(w.rooms), 2},
		{"items", len(w.Items), len(w.items), 2},
		{"npcs", len(w.NPCs), len(w.npcs), 2},
		{"quests", len(w.Quests), len(w.quests), 1},
	}
	for _, s := range sizes {
		if s.list != s.want || s.index != s.want {
			t.Errorf("%s: %d in the list and %d in the index, want %d", s.name, s.list, s.index, s.want)
		}
	}

	// Rooms
	hall, ok := w.Room("loc.hall")
	if !ok {
		t.Fatal(`Room("loc.hall") not found`)
	}
	if hall != &w.Rooms[0] {
		t.Error(`Room("loc.hall") does not point into w.Rooms`)
	}
	if hall.Description != "The entrance of the campus." || hall.Exits["north"] != "loc.bocal" {
		t.Errorf("hall = %+v, want description and north exit from the file", hall)
	}
	if want := []string{"item.coffee", "item.coffee"}; !slices.Equal(hall.Items, want) {
		t.Errorf("hall.Items = %v, want %v (repeated IDs are kept)", hall.Items, want)
	}
	if want := []string{"npc.guard"}; !slices.Equal(hall.NPCs, want) {
		t.Errorf("hall.NPCs = %v, want %v", hall.NPCs, want)
	}

	// Items
	coffee, ok := w.Item("item.coffee")
	if !ok {
		t.Fatal(`Item("item.coffee") not found`)
	}
	if coffee != &w.Items[0] {
		t.Error(`Item("item.coffee") does not point into w.Items`)
	}
	if coffee.Name != "Coffee" || !coffee.Obtainable {
		t.Errorf("coffee = %+v, want name Coffee, obtainable", coffee)
	}
	if badge, ok := w.Item("item.badge"); !ok || badge.Obtainable {
		t.Errorf(`Item("item.badge") = %+v, %v, want an item that is not obtainable`, badge, ok)
	}

	// NPCs
	guard, ok := w.NPC("npc.guard")
	if !ok {
		t.Fatal(`NPC("npc.guard") not found`)
	}
	if guard != &w.NPCs[0] {
		t.Error(`NPC("npc.guard") does not point into w.NPCs`)
	}
	if guard.Hostile || len(guard.Dialogue) != 2 || guard.Dialogue[0] != "Badge, please." {
		t.Errorf("guard = %+v, want a peaceful NPC with 2 lines of dialogue", guard)
	}
	rat, ok := w.NPC("npc.cockroach")
	if !ok {
		t.Fatal(`NPC("npc.cockroach") not found`)
	}
	if !rat.Hostile || rat.Respawn != 30 {
		t.Errorf("cockroach = %+v, want hostile with respawn 30", rat)
	}
	if rat.Stats.HP != 30 || rat.Stats.Attack != 8 || rat.Stats.Defense != 2 {
		t.Errorf("cockroach stats = %+v, want hp 30, attack 8, defense 2", rat.Stats)
	}

	// Quests
	quest, ok := w.Quest("quest.deliver_badge")
	if !ok {
		t.Fatal(`Quest("quest.deliver_badge") not found`)
	}
	if quest != &w.Quests[0] {
		t.Error(`Quest("quest.deliver_badge") does not point into w.Quests`)
	}
	if quest.Type != QuestDeliver || quest.Target != "item.badge" || quest.To != "npc.guard" ||
		quest.Count != 1 || quest.Reward != "item.coffee" || quest.Giver != "npc.guard" {
		t.Errorf("quest = %+v, want the values from the file", quest)
	}

	// Unknown IDs
	if _, ok := w.Room("loc.nowhere"); ok {
		t.Error(`Room("loc.nowhere") found, want not found`)
	}
	if _, ok := w.Item("loc.hall"); ok {
		t.Error(`Item("loc.hall") found: the indexes must be separate`)
	}
}

// The minimal world used by TestLoadErrors must load: otherwise its cases
// would fail for the wrong reason.
func TestLoadMinimalWorld(t *testing.T) {
	if _, err := Load(writeWorld(t, worldWith("", ""))); err != nil {
		t.Fatalf("Load() error on the minimal world: %v", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	w, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Load() error = %v, want it to match fs.ErrNotExist", err)
	}
	if w != nil {
		t.Errorf("Load() returned a World on error: %+v", w)
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantErr   error  // nil: any error is accepted
		wantInMsg string // the message must contain it, if set
	}{
		// File level
		{"empty file", ``, nil, ""},
		{"invalid json", `{"spawn": "loc.hall",`, nil, ""},
		{"trailing data", worldWith("", "") + ` {}`, nil, ""},
		{"unknown top-level field", `{"spawn": "loc.hall", "romos": []}`, nil, ""},

		// Duplicate IDs, one case per kind
		{"duplicate room", worldWith("rooms", "["+roomJSON+", "+roomJSON+"]"), ErrDuplicateID, "loc.hall"},
		{"duplicate item", worldWith("items", "["+itemJSON+", "+itemJSON+"]"), ErrDuplicateID, "item.key"},
		{"duplicate npc", worldWith("npcs", "["+npcJSON+", "+npcJSON+"]"), ErrDuplicateID, "npc.guard"},
		{"duplicate quest", worldWith("quests", "["+questJSON+", "+questJSON+"]"), ErrDuplicateID, "quest.q"},

		// Unknown fields, one case per section
		{"unknown player field", worldWith("player", `{"hp": 100, "mana": 5}`), nil, ""},
		{"unknown room field", worldWith("rooms", `[{"id": "loc.hall", "exist": {}}]`), nil, ""},
		{"unknown item field", worldWith("items", `[{"id": "item.key", "weight": 2}]`), nil, ""},
		{"unknown npc field", worldWith("npcs", `[{"id": "npc.guard", "dialog": []}]`), nil, ""},
		{"unknown npc stats field", worldWith("npcs", `[{"id": "npc.guard", "stats": {"hp": 1, "speed": 3}}]`), nil, ""},
		{"unknown quest field", worldWith("quests", `[{"id": "quest.q", "rewards": []}]`), nil, ""},

		// Wrong types
		{"hp is a string", worldWith("player", `{"hp": "100"}`), nil, ""},
		{"rooms is an object", worldWith("rooms", `{}`), nil, ""},
		{"items is an object", worldWith("items", `{"item.key": {}}`), nil, ""},
		{"exits is a string", worldWith("rooms", `[{"id": "loc.hall", "exits": "north"}]`), nil, ""},
		{"room items is a string", worldWith("rooms", `[{"id": "loc.hall", "items": "item.key"}]`), nil, ""},
		{"obtainable is a string", worldWith("items", `[{"id": "item.key", "obtainable": "yes"}]`), nil, ""},
		{"dialogue is a string", worldWith("npcs", `[{"id": "npc.guard", "dialogue": "Hello"}]`), nil, ""},
		{"stats is an array", worldWith("npcs", `[{"id": "npc.guard", "stats": [1, 2, 3]}]`), nil, ""},
		{"count is a string", worldWith("quests", `[{"id": "quest.q", "count": "1"}]`), nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := Load(writeWorld(t, tt.content))
			if err == nil {
				t.Fatal("Load() error = nil, want an error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Load() error = %v, want it to match %v", err, tt.wantErr)
			}
			if tt.wantInMsg != "" && !strings.Contains(err.Error(), tt.wantInMsg) {
				t.Errorf("Load() error = %v, want it to mention %q", err, tt.wantInMsg)
			}
			if w != nil {
				t.Errorf("Load() returned a World on error: %+v", w)
			}
		})
	}
}

// Load only reads the file: checking that the world makes sense is Validate's job.
func TestLoadDoesNotValidate(t *testing.T) {
	content := worldWith("rooms", `[{"id": "loc.hall", "items": ["item.nowhere"], "npcs": ["npc.nobody"]}]`)
	content = strings.Replace(content, `"spawn": "loc.hall"`, `"spawn": "loc.nowhere"`, 1)
	if _, err := Load(writeWorld(t, content)); err != nil {
		t.Errorf("Load() error = %v, want nil (validation belongs to Validate)", err)
	}
}
