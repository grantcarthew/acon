package cmd

import (
	"fmt"

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
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
		spaceKey := args[0]

		space, err := client.GetSpace(cmd.Context(), spaceKey)
		if err != nil {
			return fmt.Errorf("getting space: %w", err)
		}

		if outputJSON {
			return printJSON(space)
		}
		fmt.Printf("ID: %s\n", space.ID)
		fmt.Printf("Key: %s\n", space.Key)
		fmt.Printf("Name: %s\n", space.Name)
		fmt.Printf("Type: %s\n", space.Type)
		return nil
	},
}

var spaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List spaces",
	Long:  "List Confluence spaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)

		spaces, err := client.ListSpaces(cmd.Context(), spaceLimit)
		if err != nil {
			return fmt.Errorf("listing spaces: %w", err)
		}

		if outputJSON {
			return printJSON(spaces)
		}
		fmt.Println("Confluence Spaces:")
		for _, space := range spaces {
			fmt.Printf("Key: %s\n", space.Key)
			fmt.Printf("Name: %s\n", space.Name)
			fmt.Printf("Type: %s\n", space.Type)
			fmt.Printf("ID: %s\n", space.ID)
			fmt.Println("---")
		}
		return nil
	},
}

func init() {
	spaceViewCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")
	spaceListCmd.Flags().IntVarP(&spaceLimit, "limit", "l", 25, "Maximum number of spaces to list")
	spaceListCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	spaceCmd.AddCommand(spaceViewCmd)
	spaceCmd.AddCommand(spaceListCmd)
}
