package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// withStubPrompt swaps the package-level prompt I/O for the duration of a
// test and restores it on cleanup. tty controls whether isPromptTTY reports
// a terminal, and input is the bytes confirm/promptReader will read.
func withStubPrompt(t *testing.T, tty bool, input string) *bytes.Buffer {
	t.Helper()
	origIn := promptReader
	origOut := promptOut
	origTTY := isPromptTTY

	var out bytes.Buffer
	promptReader = bufio.NewReader(strings.NewReader(input))
	promptOut = &out
	isPromptTTY = func() bool { return tty }

	t.Cleanup(func() {
		promptReader = origIn
		promptOut = origOut
		isPromptTTY = origTTY
	})
	return &out
}

func TestConfirm_NonTTYReturnsDefault(t *testing.T) {
	out := withStubPrompt(t, false, "n\n")
	if !confirm("Seed?", true) {
		t.Errorf("non-TTY: expected default=true to be returned")
	}
	if !strings.Contains(out.String(), "Seed?") {
		t.Errorf("non-TTY: expected question to be echoed, got %q", out.String())
	}
}

func TestConfirm_EmptyLineReturnsDefault(t *testing.T) {
	withStubPrompt(t, true, "\n")
	if !confirm("Seed?", true) {
		t.Errorf("empty input with def=true should yield true")
	}
}

func TestConfirm_ParsesYesNo(t *testing.T) {
	cases := []struct {
		input string
		def   bool
		want  bool
	}{
		{"y\n", false, true},
		{"Y\n", false, true},
		{"yes\n", false, true},
		{"YES\n", false, true},
		{"n\n", true, false},
		{"N\n", true, false},
		{"no\n", true, false},
		{"  y  \n", false, true},
		{"garbage\n", true, true},   // unrecognized → def
		{"garbage\n", false, false}, // unrecognized → def
	}
	for _, c := range cases {
		t.Run(strings.TrimSpace(c.input), func(t *testing.T) {
			withStubPrompt(t, true, c.input)
			got := confirm("?", c.def)
			if got != c.want {
				t.Errorf("input=%q def=%v: got %v, want %v", c.input, c.def, got, c.want)
			}
		})
	}
}

func TestConfirm_EOFReturnsDefault(t *testing.T) {
	withStubPrompt(t, true, "")
	if !confirm("?", true) {
		t.Errorf("EOF with def=true should yield true")
	}
}
