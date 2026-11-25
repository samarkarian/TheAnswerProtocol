package world

import "github.com/samarkarian/TheAnswerProtocol/internal/direction"

// Stats are the combat statistics shared by players and NPCs.
type Stats struct {
	HP      int `json:"hp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
}

// PlayerStats are the statistics every player starts with.
type PlayerStats struct {
	Stats
	RespawnHP int `json:"respawn_hp"` // HP given back to a player after death
}

// Room defines a Room data
type Room struct {
	ID          string                         `json:"id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Exits       map[direction.Direction]string `json:"exits"` // direction -> room ID
	Items       []string                       `json:"items"`
	NPCs        []string                       `json:"npcs"`
}

// Item defines an Item data
type Item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Obtainable  bool   `json:"obtainable"`
}

// NPC defines a NPC data
type NPC struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Dialogue    []string `json:"dialogue"`
	Hostile     bool     `json:"hostile"`
	Stats       *Stats   `json:"stats"` // nil when the NPC has no stats; set for every hostile NPC
	Respawn     int      `json:"respawn"`
}

// QuestType is the kind of objective of a quest.
type QuestType string

// Quest types.
const (
	QuestFetch   QuestType = "fetch"   // bring Count items Target to the giver
	QuestDefeat  QuestType = "defeat"  // defeat Count NPCs Target
	QuestDeliver QuestType = "deliver" // bring item Target to the NPC To
)

// Valid reports whether t is a known quest type.
func (t QuestType) Valid() bool {
	switch t {
	case QuestFetch, QuestDefeat, QuestDeliver:
		return true
	default:
		return false
	}
}

// Quest defines a Quest data
type Quest struct {
	ID          string    `json:"id"`
	Giver       string    `json:"giver"`
	Type        QuestType `json:"type"`
	Description string    `json:"description"`
	Target      string    `json:"target"`
	Count       int       `json:"count"`
	Reward      string    `json:"reward"`
	To          string    `json:"to"`
}

// World defines the world data
type World struct {
	Spawn  string      `json:"spawn"`
	Player PlayerStats `json:"player"`
	Rooms  []Room      `json:"rooms"`
	Items  []Item      `json:"items"`
	NPCs   []NPC       `json:"npcs"`
	Quests []Quest     `json:"quests"`

	// indexes built by Load, not in world.json
	rooms  map[string]*Room
	items  map[string]*Item
	npcs   map[string]*NPC
	quests map[string]*Quest
}

// Room returns the room with the given ID, and whether it exists.
func (w *World) Room(id string) (*Room, bool) {
	room, ok := w.rooms[id]
	return room, ok
}

// Item returns the item with the given ID, and whether it exists.
func (w *World) Item(id string) (*Item, bool) {
	item, ok := w.items[id]
	return item, ok
}

// NPC returns the NPC with the given ID, and whether it exists.
func (w *World) NPC(id string) (*NPC, bool) {
	npc, ok := w.npcs[id]
	return npc, ok
}

// Quest returns the quest with the given ID, and whether it exists.
func (w *World) Quest(id string) (*Quest, bool) {
	quest, ok := w.quests[id]
	return quest, ok
}
