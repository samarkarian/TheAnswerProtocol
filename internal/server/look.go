package server

import (

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)


func Look(gameServer *GameServer, currentRoom string, command protocol.Command) (string, error) {

    dataRoom := gameServer.Game.Rooms[currentRoom]

    s, err := protocol.FormatOKJSON(dataRoom)
    if err != nil {
        return "", err
    }

    return s, nil
}