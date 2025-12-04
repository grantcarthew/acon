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
	Short: "Debug converter functions",
}

var debugMdCmd = &cobra.Command{
	Use:   "md",
	Short: "Convert markdown to storage format",
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

var debugStorageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Convert storage format to markdown",
	RunE: func(cmd *cobra.Command, args []string) error {
		storage, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		markdown, err := converter.StorageToMarkdown(string(storage))
		if err != nil {
			return fmt.Errorf("converting storage to markdown: %w", err)
		}
		fmt.Println(markdown)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.AddCommand(debugMdCmd)
	debugCmd.AddCommand(debugStorageCmd)
}
