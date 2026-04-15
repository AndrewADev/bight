package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AndrewADev/bight/internal/output"
	"gopkg.in/yaml.v3"
)

var userHomeDir = os.UserHomeDir

type Config struct {
	Project  string    `yaml:"project"`
	Defaults Defaults  `yaml:"defaults"`
	EnvFiles []EnvFile `yaml:"env_files"`
}

type Defaults struct {
	BranchTemplate  string `yaml:"branch_template"`
	CollectComments string `yaml:"collect-comments"`
}

type EnvFile struct {
	Path string `yaml:"path"`
	Vars []Var  `yaml:"vars"`
}

type Var struct {
	Name      string `yaml:"name"`
	Strategy  string `yaml:"strategy"`
	On        string `yaml:"on"`
	Sensitive bool   `yaml:"sensitive"`
}

func globalConfigPath() (string, bool) {
	home, err := userHomeDir()
	if err != nil {
		return "", false
	}
	return filepath.Join(home, ".bight.yml"), true
}

func merge(global, repo *Config) *Config {
	out := *global
	if repo.Project != "" {
		out.Project = repo.Project
	}
	if repo.Defaults.BranchTemplate != "" {
		out.Defaults.BranchTemplate = repo.Defaults.BranchTemplate
	}
	if repo.Defaults.CollectComments != "" {
		out.Defaults.CollectComments = repo.Defaults.CollectComments
	}
	if len(repo.EnvFiles) > 0 {
		out.EnvFiles = repo.EnvFiles
	}
	return &out
}

// LoadFrom loads config from a specific repo config file path, merging with
// the global config (~/.bight.yml) as usual. Unlike Load, it does not search
// for .bight.yml or .bight.yaml in the current directory.
func LoadFrom(repoConfigPath string) (*Config, error) {
	var global *Config
	if path, ok := globalConfigPath(); ok {
		if g, err := load(path); err == nil {
			if len(g.EnvFiles) > 0 {
				fmt.Fprintln(os.Stderr, output.WarnStderr("bight: warning: env_files in ~/.bight.yml is not supported and will be ignored; define env_files in the repo's .bight.yml instead"))
				g.EnvFiles = nil
			}
			global = g
		}
	}

	repo, err := load(repoConfigPath)
	if err != nil {
		return nil, err
	}

	if global != nil {
		return merge(global, repo), nil
	}
	return repo, nil
}

func Load() (*Config, error) {
	var global *Config
	if path, ok := globalConfigPath(); ok {
		if g, err := load(path); err == nil {
			if len(g.EnvFiles) > 0 {
				fmt.Fprintln(os.Stderr, output.WarnStderr("bight: warning: env_files in ~/.bight.yml is not supported and will be ignored; define env_files in the repo's .bight.yml instead"))
				g.EnvFiles = nil
			}
			global = g
		}
		// silently ignore missing or unreadable global config
	}

	var repo *Config
	for _, name := range []string{".bight.yml", ".bight.yaml"} {
		r, err := load(name)
		if err == nil {
			repo = r
			break
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	switch {
	case repo != nil && global != nil:
		return merge(global, repo), nil
	case repo != nil:
		return repo, nil
	case global != nil:
		return global, nil
	default:
		return nil, os.ErrNotExist
	}
}

func load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
