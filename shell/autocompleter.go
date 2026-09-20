package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type AutoCompleter struct {
	shell      *Shell
	isFirstTab bool
	refresh    func()
}

func (a *AutoCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	input := string(line[:pos])
	var matches []string

	for command := range a.shell.builtin {
		if strings.HasPrefix(command, input) {
			matches = append(matches, command)
		}
	}

	for _, dir := range filepath.SplitList(a.shell.path) {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, f := range files {
			info, err := f.Info()
			if err != nil {
				continue
			}

			if !info.IsDir() &&
				info.Mode().Perm()&0111 != 0 &&
				strings.HasPrefix(f.Name(), input) {
				matches = append(matches, f.Name())
			}
		}
	}

	return a.showMatches(input, matches, pos)
}

func (a *AutoCompleter) showMatches(input string, matches []string, pos int) (newLine [][]rune, length int) {

	slices.Sort(matches)
	matches = slices.Compact(matches)

	switch len(matches) {
	case 0:
		fmt.Print("\a")
		a.isFirstTab = true
	case 1:
		a.isFirstTab = true
		match := strings.TrimPrefix(matches[0], input)

		return [][]rune{
			[]rune(match + " "),
		}, len([]rune(input))
	default:
		//este codigo es muy feo pero basicamente buscamos si en los resultados tenemos
		//alguna palabra común para poder añadirla, por ejemplo si mis resultados al primer tab
		//son foo_bar_xyz y foo_bar_zyx si y le damos al tab con esto se nos añade bar_
		var wordToadd strings.Builder
		for i, c := range matches[0][pos:] {
			if isXinallY(c, i, pos, matches[0:]) {
				wordToadd.WriteRune(c)
				continue
			}
			break
		}

		if wordToadd.Len() > 0 {
			return [][]rune{[]rune(wordToadd.String())}, 0
		}

		if a.isFirstTab {
			a.isFirstTab = false
			fmt.Print("\a")
		}

		fmt.Print("\n" + strings.Join(matches, "  ") + "\n")

		if a.refresh != nil {
			a.refresh()
		}

		a.isFirstTab = true

		wordToadd.Reset()
	}
	return nil, pos
}

func isXinallY(c rune, i, pos int, matches []string) bool {
	for _, match := range matches {
		if match[i+pos] != byte(c) {
			return false
		}
	}
	return true
}
