package main

import (
	"github.com/codecrafters-io/shell-starter-go/shell"
)

func main() {
	shell := shell.New()
	if err := shell.Run(); err != nil {
		panic(err)
	}
}
