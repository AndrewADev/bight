package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AndrewADev/bight/internal/config"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Manually apply env patching for the current branch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			branch, err := resolveBranch()
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("no .bight.yml or .bight.yaml found in current directory")
			}
			if err != nil {
				return err
			}

			return patchEnvFiles(cfg, branch)
		},
	}
}
