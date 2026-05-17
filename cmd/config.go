package cmd

import (
	"os"

	"github.com/AndrewADev/bight/internal/config"
)

var configPath string

// envConfigPath is the environment variable that, when set, points bight at a
// repo config file in lieu of auto-discovering .bight.yml in the current
// directory. The --config flag, if set, takes precedence over the env var.
const envConfigPath = "BIGHT_CONFIG"

type configSource int

const (
	sourceAuto configSource = iota
	sourceFlag
	sourceEnv
)

// configSourceSuffix returns a parenthesized hint describing how the config
// was selected, suitable for appending to a success message. Returns "" for
// the default auto-discovered case.
func configSourceSuffix(s configSource) string {
	switch s {
	case sourceFlag:
		return " (from --config)"
	case sourceEnv:
		return " (from " + envConfigPath + ")"
	default:
		return ""
	}
}

// configOriginSuffix is the failure-message counterpart: when an explicit
// path was requested (flag or env) it includes that path so the user knows
// which file failed to load. Falls back to ".bight.yml" for the default case.
func configOriginSuffix(path string, s configSource) string {
	switch s {
	case sourceFlag:
		return " " + path + " (from --config)"
	case sourceEnv:
		return " " + path + " (from " + envConfigPath + ")"
	default:
		return " .bight.yml"
	}
}

func loadConfig() (*config.Config, string, configSource, error) {
	if configPath != "" {
		cfg, _, err := config.LoadFrom(configPath)
		// Always echo back the requested path so diagnostics can show *what*
		// was asked for, even when the file failed to load.
		return cfg, configPath, sourceFlag, err
	}
	if p := os.Getenv(envConfigPath); p != "" {
		cfg, _, err := config.LoadFrom(p)
		return cfg, p, sourceEnv, err
	}
	cfg, path, err := config.Load()
	return cfg, path, sourceAuto, err
}
