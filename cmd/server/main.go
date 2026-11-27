package main

import (
	"fmt"
	"os"

	"github.com/samarkarian/TheAnswerProtocol/internal/server"
)

func main() {
	err := server.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ERR]:", err)
		os.Exit(1)
	}
}
