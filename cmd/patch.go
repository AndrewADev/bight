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

// Event names — see `on:` in .bight.yml.
const (
	EventCheckout     = "checkout"
	EventWorktreeInit = "worktree-init"
)

type dryRunResult struct {
	path      string
	varName   string
	value     string
	sensitive bool
	event     string
	err       error
}

// planFile decides what bight will do to env_file `ef` for this run, given
// whether the destination existed before. willCopy is whether the file
// should be (re)seeded from ef.Copy; willInit is whether worktree-init
// vars should fire (i.e. whether the file is being brought into existence
// by this run).
func planFile(ef config.EnvFile, existedBefore bool) (willCopy, willInit bool) {
	if ef.Copy != nil {
		if !existedBefore || ef.Copy.Overwrite {
			return true, true
		}
		// existed and overwrite disabled — skip copy, don't fire init
		return false, false
	}
	// no copy configured: init fires iff the file did not exist before
	return false, !existedBefore
}

func eventsFor(willInit bool) []string {
	if willInit {
		return []string{EventWorktreeInit, EventCheckout}
	}
	return []string{EventCheckout}
}

// shouldApply reports whether v fires for `event`. Each var fires exactly
// in its declared event — not across all events — so e.g. `on: worktree-init`
// vars are NOT also re-applied during the checkout pass.
func shouldApply(v config.Var, event string) bool {
	return v.On == event
}

// hasVarsForEvents reports whether any var in ef matches any of the given
// events. Used to decide whether a patch step will actually write anything.
func hasVarsForEvents(ef config.EnvFile, events []string) bool {
	set := make(map[string]bool, len(events))
	for _, e := range events {
		set[e] = true
	}
	for _, v := range ef.Vars {
		if set[v.On] {
			return true
		}
	}
	return false
}

// dryRunEnvFiles previews what patchEnvFiles would do without modifying
// any files. It stats each destination on the real filesystem to decide
// which events will fire (worktree-init only fires when the file is
// missing). This means dry-run output reflects the current filesystem
// state and may differ between runs as files come and go — by design.
func dryRunEnvFiles(cfg *config.Config, branch string) []dryRunResult {
	ctx := strategy.Context{Branch: branch, Project: cfg.Project}
	var results []dryRunResult

	for _, ef := range cfg.EnvFiles {
		_, statErr := os.Stat(ef.Path)
		existedBefore := statErr == nil
		_, willInit := planFile(ef, existedBefore)

		for _, event := range eventsFor(willInit) {
			for _, v := range ef.Vars {
				if !shouldApply(v, event) {
					continue
				}
				val, err := strategy.Apply(v.Strategy, ctx, cfg)
				results = append(results, dryRunResult{
					path:      ef.Path,
					varName:   v.Name,
					value:     val,
					sensitive: v.Sensitive,
					event:     event,
					err:       err,
				})
			}
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

		willCopy, willInit := planFile(ef, existedBefore)
		events := eventsFor(willInit)

		// If copy is configured but skipped because the file exists, surface
		// that as a warning so the user understands why nothing was copied.
		if ef.Copy != nil && !willCopy {
			fmt.Fprintln(os.Stderr, output.WarnStderr(
				fmt.Sprintf("bight: warning: %s exists; skipping copy from %s (set copy.overwrite: true to clobber)",
					ef.Path, ef.Copy.Source)))
		}

		// Single backup point per file, before any mutation. Only back up
		// when something will actually change the file — either a copy, or
		// a var fires for an event that will run this pass.
		if ef.Backup && existedBefore && (willCopy || hasVarsForEvents(ef, events)) {
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

		for _, event := range events {
			patches := make(map[string]string)
			sensitiveVars := make(map[string]bool)
			for _, v := range ef.Vars {
				if !shouldApply(v, event) {
					continue
				}
				val, err := strategy.Apply(v.Strategy, ctx, cfg)
				if err != nil {
					return fmt.Errorf("var %s: %w", v.Name, err)
				}
				patches[v.Name] = val
				sensitiveVars[v.Name] = v.Sensitive
			}
			if len(patches) == 0 {
				continue
			}

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
				fmt.Printf("bight: %s %s %s=%s %s\n",
					ef.Path, output.Dim("→"),
					output.Cyan(name), output.Bold(display),
					output.Dim("("+event+")"))
			}
		}
	}
	return nil
}
