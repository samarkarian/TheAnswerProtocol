package world

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// Errors returned by Load
var ErrDuplicateID = errors.New("world: duplicate id")

// createRoomsMap creates the rooms map attribute
func (w *World) createRoomsMap() error {
	w.rooms = make(map[string]*Room)
	for i := range w.Rooms {
		roomID := w.Rooms[i].ID
		_, ok := w.rooms[roomID]
		if ok {
			return fmt.Errorf("%w: %q", ErrDuplicateID, roomID)
		}
		w.rooms[roomID] = &w.Rooms[i]
	}
	return nil
}

// createItemsMap creates the items map attribute
func (w *World) createItemsMap() error {
	w.items = make(map[string]*Item)
	for i := range w.Items {
		itemID := w.Items[i].ID
		_, ok := w.items[itemID]
		if ok {
			return fmt.Errorf("%w: %q", ErrDuplicateID, itemID)
		}
		w.items[itemID] = &w.Items[i]
	}
	return nil
}

// createNPCsMap creates the npcs map attribute
func (w *World) createNPCsMap() error {
	w.npcs = make(map[string]*NPC)
	for i := range w.NPCs {
		npcID := w.NPCs[i].ID
		_, ok := w.npcs[npcID]
		if ok {
			return fmt.Errorf("%w: %q", ErrDuplicateID, npcID)
		}
		w.npcs[npcID] = &w.NPCs[i]
	}
	return nil
}

// createQuestsMap creates the quests map attribute
func (w *World) createQuestsMap() error {
	w.quests = make(map[string]*Quest)
	for i := range w.Quests {
		questID := w.Quests[i].ID
		_, ok := w.quests[questID]
		if ok {
			return fmt.Errorf("%w: %q", ErrDuplicateID, questID)
		}
		w.quests[questID] = &w.Quests[i]
	}
	return nil
}

// Load reads and decodes the world file and indexes its rooms by ID. It does not check that the world is consistent: see Validate.
func Load(path string) (*World, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("world: %w", err)
	}

	var world World

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&world); err != nil {
		return nil, fmt.Errorf("world: decode %s: %w", path, err)
	}

	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("world: decode %s: unexpected data after the JSON object", path)
	}

	if err := world.createRoomsMap(); err != nil {
		return nil, err
	}

	if err := world.createItemsMap(); err != nil {
		return nil, err
	}

	if err := world.createNPCsMap(); err != nil {
		return nil, err
	}

	if err := world.createQuestsMap(); err != nil {
		return nil, err
	}

	return &world, nil
}
