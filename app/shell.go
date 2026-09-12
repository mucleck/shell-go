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

	ErrPathNotFound = errors.New("cannot find path")
	ErrPwdNotFound  = errors.New("cant get current working dir, weird")
)

const messageCommandNotFound = "command not found"

type Command struct {
	name string
	args []string
}

func ExecuteShell() error {
	var err error
	path, err = getPath()
	if err != nil {
		return err
	}

	for {
		fmt.Print(shellPrefix)

		input, err := readInput()
		if err != nil {
			log.Println(err)
			continue
		}

		command := parseInput(input)

		if handleCommand(command) {
			break
		}
	}

	return nil
}

func handleCommand(c Command) bool {
	switch {
	case c.name == "exit":
		return true
	case isBuiltinCommand(c.name):
		runBuiltinCommand(c)
	case commandExistsInPath(c.name) != "":
		runCommand(c)
	default:
		fmt.Println(c.name + ": " + messageCommandNotFound)
	}
	return false
}

func isBuiltinCommand(name string) bool {
	switch name {
	case "type", "echo", "pwd", "cd":
		return true
	}
	return false
}

func runBuiltinCommand(c Command) {
	switch c.name {
	case "type":
		typeCommand(c.args)
	case "echo":
		echoCommand(c.args)
	case "pwd":
		pwdCommand()
	case "cd":
		cdCommand(c.args)
	}
}

func cdCommand(args []string) {
	if len(args) == 0 || args[0] == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Println(err)
		}

		if err := os.Chdir(homeDir); err != nil {
			log.Println(err)
		}
		return
	}

	if err := os.Chdir(args[0]); err != nil {
		//log.Println(err)
		fmt.Println("cd: " + args[0] + ": No such file or directory")
	}
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

	regex := regexp.MustCompile(`echo|type|exit|pwd`)

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

func runCommand(c Command) {
	cmd := exec.Command(c.name, c.args...)
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

func parseInput(input string) Command {
	input = strings.TrimSuffix(strings.TrimSpace(input), "\n")

	firstSpace := strings.Index(input, " ")
	if firstSpace == -1 {
		return Command{
			name: input,
		}
	}

	return Command{
		name: input[:firstSpace],
		args: strings.Split(input[firstSpace+1:], " "),
	}
}

func getPath() (path string, err error) {
	path, ok := os.LookupEnv("PATH")
	if !ok {
		return "", ErrPathNotFound
	}

	return path, nil
}
