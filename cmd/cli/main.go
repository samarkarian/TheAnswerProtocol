package main

import (
	"fmt"

	"github.com/samarkarian/TheAnswerProtocol/internal/client"
)

func main() {
	err := client.Connect()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}
