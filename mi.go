package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type state int

const (
	normalState state = iota
	singleQuoteState
)

func main() {

	input := bufio.NewReader(os.Stdin)
	for {
		var args []string
		var word strings.Builder
		state := normalState
		for {
			c, _, err := input.ReadRune()

			if err == io.EOF || c == '\n' {
				if word.Len() > 0 {
					args = append(args, word.String())
				}
				break
			} else if err != nil {
				panic(err)
			}

			switch state {
			case normalState:
				if c == ' ' {
					if word.Len() > 0 {
						args = append(args, word.String())
						word.Reset()
					}
					continue
				}

				if c == '\'' {
					state = singleQuoteState
					continue
				}

				word.WriteRune(c)
			case singleQuoteState:
				if c == '\'' {
					state = normalState
				} else {
					word.WriteRune(c)
				}
			}
		}
	}
}
