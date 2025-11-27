package config

import (
	"fmt"
	"os"
)

type Config struct {
	BaseURL  string
	Email    string
	APIToken string
	SpaceKey string
}

func Load() (Config, error) {
	cfg := Config{
		BaseURL:  os.Getenv("CONFLUENCE_BASE_URL"),
		Email:    os.Getenv("CONFLUENCE_EMAIL"),
		SpaceKey: os.Getenv("CONFLUENCE_SPACE_KEY"),
	}

	if val := os.Getenv("CONFLUENCE_API_TOKEN"); val != "" {
		cfg.APIToken = val
	} else if val := os.Getenv("ATLASSIAN_API_TOKEN"); val != "" {
		cfg.APIToken = val
	} else if val := os.Getenv("JIRA_API_TOKEN"); val != "" {
		cfg.APIToken = val
	}

	if cfg.BaseURL == "" {
		return Config{}, fmt.Errorf("CONFLUENCE_BASE_URL environment variable not set")
	}
	if cfg.Email == "" {
		return Config{}, fmt.Errorf("CONFLUENCE_EMAIL environment variable not set")
	}
	if cfg.APIToken == "" {
		return Config{}, fmt.Errorf("API token not set (set CONFLUENCE_API_TOKEN, ATLASSIAN_API_TOKEN, or JIRA_API_TOKEN)")
	}

	return cfg, nil
}
