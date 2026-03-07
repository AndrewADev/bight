package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AndrewADev/bight/internal/config"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
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

			if dryRun {
				results := dryRunEnvFiles(cfg, branch)
				var hasErr bool
				for _, r := range results {
					if r.err != nil {
						fmt.Fprintf(os.Stderr, "bight (dry-run): var %s: %v\n", r.varName, r.err)
						hasErr = true
						continue
					}
					fmt.Printf("bight (dry-run): %s → %s=%s\n", r.path, r.varName, r.value)
				}
				if hasErr {
					return fmt.Errorf("dry-run completed with errors")
				}
				return nil
			}

			return patchEnvFiles(cfg, branch)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be written without modifying any files")
	return cmd
}
