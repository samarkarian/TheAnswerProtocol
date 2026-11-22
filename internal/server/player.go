package server

import "net"

type Player struct {
	conn net.Conn
	name string
}

func NewPlayer(conn net.Conn) *Player {
	return &Player{
		conn: conn,
	}
}
