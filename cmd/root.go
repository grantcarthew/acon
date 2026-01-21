package cmd

import (
	"fmt"
	"os"

	"github.com/grantcarthew/acon/internal/api"
	"github.com/grantcarthew/acon/internal/config"
	"github.com/spf13/cobra"
)

var (
	showVersion bool
	appVersion  string
)

var rootCmd = &cobra.Command{
	Use:   "acon",
	Short: "Atlassian Confluence CLI",
	Long: `A command line interface for Atlassian Confluence

Environment Variables:
  ATLASSIAN_BASE_URL        Atlassian instance URL (shared with ajira, /wiki appended)
  ATLASSIAN_EMAIL           User email (shared with ajira)
  ATLASSIAN_API_TOKEN       API token (shared with ajira)
  CONFLUENCE_BASE_URL       Confluence URL (overrides ATLASSIAN_BASE_URL)
  CONFLUENCE_EMAIL          User email (overrides ATLASSIAN_EMAIL)
  CONFLUENCE_API_TOKEN      API token (overrides ATLASSIAN_API_TOKEN)
  CONFLUENCE_SPACE_KEY      Default space key (optional)`,
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("acon version %s\n", appVersion)
			fmt.Println("Repository: https://github.com/grantcarthew/acon")
			fmt.Println("Report issues: https://github.com/grantcarthew/acon/issues/new")
			return
		}
		cmd.Help()
	},
}

func Execute(version string) {
	appVersion = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version")
	rootCmd.AddCommand(pageCmd)
	rootCmd.AddCommand(spaceCmd)
}

// initClient loads configuration and creates an API client.
// Returns the client and config for commands that need access to config values like SpaceKey.
func initClient() (*api.Client, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
	return client, &cfg, nil
}
