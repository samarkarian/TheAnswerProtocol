package server

import (
	"errors"
	"fmt"

	"github.com/samarkarian/TheAnswerProtocol/internal/direction"
	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func Move(gameServer *GameServer, command protocol.Command, player *Player, registry *Registry) (string, error) {

	registry.mu.Lock()
	defer registry.mu.Unlock()

	dataRoom, ok := gameServer.Game.Room(player.currentRoom)
	if !ok {
		return "", errors.New("room not found")
	}

	if command.Args == "" {
		return protocol.FormatErr(protocol.ErrBadRequest), nil
	}

	dir, ok := direction.Parse(command.Args)
	if !ok {
		return protocol.FormatErr(protocol.ErrNoExit), nil
	}

	roomID, ok := dataRoom.Exits[dir]
	if !ok {
		return protocol.FormatErr(protocol.ErrNoExit), nil
	}

	s := fmt.Sprintf("room=%s", roomID)
	response, err := protocol.FormatOK(s)
	if err != nil {
		return "", err
	}
	player.currentRoom = roomID

	return response, nil

}
