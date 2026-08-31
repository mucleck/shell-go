package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

const messageCommandNotFound = "command not found"

func main() {
	for {
		fmt.Print("$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic("error")
		} 

		command = strings.TrimSpace(command)
		if checkExit(command) {
			break
		} else if strings.HasPrefix(command, "type") {
			handleType(command)
		}else if strings.HasPrefix(command, "echo") {
			handleEcho(command)
		} else {
			fmt.Println(command + ": " + messageCommandNotFound)
		}
	}
}

func checkExit(command string) bool {
	if command == "exit" {
		return true
	}
	return false
}

func handleEcho(command string) {
	fmt.Println(command[5:])
}

func handleType(command string) {
	regex := regexp.MustCompile(`echo|type|exit`)

	if !regex.MatchString(command[5:]) {
		fmt.Println(command + ": not found")
	}

	fmt.Println(command[5:] + " is a shell builtin")
}
