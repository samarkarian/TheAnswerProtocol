package server

import (
	"fmt"
	"net"
	"os"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
	"github.com/samarkarian/TheAnswerProtocol/internal/world"
)

type GameServer struct {
	Game *world.World
}

func LoadGameData() (*world.World, error) {

	game, err := world.Load("data/world.json")
	if err != nil {
		return nil, err
	}

	err = game.Validate()
	if err != nil {
		return nil, err
	}

	return game, nil
}

func Start() error {

	game, err := LoadGameData()
	if err != nil {
		return err
	}
	gameServer := GameServer{
		Game: game,
	}

	ln, err := net.Listen("tcp", "0.0.0.0:4243")
	if err != nil {
		return err
	}

	registry := NewRegistry()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		s, err := protocol.FormatOK("hello proto=1")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			if _, err := fmt.Fprint(conn, protocol.FormatErr(protocol.ErrSendFailed)); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if err := conn.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			continue
		}
		if _, err := fmt.Fprint(conn, s); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		go HandleConnection(NewPlayer(conn, gameServer.Game.Spawn), registry, &gameServer)
	}
}
