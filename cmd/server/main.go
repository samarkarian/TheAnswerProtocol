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

func NewRegistery() *Registery {
	return &Registery{
		players: make(map[string]*Player),
	}
}

func NewPlayer(name string) *Player {
	return &Player{
		name:    name,
		connect: true,
	}
}

func (r *Registery) Add(name string) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.players[name]; ok {
		return fmt.Errorf("%w", ErrNameTaken)
	}
	r.players[name] = NewPlayer(name)
	return nil
}

func (r *Registery) Remove(name string) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.players, name)

	return nil
}

func handle(conn net.Conn, r *Registery) {

	name := ""
	defer conn.Close()
	defer func() {
        if name != "" {
            r.Remove(name)
        }
    }()

	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		cmd, args, found := strings.Cut(strings.TrimSpace(sc.Text()), " ")
		if !found {
			fmt.Fprintln(os.Stderr, "No argument found")
		}
		cmd = strings.ToUpper(cmd)
		switch cmd {
		case "CONNECT":
			args = strings.ToLower(args)
			if err := r.Add(args); err != nil {
				fmt.Fprintln(conn, err)
				return
			}
			name = args
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
	registry := NewRegistery()

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
		go handle(conn, registry)
	}
}
