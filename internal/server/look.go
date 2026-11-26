package server

import (
	"errors"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func Look(gameServer *GameServer, currentRoom string, command protocol.Command, registry *Registry) (string, error) {

	dataRoom, ok := gameServer.Game.Room(currentRoom)
	if !ok {
		return "", errors.New("room not found")
	}
	ok = protocol.ParseNoArgs(command.Args)
	if !ok {
		return protocol.FormatErr(protocol.ErrBadRequest), nil
	}

	lookRoom := protocol.Room{
		ID:          dataRoom.ID,
		Name:        dataRoom.Name,
		Description: dataRoom.Description,
		Exits:       dataRoom.Exits,
	}
	lookResponse := protocol.LookResponse{
		Room:    lookRoom,
		Players: registry.getPlayersInRoom(currentRoom),
		Items:   dataRoom.Items,
		NPCs:    dataRoom.NPCs,
	}

	response, err := protocol.FormatOKJSON(lookResponse)
	if err != nil {
		return "", err
	}

	return response, nil
}
