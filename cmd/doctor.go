package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/AndrewADev/bight/internal/config"
	"github.com/AndrewADev/bight/internal/hook"
	"github.com/AndrewADev/bight/internal/output"
	"github.com/spf13/cobra"
)

type result struct {
	status string // "ok", "info", "warn", "fail"
	msg    string
}

func ok(msg string) result   { return result{"ok", msg} }
func info(msg string) result { return result{"info", msg} }
func warn(msg string) result { return result{"warn", msg} }
func fail(msg string) result { return result{"fail", msg} }

// Filesystem findings encapsulated in struct to allow testing
// indepedent of the filesystem.
type checkDeps struct {
	gitOK            bool
	hookErr          error
	existingEnvFiles map[string]bool
}

func runChecks(cfg *config.Config, cfgErr error, deps checkDeps) []result {
	var results []result

	// Check 1: git repo
	if deps.gitOK {
		results = append(results, ok("git repo detected"))
	} else {
		results = append(results, fail("git repo: .git/hooks/ not found — are you in a git repo?"))
	}

	// Check 2: config loadable
	if cfgErr != nil {
		results = append(results, fail(fmt.Sprintf("config: failed to load .bight.yml — %s", cfgErr)))
	} else {
		results = append(results, ok("config: .bight.yml loaded"))
	}

	// Check 3: config valid (only if loadable)
	if cfgErr == nil {
		if cfg.Project == "" {
			results = append(results, fail("config: project field is empty"))
		} else if len(cfg.EnvFiles) == 0 {
			results = append(results, fail(fmt.Sprintf("config: project = %q, but no env_files defined", cfg.Project)))
		} else {
			results = append(results, ok(fmt.Sprintf("config: project = %q, %d env file(s)", cfg.Project, len(cfg.EnvFiles))))
		}
	}

	// Check 4: hook installed (info only)
	if deps.hookErr != nil {
		results = append(results, info("hook: not installed — run `bight install` to automate on checkout"))
	} else {
		results = append(results, ok("hook: installed"))
	}

	// Checks 5–7 require a loaded config
	if cfgErr != nil {
		results = append(results, warn("skipping env file and var checks — config could not be loaded"))
		return results
	}

	// Check 5: env files exist (warn only) + inert files with no checkout vars (info)
	for _, ef := range cfg.EnvFiles {
		if deps.existingEnvFiles[ef.Path] {
			results = append(results, ok(fmt.Sprintf("env file: %s", ef.Path)))
		} else {
			results = append(results, warn(fmt.Sprintf("env file: %s — not found (will be created on first patch)", ef.Path)))
		}

		applicable := 0
		for _, v := range ef.Vars {
			if v.On == "checkout" {
				applicable++
			}
		}
		if applicable == 0 {
			results = append(results, info(fmt.Sprintf("env file: %s — no vars apply on checkout; file will be left untouched", ef.Path)))
		}
	}

	// Check 6: strategies valid
	validStrategies := map[string]bool{"template": true, "random": true, "deterministic": true}
	var badStrategies []string
	for _, ef := range cfg.EnvFiles {
		for _, v := range ef.Vars {
			if !validStrategies[v.Strategy] {
				badStrategies = append(badStrategies, fmt.Sprintf("%s (strategy: %q)", v.Name, v.Strategy))
			}
		}
	}
	if len(badStrategies) > 0 {
		results = append(results, fail(fmt.Sprintf("vars: unknown strategy in: %v", badStrategies)))
	} else {
		results = append(results, ok("vars: all strategies valid"))
	}

	// Check 7: triggers valid
	validTriggers := map[string]bool{"checkout": true}
	var badTriggers []string
	for _, ef := range cfg.EnvFiles {
		for _, v := range ef.Vars {
			if !validTriggers[v.On] {
				badTriggers = append(badTriggers, fmt.Sprintf("%s (on: %q)", v.Name, v.On))
			}
		}
	}
	if len(badTriggers) > 0 {
		results = append(results, fail(fmt.Sprintf("vars: unknown trigger in: %v", badTriggers)))
	} else {
		results = append(results, ok("vars: all triggers valid"))
	}

	return results
}

func coloredStatus(r result) string {
	tag := fmt.Sprintf("%-6s", "["+r.status+"]")
	switch r.status {
	case "ok":
		return output.Green(tag)
	case "info":
		return output.Cyan(tag)
	case "warn":
		return output.Yellow(tag)
	case "fail":
		return output.Red(tag)
	default:
		return tag
	}
}

func doctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "doctor",
		Short:        "Validate bight setup and config",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, gitErr := hook.HooksDir()
			cfg, cfgErr := loadConfig()

			existing := map[string]bool{}
			if cfg != nil {
				for _, ef := range cfg.EnvFiles {
					_, err := os.Stat(ef.Path)
					existing[ef.Path] = err == nil
				}
			}
			results := runChecks(cfg, cfgErr, checkDeps{
				gitOK:            gitErr == nil,
				hookErr:          hook.Check(),
				existingEnvFiles: existing,
			})

			fmt.Println("bight doctor:")
			anyFailed := false
			for _, r := range results {
				fmt.Printf("  %s %s\n", coloredStatus(r), r.msg)
				if r.status == "fail" {
					anyFailed = true
				}
			}

			if anyFailed {
				return errors.New("one or more checks failed")
			}
			return nil
		},
	}
}
