package cmd

import (
	"fmt"
	"os"

	"github.com/grantcarthew/acon/internal/api"
	"github.com/grantcarthew/acon/internal/config"
	"github.com/spf13/cobra"
)

var (
	spaceLimit int
)

var spaceCmd = &cobra.Command{
	Use:   "space",
	Short: "Manage Confluence spaces",
	Long:  "View and list Confluence spaces",
}

var spaceViewCmd = &cobra.Command{
	Use:   "view SPACE_KEY",
	Short: "View a space",
	Long:  "View details of a Confluence space",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
		spaceKey := args[0]

		space, err := client.GetSpace(spaceKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting space: %v\n", err)
			os.Exit(1)
		}

		if outputJSON {
			printJSON(space)
		} else {
			fmt.Printf("ID: %s\n", space.ID)
			fmt.Printf("Key: %s\n", space.Key)
			fmt.Printf("Name: %s\n", space.Name)
			fmt.Printf("Type: %s\n", space.Type)
		}
	},
}

var spaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List spaces",
	Long:  "List Confluence spaces",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)

		spaces, err := client.ListSpaces(spaceLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing spaces: %v\n", err)
			os.Exit(1)
		}

		if outputJSON {
			printJSON(spaces)
		} else {
			fmt.Println("Confluence Spaces:\n")
			for _, space := range spaces {
				fmt.Printf("Key: %s\n", space.Key)
				fmt.Printf("Name: %s\n", space.Name)
				fmt.Printf("Type: %s\n", space.Type)
				fmt.Printf("ID: %s\n", space.ID)
				fmt.Println("---")
			}
		}
	},
}

func init() {
	spaceViewCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")
	spaceListCmd.Flags().IntVarP(&spaceLimit, "limit", "l", 25, "Maximum number of spaces to list")
	spaceListCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	spaceCmd.AddCommand(spaceViewCmd)
	spaceCmd.AddCommand(spaceListCmd)
}
