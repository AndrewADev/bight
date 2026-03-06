package hook

import (
	"fmt"
	"os"
	"path/filepath"
)

const hookScript = "#!/bin/sh\n%s post-checkout \"$@\"\n"

func Install() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	hooksDir := filepath.Join(".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		return fmt.Errorf(".git/hooks/ not found — are you in a git repo?")
	}

	hookPath := filepath.Join(hooksDir, "post-checkout")
	content := fmt.Sprintf(hookScript, exe)
	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("writing hook: %w", err)
	}

	return nil
}
