// Package shell is a package
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

/*
*	Everything here is really bad done... I will refactor all the auto completition logic...
*	Ill  make two points to mark here to be done
*	-[ ] Stablish the stuff we want to acommplis, this doesnt just mean to pass the test. It really means that any design for this should be
*				 a solid solition instead of a passing tests guesser
*	-[ ] Do the job on small functions
* */

func (a *AutoCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	//matches := getMatches(input)
	//candidates := filterMatches(matches)

	input := string(line[:pos])
	var matches []string
	for command := range a.shell.builtin {
		if strings.HasPrefix(command, input) {
			matches = append(matches, command)
		}
	}

	var isDir bool
	directory := a.shell.path

	spacePos := strings.Index(input, " ") // cambiar por cut
	if spacePos != -1 {
		//directory, _ = os.Getwd()
		input = input[spacePos+1:]
		if strings.HasSuffix(input, string(filepath.Separator)) {
			directory = filepath.Clean(input)
			input = ""
		} else {
			dir := filepath.Dir(input)
			name := filepath.Base(input)
			info, err := os.Stat(dir)

			if err == nil && info.IsDir() {
				directory = dir
				input = name
			} else {
				directory, _ = os.Getwd()
			}
		}
		pos = len(input)
	}

	for _, dir := range filepath.SplitList(directory) {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, f := range files {
			info, err := f.Info()
			if err != nil {
				continue
			}

			if strings.HasPrefix(f.Name(), input) {
				if spacePos != -1 || info.Mode().Perm()&0o111 != 0 {
					if info.IsDir() {
						isDir = true
						matches = append(matches, f.Name()+"/")
						pos++
					} else {
						matches = append(matches, f.Name())
					}
				}
			} else if input == "." {

				isDir = true
				matches = append(matches, f.Name()+"/")
			}
		}
	}

	return a.showMatches(input, matches, pos, isDir)
}

func (a *AutoCompleter) showMatches(input string, matches []string, pos int, isDir bool) (newLine [][]rune, length int) {
	slices.Sort(matches)
	matches = slices.Compact(matches)

	switch len(matches) {
	case 0:
		fmt.Print("\a")
		a.isFirstTab = true
	case 1:
		a.isFirstTab = true
		match := strings.TrimPrefix(matches[0], input)

		if isDir {
			return [][]rune{
				[]rune(match),
			}, len([]rune(input))
		}
		return [][]rune{
			[]rune(match + " "),
		}, len([]rune(input))
	default:
		// este codigo es muy feo pero basicamente buscamos si en los resultados tenemos
		// alguna palabra común para poder añadirla, por ejemplo si mis resultados al primer tab
		// son foo_bar_xyz y foo_bar_zyx si y le damos al tab con esto se nos añade bar_
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
