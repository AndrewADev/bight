package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// promptReader is the source confirm() and other prompts read answers
// from. Tests swap it via setPromptReader.
var promptReader = bufio.NewReader(os.Stdin)

// promptOut is where the question is written. Tests swap it.
var promptOut io.Writer = os.Stdout

// isPromptTTY reports whether stdin is connected to a terminal. Tests
// swap it to force the TTY / non-TTY branch independently of promptReader.
var isPromptTTY = defaultIsPromptTTY

func defaultIsPromptTTY() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// confirm prints question with a [Y/n] / [y/N] suffix and reads a y/n
// answer from stdin. If stdin is not a TTY, the question is echoed for
// visibility and def is returned without blocking. An empty line, EOF,
// or an unrecognized answer also yields def.
func confirm(question string, def bool) bool {
	suffix := "[Y/n]"
	if !def {
		suffix = "[y/N]"
	}

	if !isPromptTTY() {
		fmt.Fprintf(promptOut, "%s %s\n", question, suffix)
		return def
	}

	fmt.Fprintf(promptOut, "%s %s ", question, suffix)
	line, err := promptReader.ReadString('\n')
	if err != nil && line == "" {
		return def
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return def
	}
}
