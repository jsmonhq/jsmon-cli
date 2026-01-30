package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleJSURLs displays JS URLs for a workspace in JSON format
func HandleJSURLs(workspaceID, apiKey string, headers map[string]string, page int, runID, search, status string, limit int) {
	client := api.NewClient(apiKey, headers)

	response, err := client.GetJSURLs(workspaceID, page, runID, search, status, limit)
	if err != nil {
		// Check for authentication error (wrong or missing API key)
		if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
			fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%sError fetching JS URLs: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Extract URLs into a simple string slice
	var urls []string
	for _, jsurl := range response.Data {
		if jsurl.Value != "" {
			urls = append(urls, jsurl.Value)
		}
	}

	// Output as a JSON array
	jsonOutput, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
