package world

import (
	"encoding/json"
	"errors"
	"path/filepath"
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

// square returns a valid world: four rooms in a loop, plus a dead-end branch.
//
//	a - b
//	|   |
//	d - c - e
func square() World {
	return World{
		Spawn: "loc.a",
		Rooms: []Room{
			room("loc.a", exits{east: "loc.b", south: "loc.d"}),
			room("loc.b", exits{west: "loc.a", south: "loc.c"}),
			room("loc.c", exits{north: "loc.b", west: "loc.d", east: "loc.e"}),
			room("loc.d", exits{north: "loc.a", east: "loc.c"}),
			room("loc.e", exits{west: "loc.c"}),
		},
	}
}

// find returns the room with the given ID in a world being built.
func find(t *testing.T, w *World, id string) *Room {
	t.Helper()
	for i := range w.Rooms {
		if w.Rooms[i].ID == id {
			return &w.Rooms[i]
		}
	}
	t.Fatalf("test setup: no room %q", id)
	return nil
}

// load encodes w to JSON and loads it back with Load, so the index is built.
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
			},
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

		// Rooms IDs
		{
			name: "empty id",
			mutate: func(t *testing.T, w *World) {
				find(t, w, "loc.e").ID = "" // loc.c east now leads nowhere
			},
			wantErrs: []error{ErrValueIsEmpty, ErrInvalidPrefix, ErrUnknownExit},
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

		// Map structure
		{
			name: "unreachable rooms",
			mutate: func(t *testing.T, w *World) {
				w.Rooms = append(w.Rooms,
					room("loc.x", exits{east: "loc.y"}),
					room("loc.y", exits{west: "loc.x"}),
				)
			},
			wantErrs:  []error{ErrUnreachableRoom},
			wantInMsg: []string{"loc.x", "loc.y"},
		},
		{
			name: "line without loop",
			mutate: func(t *testing.T, w *World) {
				*w = World{Spawn: "loc.a", Rooms: []Room{
					room("loc.a", exits{east: "loc.b"}),
					room("loc.b", exits{west: "loc.a", east: "loc.c"}),
					room("loc.c", exits{west: "loc.b"}),
				}}
			},
			wantErrs: []error{ErrNoLoop},
			notWant:  []error{ErrUnreachableRoom},
		},
		{
			name: "single room",
			mutate: func(t *testing.T, w *World) {
				*w = World{Spawn: "loc.a", Rooms: []Room{room("loc.a", nil)}}
			},
			wantErrs: []error{ErrIsolatedRoom},
		},

		// Structure checks expect the data to be correct.
		{
			name: "structure checks skipped when data is invalid",
			mutate: func(t *testing.T, w *World) {
				delete(find(t, w, "loc.b").Exits, west)
				w.Rooms = append(w.Rooms,
					room("loc.x", exits{east: "loc.y"}),
					room("loc.y", exits{west: "loc.x"}),
				)
			},
			wantErrs: []error{ErrMissingReturnExit},
			notWant:  []error{ErrUnreachableRoom, ErrNoLoop},
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
