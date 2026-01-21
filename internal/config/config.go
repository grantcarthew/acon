package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BaseURL  string
	Email    string
	APIToken string
	SpaceKey string
}

func Load() (Config, error) {
	cfg := Config{
		SpaceKey: os.Getenv("CONFLUENCE_SPACE_KEY"),
	}

	// Base URL: CONFLUENCE_BASE_URL or ATLASSIAN_BASE_URL + /wiki
	cfg.BaseURL = os.Getenv("CONFLUENCE_BASE_URL")
	if cfg.BaseURL == "" {
		if atlasURL := os.Getenv("ATLASSIAN_BASE_URL"); atlasURL != "" {
			atlasURL = strings.TrimSuffix(atlasURL, "/")
			atlasURL = strings.TrimSuffix(atlasURL, "/wiki")
			cfg.BaseURL = atlasURL + "/wiki"
		}
	}

	// Email: CONFLUENCE_EMAIL or ATLASSIAN_EMAIL
	cfg.Email = os.Getenv("CONFLUENCE_EMAIL")
	if cfg.Email == "" {
		cfg.Email = os.Getenv("ATLASSIAN_EMAIL")
	}

	// API Token: CONFLUENCE_API_TOKEN, ATLASSIAN_API_TOKEN, or JIRA_API_TOKEN
	if val := os.Getenv("CONFLUENCE_API_TOKEN"); val != "" {
		cfg.APIToken = val
	} else if val := os.Getenv("ATLASSIAN_API_TOKEN"); val != "" {
		cfg.APIToken = val
	} else if val := os.Getenv("JIRA_API_TOKEN"); val != "" {
		cfg.APIToken = val
	}

	if cfg.BaseURL == "" {
		return Config{}, fmt.Errorf("CONFLUENCE_BASE_URL (or ATLASSIAN_BASE_URL) not set")
	}
	if cfg.Email == "" {
		return Config{}, fmt.Errorf("CONFLUENCE_EMAIL (or ATLASSIAN_EMAIL) not set")
	}
	if cfg.APIToken == "" {
		return Config{}, fmt.Errorf("API token not set (set CONFLUENCE_API_TOKEN, ATLASSIAN_API_TOKEN, or JIRA_API_TOKEN)")
	}

	return cfg, nil
}
