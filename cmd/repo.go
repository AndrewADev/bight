package cmd

import (
	"os"
	"path/filepath"

	"github.com/AndrewADev/bight/internal/git"
)

// enterRepo discovers the repository enclosing the working directory and
// changes into its root. Relative --config and BIGHT_CONFIG paths are made
// absolute against the original working directory first.
func enterRepo() (*git.Repo, error) {
	repo, err := git.Discover(".")
	if err != nil {
		return nil, err
	}
	if configPath != "" {
		abs, err := filepath.Abs(configPath)
		if err != nil {
			return nil, err
		}
		configPath = abs
	}
	if p := os.Getenv(envConfigPath); p != "" && !filepath.IsAbs(p) {
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		if err := os.Setenv(envConfigPath, abs); err != nil {
			return nil, err
		}
	}
	if err := os.Chdir(repo.Root); err != nil {
		return nil, err
	}
	return repo, nil
}
