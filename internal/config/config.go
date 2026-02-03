package config

import (
	"fmt"
	"io"
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
	return LoadWithVerbose(nil)
}

func LoadWithVerbose(verboseLog io.Writer) (Config, error) {
	logVerbose := func(format string, args ...interface{}) {
		if verboseLog != nil {
			fmt.Fprintf(verboseLog, format, args...)
		}
	}

	logVerbose("[Config] Loading configuration from environment\n")

	cfg := Config{
		SpaceKey: os.Getenv("CONFLUENCE_SPACE_KEY"),
	}

	if cfg.SpaceKey != "" {
		logVerbose("[Config] CONFLUENCE_SPACE_KEY: %s\n", cfg.SpaceKey)
	}

	// Base URL: CONFLUENCE_BASE_URL or ATLASSIAN_BASE_URL + /wiki
	cfg.BaseURL = os.Getenv("CONFLUENCE_BASE_URL")
	if cfg.BaseURL == "" {
		if atlasURL := os.Getenv("ATLASSIAN_BASE_URL"); atlasURL != "" {
			atlasURL = strings.TrimSuffix(atlasURL, "/")
			atlasURL = strings.TrimSuffix(atlasURL, "/wiki")
			cfg.BaseURL = atlasURL + "/wiki"
			logVerbose("[Config] Using ATLASSIAN_BASE_URL: %s (appended /wiki)\n", cfg.BaseURL)
		}
	} else {
		logVerbose("[Config] Using CONFLUENCE_BASE_URL: %s\n", cfg.BaseURL)
	}

	// Email: CONFLUENCE_EMAIL or ATLASSIAN_EMAIL
	cfg.Email = os.Getenv("CONFLUENCE_EMAIL")
	if cfg.Email == "" {
		cfg.Email = os.Getenv("ATLASSIAN_EMAIL")
		if cfg.Email != "" {
			logVerbose("[Config] Using ATLASSIAN_EMAIL: %s\n", cfg.Email)
		}
	} else {
		logVerbose("[Config] Using CONFLUENCE_EMAIL: %s\n", cfg.Email)
	}

	// API Token: CONFLUENCE_API_TOKEN, ATLASSIAN_API_TOKEN, or JIRA_API_TOKEN
	if val := os.Getenv("CONFLUENCE_API_TOKEN"); val != "" {
		cfg.APIToken = val
		logVerbose("[Config] Using CONFLUENCE_API_TOKEN: %s\n", maskToken(val))
	} else if val := os.Getenv("ATLASSIAN_API_TOKEN"); val != "" {
		cfg.APIToken = val
		logVerbose("[Config] Using ATLASSIAN_API_TOKEN: %s\n", maskToken(val))
	} else if val := os.Getenv("JIRA_API_TOKEN"); val != "" {
		cfg.APIToken = val
		logVerbose("[Config] Using JIRA_API_TOKEN: %s\n", maskToken(val))
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

	logVerbose("[Config] Configuration loaded successfully\n")
	return cfg, nil
}

// maskToken masks most of the token for security in logs
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "****" + token[len(token)-4:]
}
