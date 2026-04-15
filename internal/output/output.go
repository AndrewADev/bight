package output

import "os"

var stdoutColor = isColorTerminal(os.Stdout)
var stderrColor = isColorTerminal(os.Stderr)

func isColorTerminal(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	cyan   = "\033[36m"
)

func colorize(enabled bool, code, s string) string {
	if !enabled {
		return s
	}
	return code + s + reset
}

// Stdout-bound helpers.
func Bold(s string) string   { return colorize(stdoutColor, bold, s) }
func Green(s string) string  { return colorize(stdoutColor, bold+green, s) }
func Cyan(s string) string   { return colorize(stdoutColor, cyan, s) }
func Yellow(s string) string { return colorize(stdoutColor, yellow, s) }
func Red(s string) string    { return colorize(stdoutColor, bold+red, s) }
func Dim(s string) string    { return colorize(stdoutColor, dim, s) }

// Stderr-bound helpers (for warnings/errors written to os.Stderr).
func WarnStderr(s string) string  { return colorize(stderrColor, yellow, s) }
func ErrorStderr(s string) string { return colorize(stderrColor, bold+red, s) }
