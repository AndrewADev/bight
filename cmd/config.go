package cmd

import "github.com/AndrewADev/bight/internal/config"

var configPath string

func loadConfig() (*config.Config, error) {
	if configPath != "" {
		return config.LoadFrom(configPath)
	}
	return config.Load()
}
