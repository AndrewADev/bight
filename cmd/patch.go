package cmd

import (
	"fmt"

	"github.com/AndrewADev/bight/internal/config"
	"github.com/AndrewADev/bight/internal/env"
	"github.com/AndrewADev/bight/internal/strategy"
)

type dryRunResult struct {
	path    string
	varName string
	value   string
	err     error
}

func dryRunEnvFiles(cfg *config.Config, branch string) []dryRunResult {
	ctx := strategy.Context{Branch: branch, Project: cfg.Project}
	var results []dryRunResult

	for _, ef := range cfg.EnvFiles {
		for _, v := range ef.Vars {
			if v.On != "checkout" {
				continue
			}
			val, err := strategy.Apply(v.Strategy, ctx, cfg)
			results = append(results, dryRunResult{
				path:    ef.Path,
				varName: v.Name,
				value:   val,
				err:     err,
			})
		}
	}
	return results
}

func patchEnvFiles(cfg *config.Config, branch string) error {
	ctx := strategy.Context{
		Branch:  branch,
		Project: cfg.Project,
	}

	for _, ef := range cfg.EnvFiles {
		patches := make(map[string]string)
		for _, v := range ef.Vars {
			if v.On != "checkout" {
				continue
			}
			val, err := strategy.Apply(v.Strategy, ctx, cfg)
			if err != nil {
				return fmt.Errorf("var %s: %w", v.Name, err)
			}
			patches[v.Name] = val
		}

		comments, err := env.ScanComments(ef.Path, cfg.Defaults.CollectComments)
		if err != nil {
			return fmt.Errorf("scanning %s: %w", ef.Path, err)
		}

		if err := env.PatchAll(ef.Path, patches, comments); err != nil {
			return fmt.Errorf("patching %s: %w", ef.Path, err)
		}

		for name, val := range patches {
			fmt.Printf("bight: %s → %s=%s\n", ef.Path, name, val)
		}
	}
	return nil
}
