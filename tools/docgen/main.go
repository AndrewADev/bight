package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AndrewADev/bight/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	outDir := "docs/commands"
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	root := cmd.Root()
	root.DisableAutoGenTag = true

	if err := doc.GenMarkdownTree(root, outDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("wrote command docs to %s\n", filepath.Clean(outDir))
}
