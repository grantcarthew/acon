package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/grantcarthew/acon/internal/api"
	"github.com/grantcarthew/acon/internal/config"
	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via -ldflags.
	Version = "dev"

	verbose bool
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Show detailed warnings and debug information")

	rootCmd.Version = Version
	rootCmd.SetVersionTemplate(`acon version {{.Version}}
Repository: https://github.com/grantcarthew/acon
Report issues: https://github.com/grantcarthew/acon/issues/new
`)

	// Command groups for organized help output
	rootCmd.AddGroup(&cobra.Group{ID: "core", Title: "Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "utility", Title: "Utilities:"})

	// Disable default completion command (we provide our own with GroupID)
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	pageCmd.GroupID = "core"
	spaceCmd.GroupID = "core"

	rootCmd.AddCommand(pageCmd)
	rootCmd.AddCommand(spaceCmd)
}

func Execute() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return rootCmd.ExecuteContext(ctx)
}

// initClient loads configuration and creates an API client.
// Returns the client and config for commands that need access to config values like SpaceKey.
func initClient() (*api.Client, *config.Config, error) {
	var verboseLog io.Writer
	if verbose {
		verboseLog = os.Stderr
	}

	cfg, err := config.LoadWithVerbose(verboseLog)
	if err != nil {
		return nil, nil, err
	}
	client, err := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// Enable verbose logging if flag is set
	if verbose {
		client.VerboseLog = os.Stderr
	}

	return client, &cfg, nil
}
