package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleIssues fetches workspace issues and prints them in JSON format.
func HandleIssues(workspaceID, apiKey string, headers map[string]string, options api.IssuesQueryOptions) {
	client := api.NewClient(apiKey, headers)

	response, err := client.GetIssues(workspaceID, options)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
			fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%sError fetching issues: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	jsonOutput, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
