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

	// stdinReader is the source for stdin input. Override in tests.
	stdinReader io.Reader = os.Stdin
	// stdinStat returns stdin file info. Override in tests.
	stdinStat func() (os.FileInfo, error) = func() (os.FileInfo, error) { return os.Stdin.Stat() }
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

// PageURL returns the browse URL for a Confluence page.
// If spaceKey is provided, uses the canonical URL format.
// Otherwise uses a generic format that Confluence redirects.
func PageURL(baseURL, spaceKey, pageID string) string {
	if spaceKey != "" {
		return fmt.Sprintf("%s/wiki/spaces/%s/pages/%s", baseURL, spaceKey, pageID)
	}
	return fmt.Sprintf("%s/wiki/pages/%s", baseURL, pageID)
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
	RunE: func(cmd *cobra.Command, args []string) error {
		client, cfg, err := initClient()
		if err != nil {
			return err
		}

		spaceKey := pageSpace
		if spaceKey == "" {
			spaceKey = cfg.SpaceKey
		}
		if spaceKey == "" {
			return fmt.Errorf("space key required: use --space flag or set CONFLUENCE_SPACE_KEY")
		}

		space, err := client.GetSpace(cmd.Context(), spaceKey)
		if err != nil {
			return fmt.Errorf("getting space: %w", err)
		}

		content, err := readAndValidateContent(pageFile)
		if err != nil {
			return err
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

		result, err := client.CreatePage(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("creating page: %w", err)
		}

		if outputJSON {
			return printJSON(result)
		}
		fmt.Println(PageURL(cfg.BaseURL, spaceKey, result.ID))
		return nil
	},
}

var pageViewCmd = &cobra.Command{
	Use:   "view PAGE_ID",
	Short: "View a page",
	Long:  "View details of a Confluence page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := initClient()
		if err != nil {
			return err
		}

		pageID := args[0]

		page, err := client.GetPage(cmd.Context(), pageID)
		if err != nil {
			return fmt.Errorf("getting page: %w", err)
		}

		if outputJSON {
			return printJSON(page)
		}
		if page.Body != nil && page.Body.Storage != nil {
			markdown, err := converter.StorageToMarkdown(page.Body.Storage.Value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to convert to markdown: %v\n", err)
				fmt.Println(page.Body.Storage.Value)
			} else {
				fmt.Println(markdown)
			}
		}
		return nil
	},
}

var pageUpdateCmd = &cobra.Command{
	Use:   "update PAGE_ID",
	Short: "Update a page",
	Long:  "Update an existing Confluence page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, cfg, err := initClient()
		if err != nil {
			return err
		}

		pageID := args[0]

		existing, err := client.GetPage(cmd.Context(), pageID)
		if err != nil {
			return fmt.Errorf("getting existing page: %w", err)
		}

		content, err := readAndValidateContent(pageFile)
		if err != nil {
			return err
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

		result, err := client.UpdatePage(cmd.Context(), pageID, req)
		if err != nil {
			return fmt.Errorf("updating page: %w", err)
		}

		if outputJSON {
			return printJSON(result)
		}
		fmt.Println(PageURL(cfg.BaseURL, "", result.ID))
		return nil
	},
}

var pageDeleteCmd = &cobra.Command{
	Use:   "delete PAGE_ID",
	Short: "Delete a page",
	Long:  "Delete a Confluence page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := initClient()
		if err != nil {
			return err
		}

		pageID := args[0]

		if err := client.DeletePage(cmd.Context(), pageID); err != nil {
			return fmt.Errorf("deleting page: %w", err)
		}

		fmt.Printf("Page %s deleted successfully\n", pageID)
		return nil
	},
}

var pageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pages",
	Long:  "List pages in a Confluence space",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, cfg, err := initClient()
		if err != nil {
			return err
		}

		var pages []api.Page

		if pageParent != "" {
			// List children of a specific parent page
			sortValue, valid := mapChildSortValue(pageSort, pageDesc)
			if !valid {
				return fmt.Errorf("invalid sort value '%s' (valid: web, title, created, modified, id)", pageSort)
			}
			var err error
			pages, err = client.GetChildPages(cmd.Context(), pageParent, pageLimit, sortValue)
			if err != nil {
				return fmt.Errorf("listing child pages: %w", err)
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
			if spaceKey == "" {
				return fmt.Errorf("space key required: use --space flag or set CONFLUENCE_SPACE_KEY")
			}

			sortValue := mapSpaceSortValue(pageSort, pageDesc)
			if sortValue == "" && pageSort != "" {
				return fmt.Errorf("invalid sort value '%s' (valid: title, created, modified, id)", pageSort)
			}

			space, err := client.GetSpace(cmd.Context(), spaceKey)
			if err != nil {
				return fmt.Errorf("getting space: %w", err)
			}

			pages, err = client.ListPages(cmd.Context(), space.ID, pageLimit, sortValue)
			if err != nil {
				return fmt.Errorf("listing pages: %w", err)
			}
		}

		if outputJSON {
			return printJSON(pages)
		}
		spaceKey := pageSpace
		if spaceKey == "" {
			spaceKey = cfg.SpaceKey
		}
		for _, page := range pages {
			fmt.Printf("Title: %s\n", page.Title)
			fmt.Printf("Status: %s\n", page.Status)
			fmt.Printf("URL: %s\n", PageURL(cfg.BaseURL, spaceKey, page.ID))
			fmt.Println("---")
		}
		return nil
	},
}

var pageMoveCmd = &cobra.Command{
	Use:   "move PAGE_ID",
	Short: "Move a page to a new parent",
	Long:  "Move a Confluence page to a new parent page within the same space",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, cfg, err := initClient()
		if err != nil {
			return err
		}

		pageID := args[0]

		if moveParent == "" {
			return fmt.Errorf("--parent flag is required")
		}

		result, err := client.MovePage(cmd.Context(), pageID, moveParent)
		if err != nil {
			return fmt.Errorf("moving page: %w", err)
		}

		if outputJSON {
			return printJSON(result)
		}
		fmt.Println(PageURL(cfg.BaseURL, "", result.ID))
		return nil
	},
}

func readAndValidateContent(pageFile string) ([]byte, error) {
	var content []byte

	if pageFile != "" && pageFile != "-" {
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
		// Read from stdin (either no file specified, or "-" explicitly)
		// Check if stdin is a terminal (no piped input) - skip check if "-" was explicit
		if pageFile != "-" {
			stat, err := stdinStat()
			if err != nil {
				return nil, fmt.Errorf("checking stdin: %w", err)
			}
			if stat.Mode()&os.ModeCharDevice != 0 {
				return nil, fmt.Errorf("content required via --file or pipe")
			}
		}

		// Limit stdin reading
		limitedReader := io.LimitReader(stdinReader, maxContentSize+1)
		var err error
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

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func init() {
	pageCreateCmd.Flags().StringVarP(&pageTitle, "title", "t", "", "Page title (required)")
	pageCreateCmd.Flags().StringVarP(&pageFile, "file", "f", "", "Markdown file, or - for stdin")
	pageCreateCmd.Flags().StringVarP(&pageSpace, "space", "s", "", "Space key (uses config default if not specified)")
	pageCreateCmd.Flags().StringVarP(&pageParent, "parent", "p", "", "Parent page ID")
	pageCreateCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")
	pageCreateCmd.MarkFlagRequired("title")

	pageViewCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	pageUpdateCmd.Flags().StringVarP(&pageTitle, "title", "t", "", "New page title (optional)")
	pageUpdateCmd.Flags().StringVarP(&pageFile, "file", "f", "", "Markdown file, or - for stdin")
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
