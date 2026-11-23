package server

import (
	"fmt"
	"net"
	"os"
	// "encoding/json"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func Start() error {

	byteFile, err := os.ReadFile("../../data/world.json")
	if err != nil {
		return err
	}
	file := string(byteFile)
	fmt.Print(file)

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
		go HandleConnection(NewPlayer(conn), registry)
	}
}
