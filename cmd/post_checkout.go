package cmd

import (
	"errors"
	"os"

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

			repo, err := enterRepo()
			if err != nil {
				return err
			}

			cfg, _, _, err := loadConfig(repo.MainRoot())
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			if err != nil {
				return err
			}

			return patchEnvFiles(cfg, repo.Branch, repo.MainRoot())
		},
	}
}
