package world

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"
)

var (
	ErrInvalidID         = errors.New("world: invalid ID")
	ErrValueIsEmpty      = errors.New("world: empty value")
	ErrSpawnNotFound     = errors.New("world: spawn not found in rooms")
	ErrRoomPrefix        = errors.New("world: room ID must start with \"loc.\"")
	ErrIsolatedRoom      = errors.New("world: room does not have exit")
	ErrMissingReturnExit = errors.New("world: missing return exit")
	ErrUnknownExit       = errors.New("world: exit leads to an unknown room")
	ErrUnknownDirection  = errors.New("world: direction does not exist")
	ErrNoLoop            = errors.New("world: no loop")
	ErrUnreachableRoom   = errors.New("world: rooms unreachable starting from the spawn")
)

// validateRoom returns every problem found in room and its exits.
func (w *World) validateRoom(room *Room) []error {
	var errs []error
	if strings.TrimSpace(room.ID) == "" {
		errs = append(errs, fmt.Errorf("%w: for room key 'id'",
			ErrValueIsEmpty))
	}
	if strings.ContainsFunc(room.ID, unicode.IsSpace) {
		errs = append(errs, fmt.Errorf("%w: room '%s'",
			ErrInvalidID, room.ID))
	}
	if !strings.HasPrefix(room.ID, "loc.") {
		errs = append(errs, fmt.Errorf("%w: should be 'loc.%s'",
			ErrRoomPrefix, room.ID))
	}
	if strings.TrimSpace(room.Name) == "" {
		errs = append(errs, fmt.Errorf("%w: for room %s key 'name'",
			ErrValueIsEmpty, room.ID))
	}
	if strings.TrimSpace(room.Description) == "" {
		errs = append(errs, fmt.Errorf("%w: for room %s key 'description'",
			ErrValueIsEmpty, room.ID))
	}

	if len(room.Exits) == 0 {
		errs = append(errs, fmt.Errorf("%w: %s is isolated",
			ErrIsolatedRoom, room.ID))
	}

	for _, dir := range slices.Sorted(maps.Keys(room.Exits)) {
		exit := room.Exits[dir]
		oppDir, valid := dir.Opposite()
		if !valid {
			errs = append(errs, fmt.Errorf("%w: %s -%s-> %s, but %s is an invalid Direction",
				ErrUnknownDirection, room.ID, dir, exit, dir))
			continue
		}
		exitRoom, exists := w.Room(exit)
		if !exists {
			errs = append(errs, fmt.Errorf("%w: %s -%s-> %s, but %s doesn't exist",
				ErrUnknownExit, room.ID, dir, exit, exit))
			continue
		}
		oppDirExitRoom, ok := exitRoom.Exits[oppDir]
		if !ok || oppDirExitRoom != room.ID {
			errs = append(errs, fmt.Errorf("%w: %s -%s-> %s, but %s has no %s exit back",
				ErrMissingReturnExit, room.ID, dir, exit, exit, oppDir))
		}
	}
	return errs
}

// checkReachable reports whether all rooms are connected.
// It assumes that the spawn is a valid room.
func (w *World) checkReachable() error {
	spawnRoom, _ := w.Room(w.Spawn)
	queue := []*Room{spawnRoom}
	visited := map[string]bool{spawnRoom.ID: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, nextID := range current.Exits {
			if visited[nextID] {
				continue
			}
			next, _ := w.Room(nextID)
			visited[nextID] = true
			queue = append(queue, next)
		}
	}

	var unreachable []string
	for _, room := range w.Rooms {
		if !visited[room.ID] {
			unreachable = append(unreachable, room.ID)
		}
	}
	if len(unreachable) > 0 {
		return fmt.Errorf("%w: %s", ErrUnreachableRoom, strings.Join(unreachable, ", "))
	}
	return nil
}

// checkLoop reports whether the world contains at least one loop.
// It assumes every exit is valid and has its return exit.
func (w *World) checkLoop() error {
	exits := 0
	for _, room := range w.Rooms {
		exits += len(room.Exits)
	}
	// Each connection is counted twice: once from each side.
	if exits/2 < len(w.Rooms) {
		return ErrNoLoop
	}
	return nil
}

// Validate checks that the world is consistent. It returns nil if it is,
// or every problem found, joined with errors.Join: each one matches its
// sentinel error with errors.Is (for example ErrMissingReturnExit).
//
// It first checks the data itself:
//   - spawn is set and is the ID of a room
//   - every room has an ID that starts with "loc." and contains no whitespace
//   - every room has a name, a description and at least one exit
//   - every exit uses a known direction, leads to an existing room,
//     and that room has the return exit in the opposite direction
//
// Only if all of the above passes, it then checks the structure of the map:
//   - every room can be reached from the spawn
//   - the map contains at least one loop
//
// So after fixing a first set of errors, a second set may appear.
//
// Validate does not check the requirements specific to our game,
// such as the minimum number of rooms.
func (w *World) Validate() error {
	var errs []error

	if strings.TrimSpace(w.Spawn) == "" {
		errs = append(errs, fmt.Errorf("%w: for key 'spawn'",
			ErrValueIsEmpty))
	} else if _, exists := w.Room(w.Spawn); !exists {
		errs = append(errs, fmt.Errorf("%w: %q", ErrSpawnNotFound, w.Spawn))
	}

	for i := range w.Rooms {
		errs = append(errs, w.validateRoom(&w.Rooms[i])...)
	}

	// must be fixed before running checkReachable and checkLoop
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return errors.Join(w.checkReachable(), w.checkLoop())
}
