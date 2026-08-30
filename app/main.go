package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
		if checkExit(command) {
			break
		}
		if strings.HasPrefix(command, "echo") {
			handleEcho(command)
		}
	}
}

func checkExit(command string) bool {
	if strings.TrimSpace(command) == "exit" {
		return true
	}
	return false
}

func handleEcho(command string) {
	fmt.Print(command[5:])
}
