package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/grantcarthew/acon/internal/converter"
	"github.com/spf13/cobra"
)

var useGoldmark bool

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug markdown to storage conversion",
	Run: func(cmd *cobra.Command, args []string) {
		markdown, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}

		var storage string
		if useGoldmark {
			storage = converter.MarkdownToStorageGoldmark(string(markdown))
		} else {
			storage = converter.MarkdownToStorage(string(markdown))
		}
		fmt.Println(storage)
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
	debugCmd.Flags().BoolVarP(&useGoldmark, "goldmark", "g", true, "Use goldmark parser (default true)")
}
