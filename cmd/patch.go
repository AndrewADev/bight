package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AndrewADev/bight/internal/config"
	bcopy "github.com/AndrewADev/bight/internal/copy"
	"github.com/AndrewADev/bight/internal/env"
	"github.com/AndrewADev/bight/internal/hook"
	"github.com/AndrewADev/bight/internal/output"
	"github.com/AndrewADev/bight/internal/strategy"
)

// triggerCheckout is the only var-level event bight currently dispatches on.
// Vars with `on:` set to any other value are skipped.
const triggerCheckout = "checkout"

type dryRunResult struct {
	path      string
	varName   string
	value     string
	sensitive bool
	err       error
}

// shouldCopy reports whether ef should be (re)seeded from ef.Copy this run,
// given whether the destination existed before bight started.
func shouldCopy(ef config.EnvFile, existedBefore bool) bool {
	if ef.Copy == nil {
		return false
	}
	return !existedBefore || ef.Copy.Overwrite
}

func dryRunEnvFiles(cfg *config.Config, branch string) []dryRunResult {
	ctx := strategy.Context{Branch: branch, Project: cfg.Project}
	var results []dryRunResult

	for _, ef := range cfg.EnvFiles {
		for _, v := range ef.Vars {
			if v.On != triggerCheckout {
				continue
			}
			val, err := strategy.Apply(v.Strategy, ctx, cfg)
			results = append(results, dryRunResult{
				path:      ef.Path,
				varName:   v.Name,
				value:     val,
				sensitive: v.Sensitive,
				err:       err,
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

	// Resolve the main worktree root lazily — only needed when an env_file
	// has a `copy:` configured.
	var (
		mainRoot    string
		mainRootErr error
		mainOnce    bool
	)
	resolveMain := func() (string, error) {
		if !mainOnce {
			mainRoot, mainRootErr = hook.MainWorktreeRoot()
			mainOnce = true
		}
		return mainRoot, mainRootErr
	}

	for _, ef := range cfg.EnvFiles {
		_, statErr := os.Stat(ef.Path)
		existedBefore := statErr == nil
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("stat %s: %w", ef.Path, statErr)
		}

		willCopy := shouldCopy(ef, existedBefore)

		// Gather patches first so the "nothing to do" check below is honest:
		// if no var matches the checkout event AND no copy is queued, this
		// env_file gets skipped entirely (no backup, no PatchAll write).
		patches := make(map[string]string)
		sensitiveVars := make(map[string]bool)
		for _, v := range ef.Vars {
			if v.On != triggerCheckout {
				continue
			}
			val, err := strategy.Apply(v.Strategy, ctx, cfg)
			if err != nil {
				return fmt.Errorf("var %s: %w", v.Name, err)
			}
			patches[v.Name] = val
			sensitiveVars[v.Name] = v.Sensitive
		}

		if !willCopy && len(patches) == 0 {
			// Nothing will modify this file. The `len(patches) == 0` half of
			// this guard also protects against the v0.1.0 bug where an empty
			// patches map collapsed the file to "\n" via PatchAll.
			continue
		}

		// Single backup point per file, before any mutation. BackupFile is
		// itself a no-op when the source file doesn't exist, so no extra
		// gating needed for the worktree-init / fresh-file case.
		if ef.Backup {
			if err := env.BackupFile(ef.Path); err != nil {
				return fmt.Errorf("backup %s: %w", ef.Path, err)
			}
		}

		if willCopy {
			// resolveMain is best-effort: absolute and ~-prefixed sources
			// don't need it. ResolveSource will error if a relative source
			// is given without a usable base.
			base, _ := resolveMain()
			homeDir, _ := os.UserHomeDir()
			src, err := bcopy.ResolveSource(ef.Copy.Source, base, homeDir)
			if err != nil {
				return fmt.Errorf("copy %s: %w", ef.Path, err)
			}
			if err := bcopy.File(src, ef.Path); err != nil {
				return fmt.Errorf("copy %s ← %s: %w", ef.Path, src, err)
			}
			fmt.Printf("bight: %s %s %s\n", ef.Path, output.Dim("←"), output.Cyan(src))
		}

		if len(patches) > 0 {
			comments, err := env.ScanComments(ef.Path, cfg.Defaults.CollectComments)
			if err != nil {
				return fmt.Errorf("scanning %s: %w", ef.Path, err)
			}
			if err := env.PatchAll(ef.Path, patches, comments); err != nil {
				return fmt.Errorf("patching %s: %w", ef.Path, err)
			}
			for name, val := range patches {
				display := val
				if sensitiveVars[name] {
					display = "***"
				}
				fmt.Printf("bight: %s %s %s=%s\n", ef.Path, output.Dim("→"), output.Cyan(name), output.Bold(display))
			}
		}
	}
	return nil
}
