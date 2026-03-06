package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AndrewADev/bight/internal/config"
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

			branch, err := resolveBranch()
			if err != nil {
				return err
			}

			cfg, err := config.Load(".bight.yml")
			if err != nil {
				return err
			}

			return patchEnvFiles(cfg, branch)
		},
	}
}

func resolveBranch() (string, error) {
	data, err := os.ReadFile(".git/HEAD")
	if err != nil {
		return "", fmt.Errorf("reading .git/HEAD: %w", err)
	}
	head := strings.TrimSpace(string(data))
	const prefix = "ref: refs/heads/"
	if strings.HasPrefix(head, prefix) {
		return head[len(prefix):], nil
	}
	// detached HEAD — return commit hash as-is
	return head, nil
}
