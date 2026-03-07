package cmd

import (
	"errors"
	"fmt"

	"github.com/AndrewADev/bight/internal/hook"
	"github.com/spf13/cobra"
)

func uninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove post-checkout hook from .git/hooks/",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := hook.Uninstall(); err != nil {
				if errors.Is(err, hook.ErrNotInstalled) {
					fmt.Println("bight: hook not installed, ignoring")
					return nil
				}
				return err
			}
			fmt.Println("bight: hook removed")
			return nil
		},
	}
}
