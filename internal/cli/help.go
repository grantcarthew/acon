package cli

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

//go:embed agent-help/overview.md
var agentHelpOverview string

//go:embed agent-help/workflow.md
var agentHelpWorkflow string

var helpAgentsCmd = &cobra.Command{
	Use:   "agents [topic]",
	Short: "Token-efficient help for AI agents",
	Long:  "Displays token-efficient help documentation designed for AI agent consumption.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runHelpAgents,
}

func runHelpAgents(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		fmt.Println(agentHelpOverview)
		return nil
	}

	topic := args[0]

	if topic == "all" {
		fmt.Println(agentHelpOverview)
		fmt.Println("\n---")
		fmt.Println(agentHelpWorkflow)
		return nil
	}

	topics := map[string]string{
		"workflow": agentHelpWorkflow,
	}

	content, exists := topics[topic]
	if !exists {
		return fmt.Errorf("unknown help topic: %s\n\nAvailable topics:\n  workflow, all", topic)
	}

	fmt.Println(content)
	return nil
}

func init() {
	helpCmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long:  "Help provides help for any command in the application.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return rootCmd.Help()
			}
			// Find the command and show its help
			targetCmd, _, err := rootCmd.Find(args)
			if err != nil {
				return err
			}
			return targetCmd.Help()
		},
	}
	helpCmd.AddCommand(helpAgentsCmd)
	helpCmd.GroupID = "utility"
	rootCmd.SetHelpCommand(helpCmd)
}
