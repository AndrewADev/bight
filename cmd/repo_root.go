package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// findRepoRoot walks up from dir to the nearest directory containing a .git
// entry (directory or worktree file) and returns its absolute path.
func findRepoRoot(dir string) (string, error) {
	current, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		_, err := os.Stat(filepath.Join(current, ".git"))
		if err == nil {
			return current, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat %s: %w", filepath.Join(current, ".git"), err)
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf(".git not found — are you in a git repo?")
		}
		current = parent
	}
}

// chdirToRepoRoot changes the working directory to the enclosing repo root.
// Relative --config and BIGHT_CONFIG paths are made absolute against the
// original working directory first.
func chdirToRepoRoot() error {
	root, err := findRepoRoot(".")
	if err != nil {
		return err
	}
	if configPath != "" {
		abs, err := filepath.Abs(configPath)
		if err != nil {
			return err
		}
		configPath = abs
	}
	if p := os.Getenv(envConfigPath); p != "" && !filepath.IsAbs(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		if err := os.Setenv(envConfigPath, abs); err != nil {
			return err
		}
	}
	return os.Chdir(root)
}
