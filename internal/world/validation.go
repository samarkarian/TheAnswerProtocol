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
	ErrInvalidID              = errors.New("world: invalid ID")
	ErrValueIsEmpty           = errors.New("world: empty value")
	ErrSpawnNotFound          = errors.New("world: spawn not found in rooms")
	ErrInvalidPrefix          = errors.New("world: invalid prefix")
	ErrIsolatedRoom           = errors.New("world: room does not have exit")
	ErrMissingReturnExit      = errors.New("world: missing return exit")
	ErrDuplicateNPC           = errors.New("world: duplicate NPC in room")
	ErrUnknownExit            = errors.New("world: exit leads to an unknown room")
	ErrUnknownDirection       = errors.New("world: direction does not exist")
	ErrNoLoop                 = errors.New("world: no loop")
	ErrUnreachableRoom        = errors.New("world: rooms unreachable starting from the spawn")
	ErrHostileNPCMissingStats = errors.New("world: hostile NPC must have stats")
	ErrUnknownItem            = errors.New("world: unknown item")
	ErrUnknownNPC             = errors.New("world: unknown NPC")
	ErrUnknownQuestType       = errors.New("world: quest type is unknown")
	ErrMissingTarget          = errors.New("world: missing target for quest")
	ErrInvalidCount           = errors.New("world: count cannot be less than or equal to 0")
	ErrInvalidStat            = errors.New("world: invalid stat value")
	ErrInvalidQuest           = errors.New("world: invalid quest")
	ErrInvalidRespawn         = errors.New("world: respawn cannot be less than 0")
	ErrEmptyDialogue          = errors.New("world: dialogue is empty")
	ErrNPCPlacedMoreThanOne   = errors.New("world: NPC is placed more than one time")
	ErrNotPlaced              = errors.New("world: not placed in rooms")
)

// checkID returns the problems of the ID of an element of the given kind,
// for example checkID("room", "loc.", room.ID).
func checkID(kind, prefix, id string) []error {
	if strings.TrimSpace(id) == "" {
		return []error{fmt.Errorf("%w: %s without an id", ErrValueIsEmpty, kind)}
	}
	var errs []error
	if strings.ContainsFunc(id, unicode.IsSpace) {
		errs = append(errs, fmt.Errorf("%w: %s %q contains a space", ErrInvalidID, kind, id))
	}
	if !strings.HasPrefix(id, prefix) {
		errs = append(errs, fmt.Errorf("%w: %s %q must start with %q", ErrInvalidPrefix, kind, id, prefix))
	}
	return errs
}

// checkNotBlank returns an error if value is empty or only made of spaces.
func checkNotBlank(kind, id, field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%w: %s %q has no %s", ErrValueIsEmpty, kind, id, field)
	}
	return nil
}

// checkStats returns the problems of the Stats. The id is used for error description.
func checkStats(id string, stats *Stats) []error {
	var errs []error
	if stats.HP <= 0 {
		errs = append(errs, fmt.Errorf("%w: %s HP cannot be less than or equal to 0", ErrInvalidStat, id))
	}
	if stats.Attack < 0 {
		errs = append(errs, fmt.Errorf("%w: %s attack cannot be negative", ErrInvalidStat, id))
	}
	if stats.Defense < 0 {
		errs = append(errs, fmt.Errorf("%w: %s defense cannot be negative", ErrInvalidStat, id))
	}
	return errs
}

