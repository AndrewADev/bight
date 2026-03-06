package cmd

import (
	"fmt"

	"github.com/AndrewADev/bight/internal/config"
	"github.com/AndrewADev/bight/internal/env"
	"github.com/AndrewADev/bight/internal/strategy"
)

func patchEnvFiles(cfg *config.Config, branch string) error {
	ctx := strategy.Context{
		Branch:  branch,
		Project: cfg.Project,
	}

	for _, ef := range cfg.EnvFiles {
		for _, v := range ef.Vars {
			if v.On != "checkout" {
				continue
			}

			val, err := strategy.Apply(v.Strategy, ctx, cfg)
			if err != nil {
				return fmt.Errorf("var %s: %w", v.Name, err)
			}

			if err := env.Patch(ef.Path, v.Name, val); err != nil {
				return fmt.Errorf("patching %s: %w", ef.Path, err)
			}

			fmt.Printf("bight: %s → %s=%s\n", ef.Path, v.Name, val)
		}
	}
	return nil
}
