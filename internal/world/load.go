package world

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

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
