package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
	ErrNoToken      = errors.New("Nothing entered or cant parse it")
)

const messageCommandNotFound = "command not found"

type state int

const (
	normalState state = iota
	singleQuoteState
	doubleQuoteState
	backslashState
)

type fds struct {
	stdout *os.File
	stdin  *os.File
	stderr *os.File
}

type Command struct {
	name      string
	args      []string
	redirects fds
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

		command, err := parseInput(input)
		if err != nil {
			continue
		}

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
		typeCommand(c.args, &c.redirects)
	case "echo":
		echoCommand(c.args, &c.redirects)
	case "pwd":
		pwdCommand(&c.redirects)
	case "cd":
		cdCommand(c.args, &c.redirects)
	}
}

func cdCommand(args []string, c *fds) {
	if len(args) == 0 || args[0] == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Println(err)
		}

		if err := os.Chdir(homeDir); err != nil {
			fmt.Fprintf(c.stderr, "%v", err)
		}
		return
	}

	if err := os.Chdir(args[0]); err != nil {
		fmt.Fprintf(c.stderr, "cd: %s: No such file or directory\n", args[0])
	}
}

func pwdCommand(c *fds) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(c.stderr, "%v", ErrPwdNotFound)
	}
	fmt.Fprintf(c.stdout, "%s\n", dir)
}

func typeCommand(args []string, c *fds) {

	if len(args) == 0 {
		return
	}

	regex := regexp.MustCompile(`echo|type|exit|pwd`)

	if regex.MatchString(args[0]) {
		fmt.Fprintf(c.stdout, "%s is a shell builtin\n", args[0])
		return
	}

	commandPath := commandExistsInPath(args[0])

	if commandPath != "" {
		fmt.Fprintf(c.stdout, "%s is %s\n", args[0], commandPath)
		return
	}

	fmt.Println(args[0] + ": not found")
}

func echoCommand(args []string, c *fds) {
	fmt.Fprintf(c.stdout, "%s\n", strings.Join(args, " "))
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
	cmd.Stdin, cmd.Stdout, cmd.Stderr = c.redirects.stdin, c.redirects.stdout, c.redirects.stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(cmd.Stderr, err)
	}
}

func readInput() (command *bufio.Reader, err error) {
	command = bufio.NewReader(os.Stdin)
	return command, nil
}

/*
<    → stdin (fd 0)
0<   → stdin (fd 0)

>    → stdout (fd 1)
1>   → stdout (fd 1)
>>   → stdout (fd 1)
1>>  → stdout (fd 1)

2>   → stderr (fd 2)
2>>  → stderr (fd 2)
*/

func parseInput(input *bufio.Reader) (Command, error) {
	tokens := getTokens(input)

	if len(tokens) == 0 {
		return Command{}, ErrNoToken
	}

	return getCommandFromTokens(tokens), nil
}
func getCommandFromTokens(tokens []string) Command {
	var args []string
	var redirects fds

	for i := 1; i < len(tokens); i++ {
		word := tokens[i]
		switch {
		case strings.HasSuffix(word, ">"):
			generateStdoutOrStderr(word, tokens[i+1], &redirects)
			i++
		case strings.HasSuffix(word, "<"):
			// one day i do this dw
		default:
			args = append(args, word)
		}
	}

	if redirects.stdin == nil {
		redirects.stdin = os.Stdin
	}

	if redirects.stdout == nil {
		redirects.stdout = os.Stdout
	}

	if redirects.stderr == nil {
		redirects.stderr = os.Stderr
	}

	return Command{
		name:      tokens[0],
		args:      args,
		redirects: redirects,
	}
}

func generateStdoutOrStderr(stdin, file string, redirects *fds) {
	var mode int

	switch strings.Count(stdin, ">") {
	case 1:
		mode = os.O_TRUNC
	case 2:
		mode = os.O_APPEND
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|mode, 0644)
	if err != nil {
		panic(err)
	}

	prefix := stdin[0]
	switch prefix {
	case '1', '>':
		redirects.stdout = f
	case '2':
		redirects.stderr = f
	}
}

func getTokens(input *bufio.Reader) (tokens []string) {
	var word strings.Builder
	var previousState state
	state := normalState
	for {
		c, _, err := input.ReadRune()

		if err == io.EOF || c == '\n' {
			if word.Len() > 0 {
				tokens = append(tokens, word.String())
			}
			break
		} else if err != nil {
			panic(err)
		}

		switch state {
		case normalState:
			switch c {
			case '\\':
				state = backslashState
			case ' ':
				if word.Len() > 0 {
					tokens = append(tokens, word.String())
					word.Reset()
				}
				continue

			case '\'':
				state = singleQuoteState
			case '"':
				state = doubleQuoteState
			case '>':
				next, _, err := input.ReadRune()
				if err != nil {
					panic(err)
				}

				if next == '>' {
					word.WriteRune('>')
				} else {
					input.UnreadRune()
				}

				word.WriteRune('>')
				tokens = append(tokens, word.String())
				word.Reset()
			case '<':
				word.WriteRune(c)
				tokens = append(tokens, word.String())
				word.Reset()
			default:
				word.WriteRune(c)
			}

		case singleQuoteState:
			if c == '\'' {
				state = normalState
			} else {
				word.WriteRune(c)
			}
		case doubleQuoteState:
			switch c {
			case '"':
				state = normalState
			case '\\':
				previousState = state
				state = backslashState
			default:
				word.WriteRune(c)
			}
		case backslashState:
			word.WriteRune(c)
			state = previousState
		}
	}

	return tokens
}

func getPath() (path string, err error) {
	path, ok := os.LookupEnv("PATH")
	if !ok {
		return "", ErrPathNotFound
	}

	return path, nil
}
