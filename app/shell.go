package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	path        string
	shellPrefix = "$ "

	// Errors
	ErrPathNotFound = errors.New("cannot find path")
	ErrPwdNotFound  = errors.New("cant get current working dir, weird")
)

const messageCommandNotFound = "command not found"

func ExecuteShell() error {
	var err error
	path, err = getPath()
	if err != nil {
		return err
	}

	for {
		fmt.Print(shellPrefix)

		command, err := readInput()
		if err != nil {
			log.Println(err)
			continue
		}

		command, args := parseInput(command)

		if handleCommand(command, args) {
			break
		}
	}

	return nil
}

func handleCommand(c string, args []string) bool {
	switch {
	case c == "exit":
		return true
	case c == "type":
		typeCommand(args)
	case c == "echo":
		echoCommand(args)
	case c == "pwd":
		pwdCommand()
	case commandExistsInPath(c) != "":
		executeCommand(c, args)
	default:
		fmt.Println(c + ": " + messageCommandNotFound)
	}
	return false
}

func pwdCommand() {
	dir, err := os.Getwd()
	if err != nil {
		log.Println(ErrPwdNotFound)
	}
	fmt.Println(dir)
}

func typeCommand(args []string) {

	if len(args) == 0 {
		return
	}

	regex := regexp.MustCompile(`echo|type|exit`)

	if regex.MatchString(args[0]) {
		fmt.Println(args[0] + " is a shell builtin")
		return
	}

	commandPath := commandExistsInPath(args[0])

	if commandPath != "" {
		fmt.Println(args[0] + " is " + commandPath)
		return
	}

	fmt.Println(args[0] + ": not found")
}

func echoCommand(args []string) {
	fmt.Println(strings.Join(args, " "))
}

func commandExistsInPath(command string) string {
	directories := strings.SplitSeq(path, string(os.PathListSeparator))
	for dir := range directories {
		commandWithDir := filepath.Join(dir, command)
		//check file info in dir
		file, err := os.Stat(commandWithDir)
		if err != nil {
			continue
		}

		if strings.Contains(file.Mode().Perm().String(), "x") {
			return commandWithDir
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

func readInput() (command string, err error) {
	command, err = bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return command, nil
}

func parseInput(input string) (command string, args []string) {
	input = strings.TrimSuffix(strings.TrimSpace(input), "\n")

	firstSpace := strings.Index(input, " ")
	if firstSpace == -1 {
		return input, args
	}

	return input[:firstSpace], strings.Split(input[firstSpace+1:], " ")
}

func getPath() (path string, err error) {
	path, ok := os.LookupEnv("PATH")
	if !ok {
		return "", ErrPathNotFound
	}

	return path, nil
}
