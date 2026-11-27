package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func read(c net.Conn) {
	for {
		message, _ := bufio.NewReader(c).ReadString('\n')
		fmt.Print("->: " + message)
	}
}

func Connect() error {
	c, err := net.Dial("tcp", "localhost:4243")
	if err != nil {
		return err
	}

	go read(c)
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(c, "%s", text+"\n")

		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return nil
		}
	}
}
