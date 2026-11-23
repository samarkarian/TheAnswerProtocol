package server

import (
	"fmt"
	"os"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

func Quit(player *Player) {
	s, err := protocol.FormatOK("bye")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if _, err := fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrSendFailed)); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	if _, err := fmt.Fprint(player.conn, s); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
