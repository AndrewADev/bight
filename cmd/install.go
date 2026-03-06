package cmd

import (
	"fmt"

	"github.com/AndrewADev/bight/internal/hook"
	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Write post-checkout hook into .git/hooks/",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hook.Install(); err != nil {
				return err
			}
			fmt.Println("bight: hook installed")
			return nil
		},
	}
}
