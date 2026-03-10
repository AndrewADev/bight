package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

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
	Name     string `yaml:"name"`
	Strategy string `yaml:"strategy"`
	On       string `yaml:"on"`
}

func Load() (*Config, error) {
	for _, name := range []string{".bight.yml", ".bight.yaml"} {
		cfg, err := load(name)
		if err == nil {
			return cfg, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, os.ErrNotExist
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
