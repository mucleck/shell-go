package shell

import (
	"fmt"
	"os"
	"strings"
)

type builtins map[string]func(*Command)

func cdCommand(c *Command) {
	if len(c.args) == 0 || c.args[0] == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(c.redirects.stderr, "%v", err)
		}

		if err := os.Chdir(homeDir); err != nil {
			fmt.Fprintf(c.redirects.stderr, "%v", err)
		}
		return
	}

	if err := os.Chdir(c.args[0]); err != nil {
		fmt.Fprintf(c.redirects.stderr, "cd: %s: No such file or directory\n", c.args[0])
	}
}

func pwdCommand(c *Command) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(c.redirects.stderr, "%v", ErrPwdNotFound)
	}
	fmt.Fprintf(c.redirects.stdout, "%s\n", dir)
}

func typeCommand(c *Command) {
	if len(c.args) == 0 {
		return
	}

	if _, ok := c.shell.builtin[c.args[0]]; ok {
		fmt.Fprintf(c.redirects.stdout, "%s is a shell builtin\n", c.args[0])
		return
	}

	commandPath := c.commandExistsInPath(c.args[0])

	if commandPath != "" {
		fmt.Fprintf(c.redirects.stdout, "%s is %s\n", c.args[0], commandPath)
		return
	}

	fmt.Println(c.args[0] + ": not found")
}

func echoCommand(c *Command) {
	fmt.Fprintf(c.redirects.stdout, "%s\n", strings.Join(c.args, " "))
}
