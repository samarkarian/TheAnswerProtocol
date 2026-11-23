package server

import (
	"bufio"
	"fmt"
	"os"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func requireAuthentication(command protocol.Command, player *Player) bool {

	if command.Name != "CONNECT" && command.Name != "QUIT" && player.name == "" {
		fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrNotAuthenticated))
		return false
	}
	return true
}

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
		if authenticated := requireAuthentication(command, player); !authenticated {
			continue
		}
		switch command.Name {
		case "CONNECT":
			if closeConn := Connect(player, command, registry); closeConn {
				return
			}
		case "LOOK":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "MOVE":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "CHAT":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "WHO":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "GROUP":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "TAKE":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "DROP":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "INVENTORY":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "TALK":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "ATTACK":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "STATUS":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "QUEST":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		case "QUESTS":
			fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
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
