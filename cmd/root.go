package cmd

import (
	"fmt"
	"os"

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

Required Environment Variables:
  CONFLUENCE_BASE_URL       Confluence instance URL
  CONFLUENCE_EMAIL          Your email address
  CONFLUENCE_API_TOKEN      API token (or ATLASSIAN_API_TOKEN/JIRA_API_TOKEN)

Optional Environment Variables:
  CONFLUENCE_SPACE_KEY      Default space key (can be overridden with -s flag)`,
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Println(appVersion)
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
