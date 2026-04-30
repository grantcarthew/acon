package cli

import (
	"bytes"
	"context"
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

// pageURL returns the canonical browse URL for a Confluence page.
// Caller must supply a non-empty spaceKey; an empty key produces a malformed URL.
func pageURL(baseURL, spaceKey, pageID string) string {
	return fmt.Sprintf("%s/wiki/spaces/%s/pages/%s", baseURL, spaceKey, pageID)
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

		if verbose {
			fmt.Fprintf(os.Stderr, "[Page Create] Resolving space: %s\n", spaceKey)
		}

		space, err := client.GetSpace(cmd.Context(), spaceKey)
		if err != nil {
			return fmt.Errorf("getting space: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[Page Create] Space ID: %s\n", space.ID)
		}

		content, err := readAndValidateContent(pageFile)
		if err != nil {
			return err
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[Page Create] Read %d bytes of markdown content\n", len(content))
			fmt.Fprintf(os.Stderr, "[Page Create] Converting markdown to Confluence storage format\n")
		}

		htmlContent := converter.MarkdownToStorage(string(content))

		if verbose {
			fmt.Fprintf(os.Stderr, "[Page Create] Converted to %d bytes of storage format\n", len(htmlContent))
		}

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
			if verbose {
				fmt.Fprintf(os.Stderr, "[Page Create] Setting parent ID: %s\n", pageParent)
			}
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[Page Create] Creating page: %s\n", pageTitle)
		}

		result, err := client.CreatePage(cmd.Context(), req)
		if err != nil {
			return fmt.Errorf("creating page: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[Page Create] Page created successfully, ID: %s\n", result.ID)
		}

		if outputJSON {
			return printJSON(result)
		}
		fmt.Println(pageURL(cfg.BaseURL, spaceKey, result.ID))
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

		if verbose {
			fmt.Fprintf(os.Stderr, "[Page View] Fetching page: %s\n", pageID)
		}

		page, err := client.GetPage(cmd.Context(), pageID)
		if err != nil {
			return fmt.Errorf("getting page: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "[Page View] Page title: %s\n", page.Title)
		}

		if outputJSON {
			return printJSON(page)
		}
		if page.Body != nil && page.Body.Storage != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "[Page View] Converting %d bytes from storage to markdown\n", len(page.Body.Storage.Value))
			}
			markdown, err := converter.StorageToMarkdown(page.Body.Storage.Value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to convert to markdown: %v\n", err)
				fmt.Println(page.Body.Storage.Value)
			} else {
				if verbose {
					fmt.Fprintf(os.Stderr, "[Page View] Converted to %d bytes of markdown\n", len(markdown))
				}
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
		space, err := client.GetSpaceByID(cmd.Context(), result.SpaceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: page updated but could not resolve space key for URL: %v\n", err)
			fmt.Println(result.ID)
			return nil
		}
		if space.Key == "" {
			fmt.Fprintf(os.Stderr, "Warning: page updated but space %s returned empty key\n", result.SpaceID)
			fmt.Println(result.ID)
			return nil
		}
		fmt.Println(pageURL(cfg.BaseURL, space.Key, result.ID))
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

		var (
			pages         []api.Page
			hasMore       bool
			spaceKeyCache map[string]string
		)

		if pageParent != "" {
			pages, hasMore, spaceKeyCache, err = listChildPages(cmd.Context(), client)
		} else {
			pages, hasMore, spaceKeyCache, err = listPagesBySpace(cmd.Context(), client, cfg)
		}
		if err != nil {
			return err
		}

		if outputJSON {
			return printJSON(pages)
		}

		return printPageList(cmd.Context(), client, os.Stdout, cfg.BaseURL, pages, hasMore, spaceKeyCache)
	},
}

// listPagesBySpace fetches pages in a space using the user-supplied or configured
// space key. The returned cache is primed with the resolved space so the printer
// avoids a redundant lookup.
func listPagesBySpace(ctx context.Context, client *api.Client, cfg *config.Config) ([]api.Page, bool, map[string]string, error) {
	spaceKey := pageSpace
	if spaceKey == "" {
		spaceKey = cfg.SpaceKey
	}
	if spaceKey == "" {
		return nil, false, nil, fmt.Errorf("space key required: use --space flag or set CONFLUENCE_SPACE_KEY")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[Page List] Listing pages in space: %s (limit: %d, sort: %s)\n", spaceKey, pageLimit, pageSort)
	}

	sortValue := mapSpaceSortValue(pageSort, pageDesc)
	if sortValue == "" && pageSort != "" {
		return nil, false, nil, fmt.Errorf("invalid sort value '%s' (valid: title, created, modified, id)", pageSort)
	}

	space, err := client.GetSpace(ctx, spaceKey)
	if err != nil {
		return nil, false, nil, fmt.Errorf("getting space: %w", err)
	}

	pages, hasMore, err := client.ListPages(ctx, space.ID, pageLimit, sortValue)
	if err != nil {
		return nil, false, nil, fmt.Errorf("listing pages: %w", err)
	}

	return pages, hasMore, map[string]string{space.ID: spaceKey}, nil
}

