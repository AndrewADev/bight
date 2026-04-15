package main

import (
	"fmt"
	"os"

	"github.com/AndrewADev/bight/cmd"
	"github.com/AndrewADev/bight/internal/output"
)

func main() {
	if err := cmd.Root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, output.ErrorStderr("Error: "+err.Error()))
		os.Exit(1)
	}
}
