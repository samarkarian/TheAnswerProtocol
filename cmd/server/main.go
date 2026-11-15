package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func handle(conn net.Conn) {
	defer conn.Close()
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		cmd, args, found := strings.Cut(sc.Text(), " ")
		if !found {
			fmt.Fprintln(os.Stderr, "No argument found")
		}
		cmd = strings.ToUpper(cmd)
		switch cmd {
		case "CONNECT":
			fmt.Fprintln(conn, "OK connected")
		case "QUIT":
			fmt.Fprintln(conn, "OK bye")
			return
		default:
			fmt.Fprintln(conn, "ERR unknown command")
		}
		// fmt.Fprintln(conn, sc.Text())
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
