package cmd

import "github.com/spf13/cobra"

func Root() *cobra.Command {
	root := &cobra.Command{
		Use:     "bight",
		Short:   "Patch .env files on git branch checkout",
		Version: resolveVersion(),
	}
	root.AddCommand(installCmd(), uninstallCmd(), postCheckoutCmd(), runCmd(), doctorCmd())
	return root
}
