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

	lines := strings.Split(string(data), "\n")
	filtered := make([]string, 0, len(lines))
	found := false
	for _, line := range lines {
		if strings.Contains(line, " post-checkout ") {
			found = true
			continue
		}
		filtered = append(filtered, line)
	}
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
