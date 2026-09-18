package shell

import (
	"io"
	"os"
	"strings"
)

type state int

const (
	normalState state = iota
	singleQuoteState
	doubleQuoteState
	backslashState
)

func parse(input io.RuneScanner) (name string, args []string, redirects fds, err error) {
	tokens, err := getTokens(input)
	if err != nil {
		return name, args, redirects, err
	}

	if len(tokens) == 0 {
		return name, args, redirects, ErrNoToken
	}

	return getCommandFromTokens(tokens)
}

func getTokens(input io.RuneScanner) (tokens []string, err error) {
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
			return tokens, err
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
					return tokens, err
				}

				if next == '>' {
					word.WriteRune('>')
				} else {
					err = input.UnreadRune()
					if err != nil {
						return tokens, err
					}
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

	return tokens, nil
}

func getCommandFromTokens(tokens []string) (name string, args []string, redirects fds, err error) {

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

	return tokens[0], args, redirects, nil
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
