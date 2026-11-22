package server

import (
	"fmt"
	"os"

	"github.com/samarkarian/TheAnswerProtocol/internal/protocol"
)

// Quit handles "QUIT". The caller always closes the connection afterwards.
func Quit(player *Player) {
	s, err := protocol.FormatOK("bye")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprint(player.conn, protocol.FormatErr(protocol.ErrSendFailed))
		return
	}
	fmt.Fprint(player.conn, s)
}
