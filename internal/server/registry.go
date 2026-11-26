package server

import (
	"errors"
	"sync"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

type Registry struct {
	mu      sync.Mutex
	players map[string]*Player
}

func NewRegistry() *Registry {
	return &Registry{
		players: make(map[string]*Player),
	}
}

func (registry *Registry) Add(name string, player *Player) error {

	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, ok := registry.players[name]; ok {
		return errors.New(protocol.FormatErr(protocol.ErrNameInUse))
	}
	player.name = name
	registry.players[name] = player
	return nil
}

func (registry *Registry) Remove(name string) {

	registry.mu.Lock()
	defer registry.mu.Unlock()
	delete(registry.players, name)

}

func (registry *Registry) getPlayersInRoom(currentRoom string) []string {

	registry.mu.Lock()
	defer registry.mu.Unlock()

	players := []string{}
	for k, v := range registry.players {
		if v.currentRoom == currentRoom {
			players = append(players, k)
		}
	}

	return players
}
