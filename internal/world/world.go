package world

import (
	"errors"
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

// Room returns the room with the given ID, and whether it exists.
func (w *World) Room(id string) (*Room, bool) {
	room, ok := w.byID[id]
	return room, ok
}
