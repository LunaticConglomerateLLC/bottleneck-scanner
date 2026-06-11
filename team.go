package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type TeamConfig struct {
	Name            string   `yaml:"name"`
	Repos           []string `yaml:"repos"`
	Members         []string `yaml:"members"`
	ServiceAccounts []string `yaml:"service_accounts"`
}

func loadTeamConfig(path string) (TeamConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TeamConfig{}, fmt.Errorf("cannot open team file: %w", err)
	}

	var cfg TeamConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return TeamConfig{}, fmt.Errorf("cannot parse team file: %w", err)
	}

	if len(cfg.Repos) == 0 {
		return TeamConfig{}, fmt.Errorf("team file %q has no repos listed", path)
	}

	return cfg, nil
}
