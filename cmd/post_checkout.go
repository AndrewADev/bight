package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func postCheckoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "post-checkout <prev> <new> <flag>",
		Short: "Called by git post-checkout hook",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// flag == "0" means file checkout, not branch checkout
			if args[2] == "0" {
				return nil
			}

			branch, err := resolveBranch(".")
			if err != nil {
				return err
			}

			cfg, err := loadConfig()
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			if err != nil {
				return err
			}

			return patchEnvFiles(cfg, branch)
		},
	}
}

func resolveBranch(dir string) (string, error) {
	gitDir, err := resolveGitDir(dir)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "", fmt.Errorf("reading HEAD: %w", err)
	}
	head := strings.TrimSpace(string(data))
	const prefix = "ref: refs/heads/"
	if strings.HasPrefix(head, prefix) {
		return head[len(prefix):], nil
	}
	// detached HEAD — return commit hash as-is
	return head, nil
}

// resolveGitDir returns the path to the actual git directory.
// In a worktree, .git is a file containing "gitdir: <path>" rather than a directory.
func resolveGitDir(dir string) (string, error) {
	dotGit := filepath.Join(dir, ".git")
	info, err := os.Stat(dotGit)
	if err != nil {
		return "", fmt.Errorf("stat .git: %w", err)
	}
	if info.IsDir() {
		return dotGit, nil
	}
	// worktree case: .git is a file pointing to the real git dir
	data, err := os.ReadFile(dotGit)
	if err != nil {
		return "", fmt.Errorf("reading .git: %w", err)
	}
	line := strings.TrimSpace(string(data))
	const prefix = "gitdir: "
	if !strings.HasPrefix(line, prefix) {
		return "", fmt.Errorf("unexpected .git file content: %q", line)
	}
	return line[len(prefix):], nil
}
