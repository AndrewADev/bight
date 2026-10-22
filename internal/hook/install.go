package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	blockBegin = "# bight:begin"
	blockEnd   = "# bight:end"
	shebang    = "#!/bin/sh"
)

// hookBlock is the marker-delimited section bight owns inside the hook. It
// checks the binary recorded at install before running it; a missing or
// non-executable binary prints a hint and exits cleanly instead of failing
// the hook.
const hookBlock = blockBegin + `
BIGHT="%s"
if [ ! -e "$BIGHT" ]; then
  echo "bight: $BIGHT not found; run 'bight install' again" >&2
  exit 0
fi
if [ ! -x "$BIGHT" ]; then
  echo "bight: $BIGHT is not executable" >&2
  exit 0
fi
"$BIGHT" post-checkout "$@"
` + blockEnd + `
`

// hookScript is the complete hook written when no post-checkout hook exists.
const hookScript = shebang + "\n" + hookBlock

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

	return install(filepath.Join(hooksDir, "post-checkout"), exe)
}

// install writes bight's block into the hook at hookPath. A missing hook is
// created from hookScript. An existing hook keeps its shebang, other content
// and file mode; any prior bight block or legacy invocation line is removed
// before the block is appended.
func install(hookPath, exe string) error {
	info, err := os.Stat(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			content := fmt.Sprintf(hookScript, exe)
			if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
				return fmt.Errorf("writing hook: %w", err)
			}
			return nil
		}
		return fmt.Errorf("reading hook: %w", err)
	}

	data, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("reading hook: %w", err)
	}

	filtered, _ := stripHookBlock(strings.Split(string(data), "\n"))
	existing := strings.TrimRight(strings.Join(filtered, "\n"), "\n")

	var b strings.Builder
	if !strings.HasPrefix(existing, "#!") {
		b.WriteString(shebang)
		b.WriteString("\n")
	}
	if existing != "" {
		b.WriteString(existing)
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, hookBlock, exe)

	if err := os.WriteFile(hookPath, []byte(b.String()), info.Mode()); err != nil {
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
