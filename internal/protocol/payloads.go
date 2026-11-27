package protocol

import (
	"github.com/samarkarian/TheAnswerProtocol/internal/direction"
)

// LOOK
type Room struct {
	ID          string                         `json:"id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Exits       map[direction.Direction]string `json:"exits"` // direction in key, room id in value
}

type LookResponse struct {
	Room    Room     `json:"room"`
	Players []string `json:"players"`
	Items   []string `json:"items"`
	NPCs    []string `json:"npcs"`
}

// ATTACK
type AttackResult struct {
	AttackerHP int    `json:"attacker_hp"`
	TargetHP   int    `json:"target_hp"`
	Damage     int    `json:"damage"`
	Status     string `json:"status"`
}

// STATUS
type PlayerStatus struct {
	HP     int    `json:"hp"`
	MaxHP  int    `json:"max_hp"`
	Status string `json:"status"`
}

// QUEST
type QuestOffer struct {
	QuestID     string `json:"quest_id"`
	Description string `json:"description"`
	Reward      string `json:"reward"`
	Status      string `json:"status"`
}

// QUESTS: list of QuestProgress
type QuestProgress struct {
	QuestID  string `json:"quest_id"`
	Status   string `json:"status"`
	Progress string `json:"progress,omitempty"` // empty if quest is completed
}
