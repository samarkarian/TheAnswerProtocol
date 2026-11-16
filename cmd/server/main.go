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
		fmt.Println("hh")
		return fmt.Errorf("%w", ErrNameTaken)
	}
	r.players[name] = NewPlayer(name)
	fmt.Fprintln(os.Stdout, r.players)
	return nil
}

func (r *Registery) Remove(name string) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.players, name)
	fmt.Fprintln(os.Stdout, r.players)

	return nil
}

func handle(conn net.Conn, r *Registery) {

	defer conn.Close()
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		cmd, args, found := strings.Cut(strings.TrimSpace(sc.Text()), " ")
		if !found {
			fmt.Fprintln(os.Stderr, "No argument found")
		}
		name := args
		cmd = strings.ToUpper(cmd)
		switch cmd {
		case "CONNECT":
			if err := r.Add(name); err != nil {
				fmt.Fprintln(conn, err)
				return
			}
			fmt.Fprintln(conn, "OK connected")
		case "QUIT":
			r.Remove(name)
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
