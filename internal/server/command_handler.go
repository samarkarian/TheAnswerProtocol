package server

import (
	"bufio"
	"fmt"
	"os"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func requireAuthentication(command protocol.Command, player *Player) bool {

	if command.Name != "CONNECT" && command.Name != "QUIT" && player.name == "" {
		if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrNotAuthenticated)); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return false
	}
	return true
}

func HandleConnection(player *Player, registry *Registry, gameServer *GameServer) {

	defer func() {
		if err := player.conn.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	defer func() {
		if player.name != "" {
			registry.Remove(player.name)
		}
	}()

	sc := bufio.NewScanner(player.conn)
	for sc.Scan() {
		command, err := protocol.ParseCommand(sc.Text())
		if err != nil {
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
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
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "MOVE":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "CHAT":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "WHO":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "GROUP":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "TAKE":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "DROP":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "INVENTORY":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "TALK":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "ATTACK":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "STATUS":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "QUEST":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "QUESTS":
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case "QUIT":
			Quit(player)
			return
		default:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		}
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}
