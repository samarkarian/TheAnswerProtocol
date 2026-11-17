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
var ErrBadRequest = errors.New("ERR 400 BAD_REQUEST")

type Player struct {
	name    string
	connect bool
}

type Registry struct {
	mu      sync.Mutex
	players map[string]*Player
}

func NewRegistry() *Registry {
	return &Registry{
		players: make(map[string]*Player),
	}
}

func NewPlayer(name string) *Player {
	return &Player{
		name:    name,
		connect: true,
	}
}

func (r *Registry) Add(name string) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.players[name]; ok {
		return ErrNameTaken
	}
	r.players[name] = NewPlayer(name)
	return nil
}

func (r *Registry) Remove(name string) {

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.players, name)

}

func handle(conn net.Conn, r *Registry) {

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
		args = strings.TrimSpace(args)
		fields := strings.Fields(args)
		cmd = strings.ToUpper(cmd)
		switch cmd {
		case "CONNECT":
			if name != "" {
				fmt.Fprintln(conn, ErrBadRequest)
				continue
			}
			if !found || len(fields) > 1 {
				fmt.Fprintln(conn, ErrBadRequest)
				continue
			}
			args = strings.ToLower(args)
			if err := r.Add(args); err != nil {
				fmt.Fprintln(conn, err)
				continue
			}
			name = args
			fmt.Fprintln(conn, "OK connected")
		case "QUIT":
			fmt.Fprintln(conn, "OK bye")
			return
		default:
			fmt.Fprintln(conn, ErrBadRequest)
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

	registry := NewRegistry()
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
