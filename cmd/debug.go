package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/grantcarthew/acon/internal/converter"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug markdown to storage conversion",
	RunE: func(cmd *cobra.Command, args []string) error {
		markdown, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		storage := converter.MarkdownToStorage(string(markdown))
		fmt.Println(storage)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
