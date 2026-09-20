package shell

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

type Shell struct {
	path          string
	prompt        string
	builtin       builtins
	autocompleter *AutoCompleter
}

type Command struct {
	shell     *Shell
	name      string
	args      []string
	redirects fds
}

type fds struct {
	stdout *os.File
	stdin  *os.File
	stderr *os.File
}

var (
	ErrPathNotFound = errors.New("cannot find path")
	ErrPwdNotFound  = errors.New("cant get current working dir, weird")
	ErrNoToken      = errors.New("Nothing entered or cant parse it")
)

const messageCommandNotFound = "command not found"

func New() *Shell {
	s := Shell{
		path:   os.Getenv("PATH"),
		prompt: "$ ",
		builtin: builtins{
			"echo": echoCommand,
			"type": typeCommand,
			"pwd":  pwdCommand,
			"cd":   cdCommand,
			"exit": func(*Command) {},
		},
	}

	s.autocompleter = &AutoCompleter{shell: &s, isFirstTab: true}
	return &s
}

func (s *Shell) Run() error {
	rlConfig := &readline.Config{
		Prompt:       s.prompt,
		AutoComplete: s.autocompleter,
	}

	rlConfig.SetListener(func(line []rune, pos int, key rune) ([]rune, int, bool) {
		if key != '\t' {
			s.autocompleter.isFirstTab = true
		}
		return line, pos, false
	})

	rl, err := readline.NewEx(rlConfig)
	if err != nil {
		panic(err)
	}

	defer rl.Close()

	s.autocompleter.refresh = rl.Refresh

	for {
		input, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		name, args, redirects, err := parse(strings.NewReader(input))
		if err != nil {
			log.Println(err)
			continue
		}

		command := Command{
			shell:     s,
			name:      name,
			args:      args,
			redirects: redirects,
		}

		if handleCommand(command) {
			break
		}
	}
	return nil
}

func (s *Shell) isBuiltin(name string) bool {
	_, ok := s.builtin[name]
	return ok
}

func (c *Command) commandExistsInPath(name string) string {
	directories := strings.SplitSeq(c.shell.path, string(os.PathListSeparator))
	for dir := range directories {
		commandWithDir := filepath.Join(dir, name)
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

func handleCommand(c Command) bool {
	switch {
	case c.name == "exit":
		return true
	case c.shell.isBuiltin(c.name):
		c.shell.builtin[c.name](&c)
	case c.commandExistsInPath(c.name) != "":
		c.runCommand()
	default:
		fmt.Println(c.name + ": " + messageCommandNotFound)
	}
	return false
}

func (c *Command) runCommand() {
	cmd := exec.Command(c.name, c.args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = c.redirects.stdin, c.redirects.stdout, c.redirects.stderr
	if err := cmd.Run(); err != nil {
		//
	}
}
