package hook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AndrewADev/bight/internal/git"
)

const (
	blockBegin = "# bight:begin"
	blockEnd   = "# bight:end"
)

// hookScript checks the binary recorded at install before running it; a
// missing or non-executable binary prints a hint and exits cleanly instead
// of failing the hook.
const hookScript = `#!/bin/sh
` + blockBegin + `
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

// Install writes the post-checkout hook into the repo's hooks directory.
func Install(repo *git.Repo) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	hooksDir := repo.HooksDir()
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
func Check(repo *git.Repo) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	hookPath := filepath.Join(repo.HooksDir(), "post-checkout")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("hook not found")
	}

	if !strings.Contains(string(data), exe) {
		return fmt.Errorf("hook exists but does not reference current binary")
	}
	return nil
}
