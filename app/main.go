package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
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

	executable := command[5:]
	if regex.MatchString(executable) {
		fmt.Println(executable + " is a shell builtin")
		return
	}

	path, ok := os.LookupEnv("PATH")
	if !ok {
		fmt.Fprintln(os.Stderr, "Can't find PATH variable")
	}

	directories := strings.SplitSeq(path, string(os.PathListSeparator))
	for dir := range directories {
		//check file info in dir
		file, err := os.Stat(filepath.Join(dir, executable))
		if err != nil {
			continue
		}

		if strings.Contains(file.Mode().Perm().String(), "x") {
			fmt.Println(executable + " is " + dir)
			return
		} 
	} 

	fmt.Println(executable + ": not found")
}
























