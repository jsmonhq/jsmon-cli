package handlers

import (
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/v2/api"
)

// FileScanResult represents the result of a file scan
type FileScanResult struct {
	Success []string               `json:"success"`
	Failed  []map[string]string    `json:"failed"`
	Summary map[string]interface{} `json:"summary"`
}

// HandleFileUpload uploads a file of URLs for server-side file scanning.
func HandleFileUpload(filePath, workspaceID, apiKey string, headers map[string]string, options api.ScanSubmitOptions) {
	fmt.Printf("Scanning started for - %s\n", filePath)

	client := api.NewClient(apiKey, headers)
	response, err := client.UploadFile(filePath, workspaceID, options)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			if apiErr.IsAuthError() {
				fmt.Fprintf(os.Stderr, "Error: API key is invalid. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
				os.Exit(1)
			}
			if apiErr.IsInsufficientLimitsError() {
				fmt.Fprintf(os.Stderr, "%sInsufficient scan limits%s\n", ColorRed, ColorReset)
				fmt.Fprintf(os.Stderr, "%sPlease add scan limits and try again%s\n", ColorRed, ColorReset)
				os.Exit(1)
			}
			if apiErr.IsRateLimitError() {
				fmt.Fprintf(os.Stderr, "%sRate limit reached%s\n", ColorRed, ColorReset)
				fmt.Fprintf(os.Stderr, "%sPlease try again later%s\n", ColorRed, ColorReset)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "%sError uploading file: %s - %s%s\n", ColorRed, filePath, apiErr.Message, ColorReset)
		} else {
			fmt.Fprintf(os.Stderr, "%sError uploading file: %s - %v%s\n", ColorRed, filePath, err, ColorReset)
		}
		os.Exit(1)
	}

	printScanQueued("File scan", filePath, response)
}
