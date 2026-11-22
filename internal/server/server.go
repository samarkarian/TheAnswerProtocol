package server

import (
	"fmt"
	"net"
	"os"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func Start() error {

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
			fmt.Fprint(conn, protocol.FormatErr(protocol.ErrSendFailed))
			conn.Close()
			continue
		}
		fmt.Fprint(conn, s)
		go HandleConnection(NewPlayer(conn), registry)
	}
}
