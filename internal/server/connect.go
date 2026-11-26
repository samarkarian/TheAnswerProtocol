package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func Connect(player *Player, command protocol.Command, registry *Registry) bool {

	if player.name != "" {
		if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
			return true
		}
		return false
	}
	if err := protocol.ValidateUsername(command.Args); err != nil {
		if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest)); err != nil {
			return true
		}
		return false
	}
	name := strings.ToLower(command.Args)
	if err := registry.Add(name, player); err != nil {
		fmt.Fprint(os.Stderr, err)
		if _, writeErr := fmt.Fprint(player.conn, err); writeErr != nil {
			return true
		}
		return false
	}
	response, err := protocol.FormatOK("connected")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrConnectionFailed)); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return true
	}
	if _, err := fmt.Fprint(player.conn, response); err != nil {
		return true
	}
	return false
}
