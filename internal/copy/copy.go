// Package copy provides primitives for seeding an env file by copying it
// from another location. Paths may be absolute, `~`-prefixed, or relative
// to a configured base directory (typically the main worktree root).
package copy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// File copies src to dst atomically (via temp-file-then-rename), preserving
// src's permissions. If dst already exists it is replaced. Callers are
// responsible for deciding whether overwriting is allowed and for backing
// up the previous file if needed.
func File(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source %s: %w", src, err)
	}
	defer in.Close()

	srcInfo, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat source %s: %w", src, err)
	}
	mode := srcInfo.Mode().Perm()
	if mode == 0 {
		mode = 0o600
	}

	dir := filepath.Dir(dst)
	base := filepath.Base(dst)
	tmp, err := os.CreateTemp(dir, fmt.Sprintf(".%s.*.tmp", base))
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	if _, err := io.Copy(tmp, in); err != nil {
		return fmt.Errorf("copying bytes: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("syncing temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp: %w", err)
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return fmt.Errorf("renaming temp to %s: %w", dst, err)
	}
	committed = true
	return nil
}

// ResolveSource turns a user-provided source path into an absolute path.
// Absolute paths are returned as-is (after Clean), `~`-prefixed paths are
// expanded against homeDir, and everything else is resolved relative to
// baseDir (typically the main worktree root).
//
// homeDir may be empty; if so, `~`-prefixed paths return an error.
func ResolveSource(source, baseDir, homeDir string) (string, error) {
	if source == "" {
		return "", fmt.Errorf("copy source is empty")
	}
	if strings.HasPrefix(source, "~") {
		if homeDir == "" {
			return "", fmt.Errorf("cannot expand ~ in %q: home directory unknown", source)
		}
		// Accept both `~` and `~/...`. Disallow `~user/...` — out of scope.
		if source == "~" {
			return filepath.Clean(homeDir), nil
		}
		if strings.HasPrefix(source, "~/") {
			return filepath.Clean(filepath.Join(homeDir, source[2:])), nil
		}
		return "", fmt.Errorf("unsupported ~ form: %q (only ~ and ~/... are supported)", source)
	}
	if filepath.IsAbs(source) {
		return filepath.Clean(source), nil
	}
	if baseDir == "" {
		return "", fmt.Errorf("cannot resolve relative source %q: base directory unknown (not in a git worktree?)", source)
	}
	return filepath.Clean(filepath.Join(baseDir, source)), nil
}
