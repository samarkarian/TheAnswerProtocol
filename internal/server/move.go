package server

import (
	"errors"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func Move(gameServer *GameServer, currentRoom string, command protocol.Command, player *Player) (string, error) {
	dataRoom := gameServer.Game.Rooms[currentRoom]

	s, err := protocol.FormatOKJSON(dataRoom)
	if err != nil {
		return "", err
	}

	// Check si exits existe avec le file d'Ethan quand il est prêt

	is_here := 0
	dir := ""
	for key, value := range dataRoom.Exits {
		if command.Args == key {
			dir = value
			is_here = 1
		}
	}

	if is_here == 0 {
		return "s", errors.New("Exit not found")
	} else {
		player.currentRoom = dir
	}

	return s, nil
}
