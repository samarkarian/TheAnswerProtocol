package server

type GameData struct {
    Start  string          `json:"start"`
    Player PlayerAttributes `json:"player"`
    Items  Items           `json:"items"`
    Npcs   Npcs            `json:"npcs"`
    Quests Quests          `json:"quests"`
    Rooms  Rooms           `json:"rooms"`
}

type StartPlace struct {
	Start string `json:"start"`
}

type PlayerAttributes struct {
	Hp int `json:"hp"`
	Attack int `json:"attack"`
	Defense int `json:"defense"`
	RespawnHp int `json:"respawn_hp"`
}

type Items struct {
	Items map[string]Item `json:"items"`
}

type Item struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Obtainable bool `json:"obtainable"`
}

type Npc struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Dialogue []string `json:"dialogue"`
	Hostile bool `json:"hostile"`
	StatsNpc *StatsNpc `json:"stats,omitempty"`
	Respawn int `json:"respawn,omitempty"`
}

type StatsNpc struct {
	Hp int `json:"hp"`
	Attack int `json:"attack"`
	Defense int `json:"defense"`
}

type Npcs struct {
	Npcs map[string]Npc `json:"npcs"`
}

type Quest struct {
	Giver string `json:"giver"`
	Type string `json:"type"`
	Description string `json:"description"`
	Target string `json:"target"`
	Count int `json:"count"`
	Reward []Item `json:"reward"`
	To string `json:"to,omitempty"`
}

type Quests struct {
	Quests map[string]Quest `json:"quests"`
}

type Rooms struct {
	Rooms map[string]Room `json:"rooms"`
}

type Room struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Exits map[string]Room `json:"exits"`
	Items []string `json:"items"`
	Npcs []string `json:"npcs"`
}