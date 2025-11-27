package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/grantcarthew/acon/internal/api"
	"github.com/grantcarthew/acon/internal/config"
	"github.com/grantcarthew/acon/internal/converter"
	"github.com/spf13/cobra"
)

const (
	maxContentSize = 10 * 1024 * 1024 // 10MB
)

var (
	pageTitle  string
	pageFile   string
	pageSpace  string
	pageParent string
	pageLimit  int
	outputJSON bool
	updateMsg  string
)

var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Manage Confluence pages",
	Long:  "Create, view, update, and delete Confluence pages",
}

var pageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new page",
	Long:  "Create a new Confluence page from markdown file or stdin",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)

		spaceKey := pageSpace
		if spaceKey == "" {
			spaceKey = cfg.SpaceKey
		}

		space, err := client.GetSpace(spaceKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting space: %v\n", err)
			os.Exit(1)
		}

		content, err := readAndValidateContent(pageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		htmlContent := converter.MarkdownToStorage(string(content))

		req := &api.PageCreateRequest{
			SpaceID: space.ID,
			Status:  "current",
			Title:   pageTitle,
			Body: &api.PageBodyWrite{
				Representation: "storage",
				Value:          htmlContent,
			},
		}

		if pageParent != "" {
			req.ParentID = pageParent
		}

		result, err := client.CreatePage(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating page: %v\n", err)
			os.Exit(1)
		}

		if outputJSON {
			printJSON(result)
		} else {
			fmt.Printf("Page created successfully\n")
			fmt.Printf("ID: %s\n", result.ID)
			fmt.Printf("Title: %s\n", result.Title)
			fmt.Printf("URL: %s/wiki/spaces/%s/pages/%s\n", cfg.BaseURL, spaceKey, result.ID)
		}
	},
}

var pageViewCmd = &cobra.Command{
	Use:   "view PAGE_ID",
	Short: "View a page",
	Long:  "View details of a Confluence page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
		pageID := args[0]

		page, err := client.GetPage(pageID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting page: %v\n", err)
			os.Exit(1)
		}

		if outputJSON {
			printJSON(page)
		} else {
			fmt.Printf("ID: %s\n", page.ID)
			fmt.Printf("Title: %s\n", page.Title)
			fmt.Printf("Status: %s\n", page.Status)
			if page.Version != nil {
				fmt.Printf("Version: %d\n", page.Version.Number)
			}
			if page.Body != nil && page.Body.Storage != nil {
				markdown, err := converter.StorageToMarkdown(page.Body.Storage.Value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to convert to markdown: %v\n", err)
					fmt.Printf("\nContent:\n%s\n", page.Body.Storage.Value)
				} else {
					fmt.Printf("\nContent:\n%s\n", markdown)
				}
			}
		}
	},
}

var pageUpdateCmd = &cobra.Command{
	Use:   "update PAGE_ID",
	Short: "Update a page",
	Long:  "Update an existing Confluence page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
		pageID := args[0]

		existing, err := client.GetPage(pageID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting existing page: %v\n", err)
			os.Exit(1)
		}

		content, err := readAndValidateContent(pageFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		htmlContent := converter.MarkdownToStorage(string(content))

		title := pageTitle
		if title == "" {
			title = existing.Title
		}

		newVersion := 1
		if existing.Version != nil {
			newVersion = existing.Version.Number + 1
		}

		req := &api.PageUpdateRequest{
			ID:      pageID,
			SpaceID: existing.SpaceID,
			Status:  "current",
			Title:   title,
			Body: &api.PageBodyWrite{
				Representation: "storage",
				Value:          htmlContent,
			},
			Version: &api.Version{
				Number:  newVersion,
				Message: updateMsg,
			},
		}

		result, err := client.UpdatePage(pageID, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating page: %v\n", err)
			os.Exit(1)
		}

		if outputJSON {
			printJSON(result)
		} else {
			fmt.Printf("Page updated successfully\n")
			fmt.Printf("ID: %s\n", result.ID)
			fmt.Printf("Title: %s\n", result.Title)
			if result.Version != nil {
				fmt.Printf("Version: %d\n", result.Version.Number)
			}
		}
	},
}

var pageDeleteCmd = &cobra.Command{
	Use:   "delete PAGE_ID",
	Short: "Delete a page",
	Long:  "Delete a Confluence page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
		pageID := args[0]

		if err := client.DeletePage(pageID); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting page: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Page %s deleted successfully\n", pageID)
	},
}

var pageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pages",
	Long:  "List pages in a Confluence space",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)

		spaceKey := pageSpace
		if spaceKey == "" {
			spaceKey = cfg.SpaceKey
		}

		space, err := client.GetSpace(spaceKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting space: %v\n", err)
			os.Exit(1)
		}

		pages, err := client.ListPages(space.ID, pageLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing pages: %v\n", err)
			os.Exit(1)
		}

		if outputJSON {
			printJSON(pages)
		} else {
			fmt.Printf("Pages in space %s:\n\n", spaceKey)
			for _, page := range pages {
				fmt.Printf("ID: %s\n", page.ID)
				fmt.Printf("Title: %s\n", page.Title)
				fmt.Printf("Status: %s\n", page.Status)
				fmt.Println("---")
			}
		}
	},
}

func readAndValidateContent(pageFile string) ([]byte, error) {
	var content []byte
	var err error

	if pageFile != "" {
		// Check file size before reading
		info, err := os.Stat(pageFile)
		if err != nil {
			return nil, fmt.Errorf("stat file: %w", err)
		}
		if info.Size() > maxContentSize {
			return nil, fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxContentSize)
		}

		content, err = os.ReadFile(pageFile)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
	} else {
		// Limit stdin reading
		limitedReader := io.LimitReader(os.Stdin, maxContentSize+1)
		content, err = io.ReadAll(limitedReader)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		if len(content) > maxContentSize {
			return nil, fmt.Errorf("stdin too large (max %d bytes)", maxContentSize)
		}
	}

	content = bytes.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("content cannot be empty")
	}

	return content, nil
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func init() {
	pageCreateCmd.Flags().StringVarP(&pageTitle, "title", "t", "", "Page title (required)")
	pageCreateCmd.Flags().StringVarP(&pageFile, "file", "f", "", "Markdown file (use stdin if not specified)")
	pageCreateCmd.Flags().StringVarP(&pageSpace, "space", "s", "", "Space key (uses config default if not specified)")
	pageCreateCmd.Flags().StringVarP(&pageParent, "parent", "p", "", "Parent page ID")
	pageCreateCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")
	pageCreateCmd.MarkFlagRequired("title")

	pageViewCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	pageUpdateCmd.Flags().StringVarP(&pageTitle, "title", "t", "", "New page title (optional)")
	pageUpdateCmd.Flags().StringVarP(&pageFile, "file", "f", "", "Markdown file (use stdin if not specified)")
	pageUpdateCmd.Flags().StringVarP(&updateMsg, "message", "m", "", "Version update message")
	pageUpdateCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	pageListCmd.Flags().StringVarP(&pageSpace, "space", "s", "", "Space key (uses config default if not specified)")
	pageListCmd.Flags().IntVarP(&pageLimit, "limit", "l", 25, "Maximum number of pages to list")
	pageListCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	pageCmd.AddCommand(pageCreateCmd)
	pageCmd.AddCommand(pageViewCmd)
	pageCmd.AddCommand(pageUpdateCmd)
	pageCmd.AddCommand(pageDeleteCmd)
	pageCmd.AddCommand(pageListCmd)
}
