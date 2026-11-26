package world

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/samarkarian/TheAnswerProtocol/internal/direction"
)

type exits = map[direction.Direction]string

const (
	north = direction.North
	east  = direction.East
	south = direction.South
	west  = direction.West
)

// room builds a room with valid fields and the given exits.
func room(id string, ex exits) Room {
	return Room{ID: id, Name: "Name of " + id, Description: "Description of " + id, Exits: ex}
}

// newItem builds an item with valid fields.
func newItem(id string, obtainable bool) Item {
	return Item{ID: id, Name: "Name of " + id, Description: "Description of " + id, Obtainable: obtainable}
}

// newNPC builds an NPC with valid fields. A hostile NPC gets stats and a respawn delay.
func newNPC(id string, hostile bool) NPC {
	n := NPC{
		ID:          id,
		Name:        "Name of " + id,
		Description: "Description of " + id,
		Dialogue:    []string{"Hello."},
		Hostile:     hostile,
	}
	if hostile {
		n.Stats = &Stats{HP: 30, Attack: 8, Defense: 2}
		n.Respawn = 30
	}
	return n
}

// square returns a valid world: four rooms in a loop, plus a dead-end branch,
// with items, NPCs and one quest of each type.
//
//	a - b
//	|   |
//	d - c - e
//
//	a: item.coffee, npc.guard
//	b: item.badge, npc.staff
//	c: item.machine (not obtainable)
//	e: npc.rat (hostile)
//	item.reward is only given by quests.
func square() World {
	rooms := []Room{
		room("loc.a", exits{east: "loc.b", south: "loc.d"}),
		room("loc.b", exits{west: "loc.a", south: "loc.c"}),
		room("loc.c", exits{north: "loc.b", west: "loc.d", east: "loc.e"}),
		room("loc.d", exits{north: "loc.a", east: "loc.c"}),
		room("loc.e", exits{west: "loc.c"}),
	}
	rooms[0].Items, rooms[0].NPCs = []string{"item.coffee"}, []string{"npc.guard"}
	rooms[1].Items, rooms[1].NPCs = []string{"item.badge"}, []string{"npc.staff"}
	rooms[2].Items = []string{"item.machine"}
	rooms[4].NPCs = []string{"npc.rat"}

	return World{
		Spawn:  "loc.a",
		Player: PlayerStats{Stats: Stats{HP: 100, Attack: 10, Defense: 5}, RespawnHP: 50},
		Rooms:  rooms,
		Items: []Item{
			newItem("item.coffee", true),
			newItem("item.badge", true),
			newItem("item.machine", false),
			newItem("item.reward", true),
		},
		NPCs: []NPC{
			newNPC("npc.guard", false),
			newNPC("npc.staff", false),
			newNPC("npc.rat", true),
		},
		Quests: []Quest{
			{
				ID: "quest.fetch", Giver: "npc.staff", Type: QuestFetch,
				Description: "Bring the badge.", Target: "item.badge", Count: 1, Reward: "item.reward",
			},
			{
				ID: "quest.defeat", Giver: "npc.guard", Type: QuestDefeat,
				Description: "Defeat the rat twice.", Target: "npc.rat", Count: 2, Reward: "item.reward",
			},
			{
				ID: "quest.deliver", Giver: "npc.staff", Type: QuestDeliver,
				Description: "Bring the badge to the guard.", Target: "item.badge", To: "npc.guard",
				Count: 1, Reward: "item.coffee",
			},
		},
	}
}

// get returns the element of list whose ID is id.
func get[T any](t *testing.T, list []T, id string, idOf func(*T) string) *T {
	t.Helper()
	for i := range list {
		if idOf(&list[i]) == id {
			return &list[i]
		}
	}
	t.Fatalf("test setup: no element %q", id)
	return nil
}

// find returns the room with the given ID in a world being built.
func find(t *testing.T, w *World, id string) *Room {
	t.Helper()
	return get(t, w.Rooms, id, func(r *Room) string { return r.ID })
}

// findItem returns the item with the given ID in a world being built.
func findItem(t *testing.T, w *World, id string) *Item {
	t.Helper()
	return get(t, w.Items, id, func(i *Item) string { return i.ID })
}

