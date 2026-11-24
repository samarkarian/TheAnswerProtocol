package server

type GameData struct {
	Start  string           `json:"start"`
	Player PlayerAttributes `json:"player"`
	Rooms  map[string]Room  `json:"rooms"`
	Items  map[string]Item  `json:"items"`
	Npcs   map[string]Npc   `json:"npcs"`
	Quests map[string]Quest `json:"quests"`
}

type StartPlace struct {
	Start string `json:"start"`
}

type PlayerAttributes struct {
	Hp        int `json:"hp"`
	Attack    int `json:"attack"`
	Defense   int `json:"defense"`
	RespawnHp int `json:"respawn_hp"`
}

type Room struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Exits       map[string]string `json:"exits"`
	Items       []string          `json:"items"`
	Npcs        []string          `json:"npcs"`
}

type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Obtainable  bool   `json:"obtainable"`
}

type Npc struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Dialogue    []string  `json:"dialogue"`
	Hostile     bool      `json:"hostile"`
	StatsNpc    *StatsNpc `json:"stats,omitempty"`
	Respawn     int       `json:"respawn,omitempty"`
}

type StatsNpc struct {
	Hp      int `json:"hp"`
	Attack  int `json:"attack"`
	Defense int `json:"defense"`
}

type Quest struct {
	Giver       string `json:"giver"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Target      string `json:"target"`
	Count       int    `json:"count"`
	Reward      string `json:"reward"`
	To          string `json:"to,omitempty"`
}
