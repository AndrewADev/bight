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
	BranchTemplate string `yaml:"branch_template"`
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

func Load(path string) (*Config, error) {
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
