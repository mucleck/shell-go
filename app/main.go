package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var path string

const messageCommandNotFound = "command not found"

func main() {
	path = getPath()
	for {
		fmt.Print("$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic("error")
		}

		command, args := parseInput(strings.TrimSpace(command))
		if checkExit(command) {
			break
		} else if command == "type" {
			handleType(args)
		} else if command == "echo" {
			handleEcho(args)
		} else if commandExists(command) != "" {
			executeCommand(command, args)
		} else {
			fmt.Println(command + ": " + messageCommandNotFound)
		}
	}
}

func parseInput(input string) (command string, args []string) {
	input = strings.TrimSuffix(input, "\n")
	firstsPace := strings.Index(input, " ")
	if firstsPace == -1 {
		return input, args
	}

	command = input[:firstsPace]
	args = strings.Split(input[firstsPace+1:], " ")
	return command, args
}

func checkExit(command string) bool {
	if command == "exit" {
		return true
	}
	return false
}

func handleEcho(args []string) {
	fmt.Println(strings.Join(args, " "))
}

func handleType(args []string) {

	if len(args) == 0 {
		return
	}

	regex := regexp.MustCompile(`echo|type|exit`)

	if regex.MatchString(args[0]) {
		fmt.Println(args[0] + " is a shell builtin")
		return
	}

	commandPath := commandExists(args[0])

	if commandPath != "" {
		fmt.Println(args[0] + " is " + commandPath)
		return
	}

	fmt.Println(args[0] + ": not found")
}

func commandExists(command string) string {
	directories := strings.SplitSeq(path, string(os.PathListSeparator))
	for dir := range directories {
		//check file info in dir
		file, err := os.Stat(filepath.Join(dir, command))
		if err != nil {
			continue
		}

		if strings.Contains(file.Mode().Perm().String(), "x") {
			return filepath.Join(dir, command)
		}
	}
	return ""

}

func executeCommand(command string, args []string) {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(cmd.Stderr, err)
	}
}

func getPath() (path string) {
	path, ok := os.LookupEnv("PATH")
	if !ok {
		fmt.Fprintln(os.Stderr, "Can't find PATH variable")
		return "" //In case it failed
	}
	return path
}
