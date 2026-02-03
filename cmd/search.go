package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/grantcarthew/acon/internal/api"
	"github.com/spf13/cobra"
)

var (
	searchTitle   string
	searchLabel   string
	searchCreator string
	searchSpace   string
	searchLimit   int
	searchStart   int
	searchType    string
	searchCQL     string
)

var searchCmd = &cobra.Command{
	Use:   "search [QUERY]",
	Short: "Search Confluence content",
	Long:  "Search Confluence content using CQL (Confluence Query Language). Supports simple flags for common searches or raw CQL for advanced queries.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, cfg, err := initClient()
		if err != nil {
			return err
		}

		var cql string
		var textQuery string

		// Get positional text query if provided
		if len(args) > 0 {
			textQuery = args[0]
		}

		// Validate mutually exclusive options
		if searchCQL != "" && (textQuery != "" || searchTitle != "" || searchLabel != "" || searchCreator != "" || searchSpace != "" || searchType != "") {
			var conflicts []string
			if textQuery != "" {
				conflicts = append(conflicts, "QUERY")
			}
			if searchTitle != "" {
				conflicts = append(conflicts, "--title")
			}
			if searchLabel != "" {
				conflicts = append(conflicts, "--label")
			}
			if searchCreator != "" {
				conflicts = append(conflicts, "--creator")
			}
			if searchSpace != "" {
				conflicts = append(conflicts, "--space")
			}
			if searchType != "" {
				conflicts = append(conflicts, "--type")
			}
			return fmt.Errorf("--cql flag cannot be combined with other search flags (specified: %s)", strings.Join(conflicts, ", "))
		}

		// Use raw CQL if provided, otherwise build from flags
		if searchCQL != "" {
			cql = searchCQL
		} else {
			// Build CQL from search parameters
			spaceKey := searchSpace
			if spaceKey == "" {
				spaceKey = cfg.SpaceKey
			}

			params := api.SearchParams{
				Text:    textQuery,
				Title:   searchTitle,
				Label:   searchLabel,
				Creator: searchCreator,
				Space:   spaceKey,
				Type:    searchType,
			}

			var err error
			cql, err = api.BuildCQL(params)
			if err != nil {
				return fmt.Errorf("invalid search parameters: %w", err)
			}
		}

		// Execute search
		result, hasMore, err := client.Search(cmd.Context(), cql, searchLimit, searchStart)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		// Output results
		if outputJSON {
			return printJSON(result)
		}

		// Human-readable output
		if len(result.Results) == 0 {
			fmt.Println("No results found")
			return nil
		}

		for i, searchResult := range result.Results {
			// Title with space key
			spaceKey := searchResult.Content.Space.Key
			fmt.Printf("%s (%s)\n", searchResult.Title, spaceKey)

			// Full URL - construct from base URL
			if searchResult.URL != "" {
				// Handle both relative and absolute URLs
				var fullURL string
				if strings.HasPrefix(searchResult.URL, "http://") || strings.HasPrefix(searchResult.URL, "https://") {
					// Absolute URL - use as-is
					fullURL = searchResult.URL
				} else if strings.HasPrefix(searchResult.URL, "/") {
					// Relative URL - append to base (already validated above)
					fullURL = strings.TrimRight(cfg.BaseURL, "/") + searchResult.URL
				} else {
					// Invalid format - warn user and skip (API contract issue)
					fmt.Fprintf(os.Stderr, "Warning: Skipping malformed URL for '%s': %s\n", searchResult.Title, searchResult.URL)
					fullURL = ""
				}

				if fullURL != "" {
					fmt.Printf("%s\n", fullURL)
				}
			}

			// Excerpt
			if searchResult.Excerpt != "" {
				fmt.Printf("%s\n", searchResult.Excerpt)
			}

			// Modified date
			if searchResult.LastModified != "" {
				// Parse and format the date
				t, err := time.Parse(time.RFC3339, searchResult.LastModified)
				if err != nil {
					// Log warning in verbose mode only
					if verbose {
						fmt.Fprintf(os.Stderr, "Warning: Could not parse date for '%s' (raw: %s, error: %v)\n",
							searchResult.Title, searchResult.LastModified, err)
					}
					// Show "Unknown" instead of potentially malformed data
					fmt.Printf("Modified: Unknown\n")
				} else {
					fmt.Printf("Modified: %s\n", t.Format("2006-01-02"))
				}
			}

			// Separator between results (but not after the last one)
			if i < len(result.Results)-1 {
				fmt.Println()
			}
		}

		// Pagination summary - consistent with page list command
		fmt.Println()
		if hasMore {
			// Safe to do arithmetic - validated above
			nextStart := result.Start + result.Size
			fmt.Printf("Showing %d of %d results (more available - use --start %d for next page)\n",
				len(result.Results), result.TotalSize, nextStart)
		} else {
			fmt.Printf("Showing all %d results\n", result.TotalSize)
		}

		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchTitle, "title", "", "Search in page titles")
	searchCmd.Flags().StringVar(&searchLabel, "label", "", "Search by label (exact match)")
	searchCmd.Flags().StringVar(&searchCreator, "creator", "", "Filter by creator (email or 'me')")
	searchCmd.Flags().StringVarP(&searchSpace, "space", "s", "", "Filter by space key (uses config default if not specified)")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", api.DefaultSearchLimit, "Maximum number of results per page")
	searchCmd.Flags().IntVar(&searchStart, "start", 0, "Starting index for pagination (0-based)")
	searchCmd.Flags().StringVar(&searchType, "type", "", "Content type (page, blogpost, attachment, etc.)")
	searchCmd.Flags().StringVar(&searchCQL, "cql", "", "Raw CQL query (overrides all other flags)")
	searchCmd.Flags().BoolVarP(&outputJSON, "json", "j", false, "Output as JSON")

	rootCmd.AddCommand(searchCmd)
}
