package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func Connect(player *Player, command protocol.Command, registry *Registry) bool {

	if player.name != "" {
		fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		return false
	}
	if err := protocol.ValidateUsername(command.Args); err != nil {
		fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrBadRequest))
		return false
	}
	name := strings.ToLower(command.Args)
	if err := registry.Add(name, player); err != nil {
		fmt.Fprint(os.Stderr, err)
		fmt.Fprint(player.conn, err)
		return false
	}
	s, err := protocol.FormatOK("connected")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrConnectionFailed))
		return true
	}
	fmt.Fprint(player.conn, s)
	return false
}
