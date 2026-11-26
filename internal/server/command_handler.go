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
		case protocol.CmdConnect:
			if closeConn := Connect(player, command, registry); closeConn {
				return
			}
		case protocol.CmdLook:
			response, err := Look(gameServer, player.currentRoom, command, registry)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				response = protocol.FormatErr(protocol.ErrSendFailed)
			}
			_, err = fmt.Fprint(player.conn, response)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		case protocol.CmdMove:
			response, err := Move(gameServer, command, player, registry)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				response = protocol.FormatErr(protocol.ErrSendFailed)
			}
			_, err = fmt.Fprint(player.conn, response)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}
		case protocol.CmdChat:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdWho:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdGroup:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdTake:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdDrop:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdInventory:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdTalk:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdAttack:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdStatus:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdQuest:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdQuests:
			if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
				return
			}
		case protocol.CmdQuit:
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
