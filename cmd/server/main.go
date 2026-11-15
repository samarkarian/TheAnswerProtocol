package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var ErrNameTaken = errors.New("ERR 201 NAME_IN_USE")
var ErrUnknownCmd = errors.New("ERR unknown command")

type Player struct {
	name    string
	connect bool
}

type Registery struct {
	mu      sync.Mutex
	players map[string]*Player
}

func (r *Registery) Add(name string, p *Player) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.players = make(map[string]*Player)
	if _, ok := r.players[name]; ok {
		return fmt.Errorf("%w", ErrNameTaken)
	}
	r.players[name] = &Player{
		name:    name,
		connect: true,
	}
	return nil
}

func (r *Registery) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.players, name)

	return nil
}

func handle(conn net.Conn) {
	defer conn.Close()
	var register Registery
	var player Player
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		cmd, args, found := strings.Cut(sc.Text(), " ")
		if !found {
			fmt.Fprintln(os.Stderr, "No argument found")
		}
		cmd = strings.ToUpper(cmd)
		switch cmd {
		case "CONNECT":
			register.Add(args, &player)
			fmt.Fprintln(conn, "OK connected")
		case "QUIT":
			fmt.Fprintln(conn, "OK bye")
			return
		default:
			fmt.Fprintln(conn, ErrUnknownCmd)
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}

func main() {

	ln, err := net.Listen("tcp", "0.0.0.0:4243")

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		fmt.Fprintln(conn, "OK hello proto=1")
		go handle(conn)
	}
}
