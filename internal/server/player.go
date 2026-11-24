package server

import "net"

type Player struct {
	conn net.Conn
	name string
	currentRoom string
}

func NewPlayer(conn net.Conn, startRoom string) *Player {
	return &Player{
		conn: conn,
		currentRoom: startRoom,
	}
}