// findNPC returns the NPC with the given ID in a world being built.
func findNPC(t *testing.T, w *World, id string) *NPC {
	t.Helper()
	return get(t, w.NPCs, id, func(n *NPC) string { return n.ID })
}

// findQuest returns the quest with the given ID in a world being built.
func findQuest(t *testing.T, w *World, id string) *Quest {
	t.Helper()
	return get(t, w.Quests, id, func(q *Quest) string { return q.ID })
}

// load encodes w to JSON and loads it back with Load, so the indexes are built.
func load(t *testing.T, w World) *World {
	t.Helper()
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(writeWorld(t, string(data)))
	if err != nil {
		t.Fatalf("test setup: Load() error: %v", err)
	}
	return loaded
}

// disconnected adds two rooms linked together but unreachable from the spawn.
func disconnected(w *World) {
	w.Rooms = append(w.Rooms,
		room("loc.x", exits{east: "loc.y"}),
		room("loc.y", exits{west: "loc.x"}),
	)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(t *testing.T, w *World) // changes applied to square()
		wantErrs  []error                      // nil: the world must be valid
		notWant   []error                      // errors that must NOT be reported
		wantInMsg []string                     // words the message must contain
	}{
		// Valid worlds
		{name: "valid world"},
		{
			name: "valid loop without branch",
			mutate: func(t *testing.T, w *World) {
				w.Rooms = w.Rooms[:4] // remove loc.e
				delete(find(t, w, "loc.c").Exits, east)
				w.Quests = slices.Delete(w.Quests, 1, 2) // delete quest because npc is in loc.e
			},
		},
		{
			name:   "to is ignored for a fetch quest",
			mutate: func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").To = "npc.nobody" },
		},
		{
			name:   "peaceful NPC without stats",
			mutate: func(t *testing.T, w *World) { findNPC(t, w, "npc.guard").Stats = nil },
		},

		// Spawn
		{
			name:     "empty spawn",
			mutate:   func(t *testing.T, w *World) { w.Spawn = "" },
			wantErrs: []error{ErrValueIsEmpty},
			notWant:  []error{ErrSpawnNotFound},
		},
		{
			name:     "blank spawn",
			mutate:   func(t *testing.T, w *World) { w.Spawn = "   " },
			wantErrs: []error{ErrValueIsEmpty},
		},
		{
			name:      "spawn not found",
			mutate:    func(t *testing.T, w *World) { w.Spawn = "loc.hal" },
			wantErrs:  []error{ErrSpawnNotFound},
			wantInMsg: []string{"loc.hal"},
		},

		// Room IDs and fields
		{
			name: "empty id",
			mutate: func(t *testing.T, w *World) {
				find(t, w, "loc.e").ID = "" // loc.c east now leads nowhere
			},
			wantErrs: []error{ErrValueIsEmpty, ErrUnknownExit},
		},
		{
			name: "id without prefix",
			mutate: func(t *testing.T, w *World) {
				find(t, w, "loc.c").Exits[east] = "room.e"
				find(t, w, "loc.e").ID = "room.e"
			},
			wantErrs:  []error{ErrInvalidPrefix},
			wantInMsg: []string{"room.e"},
		},
		{
			name: "id with space",
			mutate: func(t *testing.T, w *World) {
				find(t, w, "loc.c").Exits[east] = "loc.e e"
				find(t, w, "loc.e").ID = "loc.e e"
			},
			wantErrs: []error{ErrInvalidID},
		},
		{
			name:     "empty name",
			mutate:   func(t *testing.T, w *World) { find(t, w, "loc.b").Name = "" },
			wantErrs: []error{ErrValueIsEmpty},
		},
		{
			name:     "blank name",
			mutate:   func(t *testing.T, w *World) { find(t, w, "loc.b").Name = "  \t " },
			wantErrs: []error{ErrValueIsEmpty},
		},
		{
			name:      "empty description",
			mutate:    func(t *testing.T, w *World) { find(t, w, "loc.d").Description = "" },
			wantErrs:  []error{ErrValueIsEmpty},
			wantInMsg: []string{"loc.d", "description"},
		},
		{
			name: "isolated room",
			mutate: func(t *testing.T, w *World) {
				delete(find(t, w, "loc.e").Exits, west) // loc.c east has no way back
			},
			wantErrs:  []error{ErrIsolatedRoom, ErrMissingReturnExit},
			wantInMsg: []string{"loc.e"},
		},

		// Room content
		{
			name: "unknown item in room",
			mutate: func(t *testing.T, w *World) {
				a := find(t, w, "loc.a")
				a.Items = append(a.Items, "item.nope")
			},
			wantErrs:  []error{ErrUnknownItem},
			wantInMsg: []string{"item.nope", "loc.a"},
		},
		{
			name: "unknown NPC in room",
			mutate: func(t *testing.T, w *World) {
				a := find(t, w, "loc.a")
				a.NPCs = append(a.NPCs, "npc.nobody")
			},
			wantErrs:  []error{ErrUnknownNPC},
			wantInMsg: []string{"npc.nobody", "loc.a"},
		},

		// Exits
		{
			name:     "unknown direction",
			mutate:   func(t *testing.T, w *World) { find(t, w, "loc.a").Exits["up"] = "loc.b" },
			wantErrs: []error{ErrUnknownDirection},
		},
		{
			name: "upper-case direction",
			mutate: func(t *testing.T, w *World) {
				e := find(t, w, "loc.e")
				delete(e.Exits, west)
				e.Exits["West"] = "loc.c"
			},
			wantErrs:  []error{ErrUnknownDirection, ErrMissingReturnExit},
			wantInMsg: []string{"West"},
		},
		{
			name:      "exit to unknown room",
			mutate:    func(t *testing.T, w *World) { find(t, w, "loc.e").Exits[north] = "loc.z" },
			wantErrs:  []error{ErrUnknownExit},
			wantInMsg: []string{"loc.z"},
		},
		{
			name:      "missing return exit",
			mutate:    func(t *testing.T, w *World) { delete(find(t, w, "loc.b").Exits, west) },
			wantErrs:  []error{ErrMissingReturnExit},
			wantInMsg: []string{"loc.a", "loc.b"},
		},
		{
			name: "return exit leads to another room",
			mutate: func(t *testing.T, w *World) {
				find(t, w, "loc.b").Exits[west] = "loc.d"
			},
			wantErrs: []error{ErrMissingReturnExit},
		},
		{
			name:     "exit to itself",
			mutate:   func(t *testing.T, w *World) { find(t, w, "loc.a").Exits[north] = "loc.a" },
			wantErrs: []error{ErrMissingReturnExit},
		},

		// Items
		{
			name: "item id without prefix",
			mutate: func(t *testing.T, w *World) {
				findItem(t, w, "item.machine").ID = "machine"
				find(t, w, "loc.c").Items = []string{"machine"}
			},
			wantErrs:  []error{ErrInvalidPrefix},
			wantInMsg: []string{"machine"},
		},
		{
			name: "item id with space",
			mutate: func(t *testing.T, w *World) {
				findItem(t, w, "item.machine").ID = "item.big machine"
				find(t, w, "loc.c").Items = []string{"item.big machine"}
			},
			wantErrs: []error{ErrInvalidID},
		},
		{
			name:      "item without name",
			mutate:    func(t *testing.T, w *World) { findItem(t, w, "item.coffee").Name = "" },
			wantErrs:  []error{ErrValueIsEmpty},
			wantInMsg: []string{"item.coffee", "name"},
		},
		{
			name:      "item with blank description",
			mutate:    func(t *testing.T, w *World) { findItem(t, w, "item.coffee").Description = "   " },
			wantErrs:  []error{ErrValueIsEmpty},
			wantInMsg: []string{"item.coffee", "description"},
		},

		// NPCs
		{
			name: "npc id without prefix",
			mutate: func(t *testing.T, w *World) {
				w.NPCs = append(w.NPCs, newNPC("merchant", false))
			},
			wantErrs:  []error{ErrInvalidPrefix},
			wantInMsg: []string{"merchant"},
		},
		{
			name:      "npc without name",
			mutate:    func(t *testing.T, w *World) { findNPC(t, w, "npc.guard").Name = "" },
			wantErrs:  []error{ErrValueIsEmpty},
			wantInMsg: []string{"npc.guard", "name"},
		},
		{
			name:      "npc without dialogue",
			mutate:    func(t *testing.T, w *World) { findNPC(t, w, "npc.guard").Dialogue = nil },
			wantErrs:  []error{ErrEmptyDialogue},
			wantInMsg: []string{"npc.guard"},
		},
		{
			name:      "hostile npc without stats",
			mutate:    func(t *testing.T, w *World) { findNPC(t, w, "npc.rat").Stats = nil },
			wantErrs:  []error{ErrHostileNPCMissingStats},
			wantInMsg: []string{"npc.rat"},
		},
		{
			name:     "npc with zero HP",
			mutate:   func(t *testing.T, w *World) { findNPC(t, w, "npc.rat").Stats.HP = 0 },
			wantErrs: []error{ErrInvalidStat},
		},
		{
			name:     "npc with negative attack",
			mutate:   func(t *testing.T, w *World) { findNPC(t, w, "npc.rat").Stats.Attack = -1 },
			wantErrs: []error{ErrInvalidStat},
		},
		{
			name:     "npc with negative defense",
			mutate:   func(t *testing.T, w *World) { findNPC(t, w, "npc.rat").Stats.Defense = -1 },
			wantErrs: []error{ErrInvalidStat},
		},
		{
			name:     "peaceful npc with invalid stats",
			mutate:   func(t *testing.T, w *World) { findNPC(t, w, "npc.guard").Stats = &Stats{HP: 0} },
			wantErrs: []error{ErrInvalidStat},
		},
		{
			name:      "negative respawn",
			mutate:    func(t *testing.T, w *World) { findNPC(t, w, "npc.rat").Respawn = -1 },
			wantErrs:  []error{ErrInvalidRespawn},
			wantInMsg: []string{"npc.rat"},
		},

		// Player
		{
			name:      "player with zero HP",
			mutate:    func(t *testing.T, w *World) { w.Player.HP = 0 },
			wantErrs:  []error{ErrInvalidStat},
			wantInMsg: []string{"player"},
		},
		{
			name:     "player with negative defense",
			mutate:   func(t *testing.T, w *World) { w.Player.Defense = -1 },
			wantErrs: []error{ErrInvalidStat},
		},
		{
			name:      "player with zero respawn_hp",
			mutate:    func(t *testing.T, w *World) { w.Player.RespawnHP = 0 },
			wantErrs:  []error{ErrInvalidStat},
			wantInMsg: []string{"respawn_hp"},
		},

		// Quests: common fields
		{
			name:      "quest id without prefix",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").ID = "fetch" },
			wantErrs:  []error{ErrInvalidPrefix},
			wantInMsg: []string{"fetch"},
		},
		{
			name:      "quest without description",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Description = "" },
			wantErrs:  []error{ErrValueIsEmpty},
			wantInMsg: []string{"quest.fetch", "description"},
		},
		{
			name:      "quest without giver",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Giver = "" },
			wantErrs:  []error{ErrValueIsEmpty},
			wantInMsg: []string{"quest.fetch", "giver"},
			notWant:   []error{ErrUnknownNPC},
		},
		{
			name:      "unknown giver",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Giver = "npc.nobody" },
			wantErrs:  []error{ErrUnknownNPC},
			wantInMsg: []string{"npc.nobody", "giver"},
		},
		{
			name:      "unknown quest type",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Type = "steal" },
			wantErrs:  []error{ErrUnknownQuestType},
			wantInMsg: []string{"steal", "quest.fetch"},
		},
		{
			name:      "zero count",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.defeat").Count = 0 },
			wantErrs:  []error{ErrInvalidCount},
			wantInMsg: []string{"quest.defeat"},
		},
		{
			name:     "negative count",
			mutate:   func(t *testing.T, w *World) { findQuest(t, w, "quest.defeat").Count = -2 },
			wantErrs: []error{ErrInvalidCount},
		},
		{
			name:      "quest without reward",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Reward = "" },
			wantErrs:  []error{ErrValueIsEmpty},
			wantInMsg: []string{"reward"},
			notWant:   []error{ErrUnknownItem},
		},
		{
			name:      "unknown reward",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Reward = "item.gold" },
			wantErrs:  []error{ErrUnknownItem},
			wantInMsg: []string{"item.gold", "reward"},
		},

		// Quests: target depends on the type
		{
			name:      "fetch without target",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Target = "" },
			wantErrs:  []error{ErrValueIsEmpty},
			wantInMsg: []string{"target"},
			notWant:   []error{ErrUnknownItem},
		},
		{
			name:      "fetch with unknown item",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Target = "item.nope" },
			wantErrs:  []error{ErrUnknownItem},
			wantInMsg: []string{"item.nope", "target"},
		},
		{
			name:     "fetch targeting an NPC",
			mutate:   func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Target = "npc.rat" },
			wantErrs: []error{ErrUnknownItem},
		},
		{
			name:      "fetch item not obtainable",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.fetch").Target = "item.machine" },
			wantErrs:  []error{ErrInvalidQuest},
			wantInMsg: []string{"item.machine"},
		},
		{
			name:     "deliver item not obtainable",
			mutate:   func(t *testing.T, w *World) { findQuest(t, w, "quest.deliver").Target = "item.machine" },
			wantErrs: []error{ErrInvalidQuest},
		},
		{
			name:      "deliver without to",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.deliver").To = "" },
			wantErrs:  []error{ErrInvalidQuest},
			wantInMsg: []string{"'to'"},
		},
		{
			name:      "deliver to unknown NPC",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.deliver").To = "npc.nobody" },
			wantErrs:  []error{ErrUnknownNPC},
			wantInMsg: []string{"npc.nobody"},
		},
		{
			name:     "defeat without target",
			mutate:   func(t *testing.T, w *World) { findQuest(t, w, "quest.defeat").Target = "" },
			wantErrs: []error{ErrValueIsEmpty},
			notWant:  []error{ErrUnknownNPC},
		},
		{
			name:      "defeat with unknown NPC",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.defeat").Target = "npc.nobody" },
			wantErrs:  []error{ErrUnknownNPC},
			wantInMsg: []string{"npc.nobody", "target"},
		},
		{
			name:     "defeat targeting an item",
			mutate:   func(t *testing.T, w *World) { findQuest(t, w, "quest.defeat").Target = "item.coffee" },
			wantErrs: []error{ErrUnknownNPC},
		},
		{
			name:      "defeat a peaceful NPC",
			mutate:    func(t *testing.T, w *World) { findQuest(t, w, "quest.defeat").Target = "npc.guard" },
			wantErrs:  []error{ErrInvalidQuest},
			wantInMsg: []string{"npc.guard"},
		},

		// Quests: unplaced item or npc
		{
			name: "unplaced giver",
			mutate: func(t *testing.T, w *World) {
				find(t, w, "loc.b").NPCs = slices.Delete(find(t, w, "loc.b").NPCs, 0, 1)
			},
			wantErrs:  []error{ErrNotPlaced},
			wantInMsg: []string{"npc.staff", "quest.fetch", "quest.deliver"},
		},
		{
			name: "unplaced item",
			mutate: func(t *testing.T, w *World) {
				find(t, w, "loc.b").Items = slices.Delete(find(t, w, "loc.b").Items, 0, 1)
			},
			wantErrs:  []error{ErrNotPlaced},
			wantInMsg: []string{"item.badge", "quest.fetch", "quest.deliver"},
		},
		{
			name: "unplaced item but reward of quest",
			mutate: func(t *testing.T, w *World) {
				w.Quests[1].Reward = find(t, w, "loc.b").Items[0]
				find(t, w, "loc.b").Items = []string{}

			},
			wantErrs:  []error{},
			wantInMsg: []string{},
		},
		{
			name: "unplaced deliver",
			mutate: func(t *testing.T, w *World) {
				find(t, w, "loc.a").NPCs = []string{}
			},
			wantErrs:  []error{ErrNotPlaced},
			wantInMsg: []string{"npc.guard", "quest.deliver"},
		},

		// Map structure
		{
			name:      "unreachable rooms",
			mutate:    func(t *testing.T, w *World) { disconnected(w) },
			wantErrs:  []error{ErrUnreachableRoom},
			wantInMsg: []string{"loc.x", "loc.y"},
		},
		{
			name: "line without loop",
			mutate: func(t *testing.T, w *World) {
				w.Rooms = []Room{
					room("loc.a", exits{east: "loc.b"}),
					room("loc.b", exits{west: "loc.a", east: "loc.c"}),
					room("loc.c", exits{west: "loc.b"}),
				}
			},
			wantErrs: []error{ErrNoLoop},
			notWant:  []error{ErrUnreachableRoom},
		},
		{
			name: "single room",
			mutate: func(t *testing.T, w *World) {
				w.Rooms = []Room{room("loc.a", nil)}
			},
			wantErrs: []error{ErrIsolatedRoom},
		},

		// Which errors block the map checks
		{
			name: "structure checks skipped when rooms are invalid",
			mutate: func(t *testing.T, w *World) {
				delete(find(t, w, "loc.b").Exits, west)
				disconnected(w)
			},
			wantErrs: []error{ErrMissingReturnExit},
			notWant:  []error{ErrUnreachableRoom, ErrNoLoop},
		},
		{
			name: "structure checks skipped when spawn is invalid",
			mutate: func(t *testing.T, w *World) {
				w.Spawn = "loc.hal" // would panic if checkReachable ran
				disconnected(w)
			},
			wantErrs: []error{ErrSpawnNotFound},
			notWant:  []error{ErrUnreachableRoom},
		},
		{
			name: "structure checks skipped when a room lists an unknown item",
			mutate: func(t *testing.T, w *World) {
				a := find(t, w, "loc.a")
				a.Items = append(a.Items, "item.nope")
				disconnected(w)
			},
			wantErrs: []error{ErrUnknownItem},
			notWant:  []error{ErrUnreachableRoom},
		},
		{
			name: "structure checks still run when only an item is invalid",
			mutate: func(t *testing.T, w *World) {
				findItem(t, w, "item.coffee").Name = ""
				disconnected(w)
			},
			wantErrs: []error{ErrValueIsEmpty, ErrUnreachableRoom},
		},
		{
			name: "structure checks still run when only a quest is invalid",
			mutate: func(t *testing.T, w *World) {
				findQuest(t, w, "quest.defeat").Count = 0
				disconnected(w)
			},
			wantErrs: []error{ErrInvalidCount, ErrUnreachableRoom},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := square()
			if tt.mutate != nil {
				tt.mutate(t, &raw)
			}
			err := load(t, raw).Validate()

			if len(tt.wantErrs) == 0 {
				if err != nil {
					t.Fatalf("Validate() unexpected error:\n%v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want %v", tt.wantErrs)
			}
			for _, want := range tt.wantErrs {
				if !errors.Is(err, want) {
					t.Errorf("Validate() error does not match %v:\n%v", want, err)
				}
			}
			for _, notWant := range tt.notWant {
				if errors.Is(err, notWant) {
					t.Errorf("Validate() error should not match %v:\n%v", notWant, err)
				}
			}
			for _, s := range tt.wantInMsg {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("Validate() error does not mention %q:\n%v", s, err)
				}
			}
		})
	}
}

// The same world must always give the same message, whatever the map order.
func TestValidateIsDeterministic(t *testing.T) {
	raw := square()
	a := find(t, &raw, "loc.a")
	a.Exits["up"] = "loc.b"
	a.Exits["down"] = "loc.z"
	delete(find(t, &raw, "loc.b").Exits, west)
	w := load(t, raw)

	first := w.Validate()
	if first == nil {
		t.Fatal("Validate() = nil, want errors")
	}
	for range 20 {
		if got := w.Validate(); got.Error() != first.Error() {
			t.Fatalf("message changed between runs:\nfirst:\n%v\nthen:\n%v", first, got)
		}
	}
}

// Our own world must be valid and meet the requirements of the subject.
func TestRealWorld(t *testing.T) {
	w, err := Load(filepath.Join("..", "..", "data", "world.json"))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if err := w.Validate(); err != nil {
		t.Fatalf("Validate() error:\n%v", err)
	}
	if len(w.Rooms) < 8 {
		t.Errorf("the world has %d rooms, the subject requires at least 8", len(w.Rooms))
	}
}
