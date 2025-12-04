package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
		wantCfg Config
	}{
		{
			name: "all required vars set with CONFLUENCE_API_TOKEN",
			env: map[string]string{
				"CONFLUENCE_BASE_URL":  "https://example.atlassian.net",
				"CONFLUENCE_EMAIL":     "user@example.com",
				"CONFLUENCE_API_TOKEN": "token123",
				"CONFLUENCE_SPACE_KEY": "SPACE",
			},
			wantCfg: Config{
				BaseURL:  "https://example.atlassian.net",
				Email:    "user@example.com",
				APIToken: "token123",
				SpaceKey: "SPACE",
			},
		},
		{
			name: "ATLASSIAN_API_TOKEN fallback",
			env: map[string]string{
				"CONFLUENCE_BASE_URL": "https://example.atlassian.net",
				"CONFLUENCE_EMAIL":    "user@example.com",
				"ATLASSIAN_API_TOKEN": "atlassian-token",
			},
			wantCfg: Config{
				BaseURL:  "https://example.atlassian.net",
				Email:    "user@example.com",
				APIToken: "atlassian-token",
			},
		},
		{
			name: "JIRA_API_TOKEN fallback",
			env: map[string]string{
				"CONFLUENCE_BASE_URL": "https://example.atlassian.net",
				"CONFLUENCE_EMAIL":    "user@example.com",
				"JIRA_API_TOKEN":      "jira-token",
			},
			wantCfg: Config{
				BaseURL:  "https://example.atlassian.net",
				Email:    "user@example.com",
				APIToken: "jira-token",
			},
		},
		{
			name: "CONFLUENCE_API_TOKEN takes priority",
			env: map[string]string{
				"CONFLUENCE_BASE_URL":  "https://example.atlassian.net",
				"CONFLUENCE_EMAIL":     "user@example.com",
				"CONFLUENCE_API_TOKEN": "confluence-token",
				"ATLASSIAN_API_TOKEN":  "atlassian-token",
				"JIRA_API_TOKEN":       "jira-token",
			},
			wantCfg: Config{
				BaseURL:  "https://example.atlassian.net",
				Email:    "user@example.com",
				APIToken: "confluence-token",
			},
		},
		{
			name: "missing base URL",
			env: map[string]string{
				"CONFLUENCE_EMAIL":     "user@example.com",
				"CONFLUENCE_API_TOKEN": "token123",
			},
			wantErr: "CONFLUENCE_BASE_URL",
		},
		{
			name: "missing email",
			env: map[string]string{
				"CONFLUENCE_BASE_URL":  "https://example.atlassian.net",
				"CONFLUENCE_API_TOKEN": "token123",
			},
			wantErr: "CONFLUENCE_EMAIL",
		},
		{
			name: "missing API token",
			env: map[string]string{
				"CONFLUENCE_BASE_URL": "https://example.atlassian.net",
				"CONFLUENCE_EMAIL":    "user@example.com",
			},
			wantErr: "API token not set",
		},
		{
			name: "space key is optional",
			env: map[string]string{
				"CONFLUENCE_BASE_URL":  "https://example.atlassian.net",
				"CONFLUENCE_EMAIL":     "user@example.com",
				"CONFLUENCE_API_TOKEN": "token123",
			},
			wantCfg: Config{
				BaseURL:  "https://example.atlassian.net",
				Email:    "user@example.com",
				APIToken: "token123",
				SpaceKey: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars
			clearEnvVars := []string{
				"CONFLUENCE_BASE_URL",
				"CONFLUENCE_EMAIL",
				"CONFLUENCE_API_TOKEN",
				"ATLASSIAN_API_TOKEN",
				"JIRA_API_TOKEN",
				"CONFLUENCE_SPACE_KEY",
			}
			for _, key := range clearEnvVars {
				t.Setenv(key, "")
			}

			// Set test env vars
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			cfg, err := Load()

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Load() expected error containing %q, got nil", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Load() error = %q, want error containing %q", err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error = %v", err)
				return
			}

			if cfg != tt.wantCfg {
				t.Errorf("Load() = %+v, want %+v", cfg, tt.wantCfg)
			}
		})
	}
}
