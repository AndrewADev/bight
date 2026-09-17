package hook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrNotInstalled = errors.New("bight: hook not installed")

func Uninstall() error {
	return uninstall(filepath.Join(".git", "hooks", "post-checkout"))
}

func uninstall(hookPath string) error {

	info, err := os.Stat(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotInstalled
		}
		return fmt.Errorf("reading hook: %w", err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("reading hook: %w", err)
	}

	filtered, found := stripHookBlock(strings.Split(string(data), "\n"))
	if !found {
		return ErrNotInstalled
	}

	// If nothing meaningful remains beyond a shebang, remove the file entirely.
	meaningful := false
	for _, line := range filtered {
		t := strings.TrimSpace(line)
		if t != "" && t != "#!/bin/sh" {
			meaningful = true
			break
		}
	}
	if !meaningful {
		return os.Remove(hookPath)
	}

	return os.WriteFile(hookPath, []byte(strings.Join(filtered, "\n")), info.Mode())
}

// stripHookBlock removes bight's lines from a hook script: the marker-delimited
// block written by current installs, or the single invocation line written by
// installs before the markers existed. The second return value reports whether
// anything was removed.
func stripHookBlock(lines []string) ([]string, bool) {
	filtered := make([]string, 0, len(lines))
	found := false
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == blockBegin:
			inBlock = true
			found = true
		case trimmed == blockEnd:
			inBlock = false
		case inBlock:
		case strings.Contains(line, " post-checkout "):
			found = true
		default:
			filtered = append(filtered, line)
		}
	}
	return filtered, found
}
