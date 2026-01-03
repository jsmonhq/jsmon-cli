package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleDomains displays domain scans for a workspace in JSON format
func HandleDomains(workspaceID, apiKey string, headers map[string]string, page int, status, search, scoreMin, scoreMax, dateFrom, dateTo, limit, monitoring string) {
	client := api.NewClient(apiKey, headers)

	response, err := client.GetDomainScans(workspaceID, page, status, search, scoreMin, scoreMax, dateFrom, dateTo, limit, monitoring)
	if err != nil {
		// Check for authentication error (wrong or missing API key)
		if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
			fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%sError fetching domain scans: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Extract domains into a simple string slice
	var domains []string
	for _, domainScan := range response.Data.Scans {
		if domainScan.Asset != "" {
			domains = append(domains, domainScan.Asset)
		}
	}

	// Output as a JSON array
	jsonOutput, err := json.MarshalIndent(domains, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}

