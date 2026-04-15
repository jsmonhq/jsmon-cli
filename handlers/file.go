package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/api"
)

// FileScanResult represents the result of a file scan
type FileScanResult struct {
	Success []string               `json:"success"`
	Failed  []map[string]string    `json:"failed"`
	Summary map[string]interface{} `json:"summary"`
}

// HandleFileUpload uploads a file for scanning
// The file should contain one JS URL per line
func HandleFileUpload(filePath, workspaceID, apiKey string, resumeFile string, headers map[string]string) {
	if resumeFile != "" {
		fmt.Fprintf(os.Stderr, "%sResume is not supported for file uploads. Re-run the file scan with -f.%s\n", ColorRed, ColorReset)
		os.Exit(1)
	}

	fmt.Printf("Scanning started for - %s\n", filePath)

	client := api.NewClient(apiKey, headers)
	if err := client.UploadFile(filePath, workspaceID); err != nil {
		if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
			fmt.Fprintf(os.Stderr, "Error: API key is invalid. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "%sError uploading file: %s - %v%s\n", ColorRed, filePath, err, ColorReset)
		result := FileScanResult{
			Success: nil,
			Failed: []map[string]string{
				{
					"url":   filePath,
					"error": err.Error(),
				},
			},
			Summary: map[string]interface{}{
				"total":   1,
				"success": 0,
				"failed":  1,
				"lastUrl": "",
			},
		}
		jsonOutput, marshalErr := json.MarshalIndent(result, "", "  ")
		if marshalErr == nil {
			fmt.Println(string(jsonOutput))
		}
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "%sFile scan submitted: %s%s\n", ColorGreen, filePath, ColorReset)
	result := FileScanResult{
		Success: []string{filePath},
		Failed:  nil,
		Summary: map[string]interface{}{
			"total":   1,
			"success": 1,
			"failed":  0,
			"lastUrl": "",
		},
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
