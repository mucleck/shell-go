package main

import (
	"strings"

	"github.com/chzyer/readline"
)

type Completer struct {
	commands []string
}

func (c *Completer) Do(line []rune, pos int) (newLine [][]rune, length int) {
	input := string(line[:pos])

	var matches [][]rune

	for _, command := range c.commands {
		if strings.HasPrefix(command, input) {
			matches = append(matches, []rune(command[pos:]))
		}
	}

	return matches, len([]rune(input))

}

func main() {

	completer := Completer{commands: []string{"echo", "perro"}}
	rl, err := readline.New("> ")
	rl.Config.AutoComplete = &completer
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		println(line)
	}
}
