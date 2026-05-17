package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const hookScript = "#!/bin/sh\n%s post-checkout \"$@\"\n"

// HooksDir returns the path to the git hooks directory for the repo at the
// current working directory. In a regular repo this is .git/hooks; in a
// worktree it resolves to the main repo's hooks dir via commondir.
func HooksDir() (string, error) {
	return hooksDir(".")
}

func hooksDir(dir string) (string, error) {
	commonDir, err := commonGitDir(dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(commonDir, "hooks"), nil
}

// MainWorktreeRoot returns the working-tree directory of the main worktree
// for the repo at the current working directory. In a regular repo this is
// the same directory; in a linked worktree it resolves to the main repo's
// working directory via commondir.
func MainWorktreeRoot() (string, error) {
	return mainWorktreeRoot(".")
}

func mainWorktreeRoot(dir string) (string, error) {
	commonDir, err := commonGitDir(dir)
	if err != nil {
		return "", err
	}
	// commonDir points to the main repo's .git directory; its parent is the
	// main worktree's working directory.
	return filepath.Dir(commonDir), nil
}

// commonGitDir returns the path to the main repo's .git directory (the
// "common" git dir), resolved relative to the directory `dir`. For a
// regular repo this is just <dir>/.git; for a worktree it's the directory
// pointed at by the worktree's commondir file.
func commonGitDir(dir string) (string, error) {
	dotGit := filepath.Join(dir, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return "", fmt.Errorf(".git not found — are you in a git repo?")
	}

	if info.IsDir() {
		return dotGit, nil
	}

	// .git is a file — we're in a worktree. Format: "gitdir: <path>"
	data, err := os.ReadFile(dotGit)
	if err != nil {
		return "", fmt.Errorf("reading .git: %w", err)
	}
	line := strings.TrimSpace(string(data))
	const prefix = "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf(".git file has unexpected format")
	}
	worktreeGitDir := strings.TrimPrefix(line, prefix)

	// commondir holds a path (possibly relative) to the main git dir.
	commonDirData, err := os.ReadFile(filepath.Join(worktreeGitDir, "commondir"))
	if err != nil {
		return "", fmt.Errorf("reading commondir: %w", err)
	}
	commonDir := strings.TrimSpace(string(commonDirData))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(worktreeGitDir, commonDir)
	}

	return filepath.Clean(commonDir), nil
}

func Install() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	hooksDir, err := HooksDir()
	if err != nil {
		return err
	}
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return fmt.Errorf("%s not found — are you in a git repo?", hooksDir)
	}

	hookPath := filepath.Join(hooksDir, "post-checkout")
	content := fmt.Sprintf(hookScript, exe)
	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("writing hook: %w", err)
	}

	return nil
}

// Check returns an error if the post-checkout hook is not installed or does
// not reference the current binary path.
func Check() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	hooksDir, err := HooksDir()
	if err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "post-checkout")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("hook not found")
	}

	if !strings.Contains(string(data), exe) {
		return fmt.Errorf("hook exists but does not reference current binary")
	}
	return nil
}
