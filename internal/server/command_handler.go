package server

import (
	"bufio"
	"fmt"
	"os"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func HandleConnection(player *Player, registry *Registry) {

	defer player.conn.Close()
	defer func() {
		if player.name != "" {
			registry.Remove(player.name)
		}
	}()

	sc := bufio.NewScanner(player.conn)
	for sc.Scan() {
		command, err := protocol.ParseCommand(sc.Text())
		if err != nil {
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
			continue
		}
		switch command.Name {
		case "CONNECT":
			if closeConn := Connect(player, command, registry); closeConn {
				return
			}
		case "QUIT":
			Quit(player)
			return
		default:
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		}
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}
