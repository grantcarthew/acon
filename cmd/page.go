package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

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
	pageSort   string
	pageDesc   bool
	outputJSON bool
	updateMsg  string
	moveParent string
)

// mapChildSortValue converts friendly sort names to API values for child pages
// Returns empty string for "title" as it's handled client-side
func mapChildSortValue(sort string, desc bool) (apiSort string, valid bool) {
	// Default to web (child-position) if no sort specified
	if sort == "" {
		sort = "web"
	}

	// Title is valid but sorted client-side, not by API
	if sort == "title" {
		return "", true
	}

	apiValue := map[string]string{
		"web":      "child-position",
		"created":  "created-date",
		"modified": "modified-date",
		"id":       "id",
	}[sort]

	if apiValue == "" {
		return "", false
	}

	if desc {
		return "-" + apiValue, true
	}
	return apiValue, true
}

// mapSpaceSortValue converts friendly sort names to API values for space page listing
func mapSpaceSortValue(sort string, desc bool) string {
	// No default - API handles it
	if sort == "" {
		if desc {
			return "-id" // Default to id desc if only --desc provided
		}
		return ""
	}

	apiValue := map[string]string{
		"title":    "title",
		"created":  "created-date",
		"modified": "modified-date",
		"id":       "id",
	}[sort]

	if apiValue == "" {
		return ""
	}

	if desc {
		return "-" + apiValue
	}
	return apiValue
}

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

		var pages []api.Page

		if pageParent != "" {
			// List children of a specific parent page
			sortValue, valid := mapChildSortValue(pageSort, pageDesc)
			if !valid {
				fmt.Fprintf(os.Stderr, "Error: invalid sort value '%s' (valid: web, title, created, modified, id)\n", pageSort)
				os.Exit(1)
			}
			var err error
			pages, err = client.GetChildPages(pageParent, pageLimit, sortValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing child pages: %v\n", err)
				os.Exit(1)
			}

			// Client-side title sort (not supported by API)
			if pageSort == "title" {
				sort.Slice(pages, func(i, j int) bool {
					if pageDesc {
						return strings.ToLower(pages[i].Title) > strings.ToLower(pages[j].Title)
					}
					return strings.ToLower(pages[i].Title) < strings.ToLower(pages[j].Title)
				})
			}
		} else {
			// List pages in space
			spaceKey := pageSpace
			if spaceKey == "" {
				spaceKey = cfg.SpaceKey
			}

			sortValue := mapSpaceSortValue(pageSort, pageDesc)
			if sortValue == "" && pageSort != "" {
				fmt.Fprintf(os.Stderr, "Error: invalid sort value '%s' (valid: title, created, modified, id)\n", pageSort)
				os.Exit(1)
			}

			space, err := client.GetSpace(spaceKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting space: %v\n", err)
				os.Exit(1)
			}

			pages, err = client.ListPages(space.ID, pageLimit, sortValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing pages: %v\n", err)
				os.Exit(1)
			}
		}

		if outputJSON {
			printJSON(pages)
		} else {
			if pageParent != "" {
				fmt.Printf("Child pages of %s:\n\n", pageParent)
			} else {
				spaceKey := pageSpace
				if spaceKey == "" {
					spaceKey = cfg.SpaceKey
				}
				fmt.Printf("Pages in space %s:\n\n", spaceKey)
			}
			for _, page := range pages {
				fmt.Printf("ID: %s\n", page.ID)
				fmt.Printf("Title: %s\n", page.Title)
				fmt.Printf("Status: %s\n", page.Status)
				fmt.Println("---")
			}
		}
	},
}

var pageMoveCmd = &cobra.Command{
	Use:   "move PAGE_ID",
	Short: "Move a page to a new parent",
	Long:  "Move a Confluence page to a new parent page within the same space",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		client := api.NewClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
		pageID := args[0]

		if moveParent == "" {
			fmt.Fprintf(os.Stderr, "Error: --parent flag is required\n")
			os.Exit(1)
		}

		result, err := client.MovePage(pageID, moveParent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error moving page: %v\n", err)
			os.Exit(1)
		}

		if outputJSON {
			printJSON(result)
		} else {
			fmt.Printf("Page moved successfully\n")
			fmt.Printf("ID: %s\n", result.ID)
			fmt.Printf("Title: %s\n", result.Title)
			fmt.Printf("New Parent ID: %s\n", moveParent)
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
	pageListCmd.Flags().StringVarP(&pageParent, "parent", "p", "", "Parent page ID (list children of this page)")
	pageListCmd.Flags().IntVarP(&pageLimit, "limit", "l", 25, "Maximum number of pages to list")
	pageListCmd.Flags().StringVar(&pageSort, "sort", "", "Sort order: web, title, created, modified, id")
	pageListCmd.Flags().BoolVar(&pageDesc, "desc", false, "Sort in descending order")
	pageListCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	pageMoveCmd.Flags().StringVarP(&moveParent, "parent", "p", "", "Target parent page ID (required)")
	pageMoveCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")
	pageMoveCmd.MarkFlagRequired("parent")

	pageCmd.AddCommand(pageCreateCmd)
	pageCmd.AddCommand(pageViewCmd)
	pageCmd.AddCommand(pageUpdateCmd)
	pageCmd.AddCommand(pageDeleteCmd)
	pageCmd.AddCommand(pageListCmd)
	pageCmd.AddCommand(pageMoveCmd)
}
