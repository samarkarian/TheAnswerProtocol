package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// Error returned when a room exists with the same ID
var ErrDuplicateRoom = errors.New("world: duplicate room")

// Room defines a Room data
type Room struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"` // direction -> room ID
}

// World defines the world data
type World struct {
	Spawn string `json:"spawn"`
	Rooms []Room `json:"rooms"`

	byID map[string]*Room // index built by Load, not in world.json
}

// Load reads and decodes the world file and indexes its rooms by ID. It does not check that the world is consistent: see Validate.
func Load(path string) (*World, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("world: %w", err)
	}
	defer f.Close()

	var world World

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&world); err != nil {
		return nil, fmt.Errorf("world: decode %s: %w", path, err)
	}

	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("world: decode %s: unexpected data after the JSON object", path)
	}

	world.byID = make(map[string]*Room)
	for i := range world.Rooms {
		roomID := world.Rooms[i].ID
		_, ok := world.byID[roomID]
		if ok {
			return nil, fmt.Errorf("%w: %q", ErrDuplicateRoom, roomID)
		}
		world.byID[roomID] = &world.Rooms[i]
	}

	return &world, nil
}
