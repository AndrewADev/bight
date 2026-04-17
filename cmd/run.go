package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AndrewADev/bight/internal/output"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Manually apply env patching for the current branch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			branch, err := resolveBranch(".")
			if err != nil {
				return err
			}

			cfg, err := loadConfig()
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
						fmt.Fprintln(os.Stderr, output.ErrorStderr(fmt.Sprintf("bight (dry-run): var %s: %v", r.varName, r.err)))
						hasErr = true
						continue
					}
					display := r.value
					if r.sensitive {
						display = "***"
					}
					fmt.Printf("bight %s: %s %s %s=%s\n", output.Dim("(dry-run)"), r.path, output.Dim("→"), output.Cyan(r.varName), output.Bold(display))
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