// listChildPages fetches children of a specific parent page. The returned cache
// is empty; the printer populates it on first miss.
func listChildPages(ctx context.Context, client *api.Client) ([]api.Page, bool, map[string]string, error) {
	if verbose {
		fmt.Fprintf(os.Stderr, "[Page List] Listing children of parent: %s (limit: %d, sort: %s)\n", pageParent, pageLimit, pageSort)
	}

	sortValue, valid := mapChildSortValue(pageSort, pageDesc)
	if !valid {
		return nil, false, nil, fmt.Errorf("invalid sort value '%s' (valid: web, title, created, modified, id)", pageSort)
	}

	pages, hasMore, err := client.GetChildPages(ctx, pageParent, pageLimit, sortValue)
	if err != nil {
		return nil, false, nil, fmt.Errorf("listing child pages: %w", err)
	}

	if pageSort == "title" {
		if verbose {
			fmt.Fprintf(os.Stderr, "[Page List] Performing client-side title sort\n")
		}
		sort.Slice(pages, func(i, j int) bool {
			if pageDesc {
				return strings.ToLower(pages[i].Title) > strings.ToLower(pages[j].Title)
			}
			return strings.ToLower(pages[i].Title) < strings.ToLower(pages[j].Title)
		})
	}

	return pages, hasMore, map[string]string{}, nil
}

// printPageList renders a human-readable listing, resolving any space IDs not
// already present in the cache.
func printPageList(ctx context.Context, client *api.Client, out io.Writer, baseURL string, pages []api.Page, hasMore bool, spaceKeyCache map[string]string) error {
	for _, page := range pages {
		key, ok := spaceKeyCache[page.SpaceID]
		if !ok {
			space, err := client.GetSpaceByID(ctx, page.SpaceID)
			switch {
			case err != nil:
				fmt.Fprintf(os.Stderr, "Warning: could not resolve space key for page %s: %v\n", page.ID, err)
				// Negative-cache the miss so we do not repeat the lookup for every page in the same space.
				spaceKeyCache[page.SpaceID] = ""
			case space.Key == "":
				fmt.Fprintf(os.Stderr, "Warning: space %s returned empty key for page %s\n", page.SpaceID, page.ID)
				spaceKeyCache[page.SpaceID] = ""
			default:
				key = space.Key
				spaceKeyCache[page.SpaceID] = key
			}
		}
		fmt.Fprintf(out, "Title: %s\n", page.Title)
		fmt.Fprintf(out, "Status: %s\n", page.Status)
		if key == "" {
			fmt.Fprintf(out, "URL: (unresolved, page ID: %s)\n", page.ID)
		} else {
			fmt.Fprintf(out, "URL: %s\n", pageURL(baseURL, key, page.ID))
		}
		fmt.Fprintln(out, "---")
	}

	resultWord := "results"
	if len(pages) == 1 {
		resultWord = "result"
	}
	if hasMore {
		fmt.Fprintf(out, "\nShowing %d %s (more available - increase --limit to see more)\n", len(pages), resultWord)
	} else {
		fmt.Fprintf(out, "\nShowing all %d %s\n", len(pages), resultWord)
	}
	return nil
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
		space, err := client.GetSpaceByID(cmd.Context(), result.SpaceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: page moved but could not resolve space key for URL: %v\n", err)
			fmt.Println(result.ID)
			return nil
		}
		if space.Key == "" {
			fmt.Fprintf(os.Stderr, "Warning: page moved but space %s returned empty key\n", result.SpaceID)
			fmt.Println(result.ID)
			return nil
		}
		fmt.Println(pageURL(cfg.BaseURL, space.Key, result.ID))
		return nil
	},
}

func readAndValidateContent(pageFile string) ([]byte, error) {
	var content []byte

	if pageFile != "" && pageFile != "-" {
		if verbose {
			fmt.Fprintf(os.Stderr, "[Content] Reading from file: %s\n", pageFile)
		}
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
		if verbose {
			fmt.Fprintf(os.Stderr, "[Content] Read %d bytes from file\n", len(content))
		}
	} else {
		if verbose {
			fmt.Fprintf(os.Stderr, "[Content] Reading from stdin\n")
		}
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
		if verbose {
			fmt.Fprintf(os.Stderr, "[Content] Read %d bytes from stdin\n", len(content))
		}
	}

	content = bytes.TrimSpace(content)
	if len(content) == 0 {
		return nil, fmt.Errorf("content cannot be empty")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[Content] Content validated: %d bytes (after trimming)\n", len(content))
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
	if err := pageCreateCmd.MarkFlagRequired("title"); err != nil {
		panic(err)
	}

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
	if err := pageMoveCmd.MarkFlagRequired("parent"); err != nil {
		panic(err)
	}

	pageCmd.AddCommand(pageCreateCmd)
	pageCmd.AddCommand(pageViewCmd)
	pageCmd.AddCommand(pageUpdateCmd)
	pageCmd.AddCommand(pageDeleteCmd)
	pageCmd.AddCommand(pageListCmd)
	pageCmd.AddCommand(pageMoveCmd)
}