// validateRoom returns every problem found in room and its exits.
func (w *World) validateRoom(room *Room) []error {
	errs := checkID("room", "loc.", room.ID)
	errs = append(errs,
		checkNotBlank("room", room.ID, "name", room.Name),
		checkNotBlank("room", room.ID, "description", room.Description),
	)

	if len(room.Exits) == 0 {
		errs = append(errs, fmt.Errorf("%w: %s is isolated",
			ErrIsolatedRoom, room.ID))
	}

	for _, dir := range slices.Sorted(maps.Keys(room.Exits)) {
		exit := room.Exits[dir]
		oppDir, valid := dir.Opposite()
		if !valid {
			errs = append(errs, fmt.Errorf("%w: %s -%s-> %s, but %s is an invalid direction",
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

	for _, item := range room.Items {
		if _, exists := w.Item(item); !exists {
			errs = append(errs, fmt.Errorf("%w: %q in %s", ErrUnknownItem, item, room.ID))
		}
	}
	foundDuplicate := []string{}
	for i, npc := range room.NPCs {
		if _, exists := w.NPC(npc); !exists {
			errs = append(errs, fmt.Errorf("%w: %q in %s", ErrUnknownNPC, npc, room.ID))
		} else if !slices.Contains(foundDuplicate, npc) { // check for duplicate
			for j, otherNPC := range room.NPCs {
				if i != j && npc == otherNPC {
					foundDuplicate = append(foundDuplicate, npc) // so we don't append twice
					errs = append(errs, fmt.Errorf("%w: %q in %s", ErrDuplicateNPC, npc, room.ID))
					break
				}
			}
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

// validateItem returns every problem found in item.
func (w *World) validateItem(item *Item) []error {
	errs := checkID("item", "item.", item.ID)
	errs = append(errs,
		checkNotBlank("item", item.ID, "name", item.Name),
		checkNotBlank("item", item.ID, "description", item.Description),
	)
	return errs
}

// validateNPC returns every problem found in NPC.
func (w *World) validateNPC(npc *NPC) []error {
	errs := checkID("npc", "npc.", npc.ID)
	errs = append(errs,
		checkNotBlank("npc", npc.ID, "name", npc.Name),
		checkNotBlank("npc", npc.ID, "description", npc.Description),
	)
	if npc.Hostile && npc.Stats == nil {
		errs = append(errs, fmt.Errorf("%w: %q", ErrHostileNPCMissingStats, npc.ID))
	} else if npc.Stats != nil {
		errs = append(errs, checkStats(npc.ID, npc.Stats)...)
	}
	if npc.Respawn < 0 {
		errs = append(errs, fmt.Errorf("%w: %s", ErrInvalidRespawn, npc.ID))
	}
	if len(npc.Dialogue) == 0 {
		errs = append(errs, fmt.Errorf("%w: %s", ErrEmptyDialogue, npc.ID))
	}
	placedRooms := []string{}
	for i := range w.Rooms {
		if slices.Contains(w.Rooms[i].NPCs, npc.ID) {
			placedRooms = append(placedRooms, "'"+w.Rooms[i].ID+"'")
		}
	}
	if len(placedRooms) > 1 {
		errs = append(errs, fmt.Errorf("%w: %q in %s", ErrNPCPlacedMoreThanOne, npc.ID, strings.Join(placedRooms, ", ")))
	}
	return errs
}

// isNPCPlaced returns an error if npc id is not placed in a room. The user is used for error description.
func (w *World) isNPCPlaced(user, id string) error {
	for i := range w.Rooms {
		if slices.Contains(w.Rooms[i].NPCs, id) {
			return nil
		}
	}
	return fmt.Errorf("%w: %q is not placed in a room, used by %q", ErrNotPlaced, id, user)
}

// isItemPlaced returns an error if item id is not placed in a room, nor is it obtained as a quest reward. The user is used for error description.
func (w *World) isItemPlaced(user, id string) error {
	for i := range w.Rooms {
		if slices.Contains(w.Rooms[i].Items, id) {
			return nil
		}
	}
	for i := range w.Quests {
		if w.Quests[i].ID != user && w.Quests[i].Reward == id {
			return nil
		}
	}
	return fmt.Errorf("%w: %q is not placed in a room, nor is it obtained as a quest reward, used by %q", ErrNotPlaced, id, user)
}

// validateQuest returns every problem found in quest.
func (w *World) validateQuest(quest *Quest) []error {
	errs := checkID("quest", "quest.", quest.ID)
	errs = append(errs,
		checkNotBlank("quest", quest.ID, "description", quest.Description),
	)

	err := checkNotBlank("quest", quest.ID, "giver", quest.Giver)
	errs = append(errs, err)
	if err == nil {
		if _, exists := w.NPC(quest.Giver); !exists {
			errs = append(errs, fmt.Errorf("%w: %q for 'giver' in %s", ErrUnknownNPC, quest.Giver, quest.ID))
		} else {
			errs = append(errs, w.isNPCPlaced(quest.ID, quest.Giver))
		}
	}

	if !quest.Type.Valid() {
		errs = append(errs, fmt.Errorf("%w: %q for %s", ErrUnknownQuestType, quest.Type, quest.ID))
	}

	switch quest.Type {
	case QuestFetch, QuestDeliver:
		// item must be obtainable
		err := checkNotBlank("quest", quest.ID, "target", quest.Target)
		errs = append(errs, err)
		if err == nil {
			item, exists := w.Item(quest.Target)
			if !exists {
				errs = append(errs, fmt.Errorf("%w: %q for 'target' in %s", ErrUnknownItem, quest.Target, quest.ID))
			} else if !item.Obtainable {
				errs = append(errs, fmt.Errorf("%w: %q, item %q must be obtainable for a %q quest", ErrInvalidQuest, quest.ID, item.ID, quest.Type))
			}
			if exists {
				errs = append(errs, w.isItemPlaced(quest.ID, item.ID))
			}
		}
	case QuestDefeat:
		// must be hostile
		err := checkNotBlank("quest", quest.ID, "target", quest.Target)
		errs = append(errs, err)
		if err == nil {
			npc, exists := w.NPC(quest.Target)
			if !exists {
				errs = append(errs, fmt.Errorf("%w: %q for 'target' in %s", ErrUnknownNPC, quest.Target, quest.ID))
			} else if !npc.Hostile {
				errs = append(errs, fmt.Errorf("%w: %q, NPC %q must be hostile for a %q quest", ErrInvalidQuest, quest.ID, npc.ID, quest.Type))
			}
			if exists {
				errs = append(errs, w.isNPCPlaced(quest.ID, npc.ID))
			}
		}
		// no default because if unknown type, it's handled before
	}

	if quest.Type == QuestDeliver {
		if strings.TrimSpace(quest.To) != "" {
			if _, exists := w.NPC(quest.To); !exists {
				errs = append(errs, fmt.Errorf("%w: %q for 'to' in %s", ErrUnknownNPC, quest.To, quest.ID))
			} else {
				errs = append(errs, w.isNPCPlaced(quest.ID, quest.To))
			}
		} else {
			errs = append(errs, fmt.Errorf("%w: %s 'to' cannot be empty with type %q", ErrInvalidQuest, quest.ID, quest.Type))
		}
	}

	if quest.Count <= 0 {
		errs = append(errs, fmt.Errorf("%w: in %s", ErrInvalidCount, quest.ID))
	}

	err = checkNotBlank("quest", quest.ID, "reward", quest.Reward)
	errs = append(errs, err)
	if err == nil {
		if _, exists := w.Item(quest.Reward); !exists {
			errs = append(errs, fmt.Errorf("%w: %q for 'reward' in %s", ErrUnknownItem, quest.Reward, quest.ID))
		}
	}

	return errs
}

// Validate checks that the world is consistent. It returns nil if it is,
// or every problem found, joined with errors.Join: each one matches its
// sentinel error with errors.Is (for example ErrMissingReturnExit).
//
// It checks the data of every element:
//   - spawn is set and is the ID of a room
//   - every ID starts with its prefix ("loc.", "item.", "npc." or "quest.")
//     and contains no whitespace
//   - every room, item and NPC has a name and a description,
//     and every quest has a description
//   - every room has at least one exit; every exit uses a known direction,
//     leads to an existing room, and that room has the return exit
//   - the items and NPCs placed in a room exist
//   - every NPC has at least one line of dialogue and a respawn delay
//     that is not negative; a hostile NPC has stats
//   - all stats have a positive HP and no negative attack or defense,
//     and the player has a positive respawn_hp
//   - every quest has a known type, an existing giver NPC, an existing
//     reward item and a count of at least 1. Its target is an obtainable
//     item for a fetch or deliver quest, or a hostile NPC for a defeat quest.
//     A deliver quest also needs an existing NPC in "to", which is ignored
//     for the other types.
//
// If the spawn and the rooms are valid, it also checks the structure of the map:
//   - every room can be reached from the spawn
//   - the map contains at least one loop
//
// Errors in the spawn or the rooms therefore hide the map errors: after
// fixing them, new errors may appear. Errors in items, NPCs, quests and
// player stats do not.
//
// Validate does not check the requirements specific to our game,
// such as the minimum number of rooms.
func (w *World) Validate() error {
	var errs, mapErrs []error

	if strings.TrimSpace(w.Spawn) == "" {
		mapErrs = append(mapErrs, fmt.Errorf("%w: for key 'spawn'",
			ErrValueIsEmpty))
	} else if _, exists := w.Room(w.Spawn); !exists {
		mapErrs = append(mapErrs, fmt.Errorf("%w: %q", ErrSpawnNotFound, w.Spawn))
	}
	for i := range w.Rooms {
		mapErrs = append(mapErrs, w.validateRoom(&w.Rooms[i])...)
	}

	for i := range w.Items {
		errs = append(errs, w.validateItem(&w.Items[i])...)
	}
	for i := range w.NPCs {
		errs = append(errs, w.validateNPC(&w.NPCs[i])...)
	}
	for i := range w.Quests {
		errs = append(errs, w.validateQuest(&w.Quests[i])...)
	}

	// player stats
	errs = append(errs, checkStats("player", &w.Player.Stats)...)
	if w.Player.RespawnHP <= 0 {
		errs = append(errs, fmt.Errorf("%w: player respawn_hp cannot be less than or equal to 0", ErrInvalidStat))
	}

	// rooms must be fixed before running checkReachable and checkLoop
	if err := errors.Join(mapErrs...); err != nil {
		errs = append(errs, mapErrs...)
		return errors.Join(errs...)
	}

	errs = append(errs, w.checkReachable(), w.checkLoop())
	return errors.Join(errs...)
}
