package main

import (
	"fmt"
	"bufio"
	"os"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

const messageCommandNotFound = "command not found\n"

func main() {
	for {
		fmt.Print("$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic("error")
		}
		fmt.Printf("%s: " + messageCommandNotFound, command[:len(command)-1])
	}

}
