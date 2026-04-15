package handlers

import (
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleCodeScan uploads a source code file for code scanning.
func HandleCodeScan(filePath, workspaceID, apiKey string, headers map[string]string) {
	fmt.Printf("Code scan started for - %s\n", filePath)

	client := api.NewClient(apiKey, headers)
	err := client.UploadCodeFile(filePath, workspaceID)
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
			fmt.Fprintf(os.Stderr, "%sError uploading code file: %s - %s%s\n", ColorRed, filePath, apiErr.Message, ColorReset)
		} else {
			fmt.Fprintf(os.Stderr, "%sError uploading code file: %s - %v%s\n", ColorRed, filePath, err, ColorReset)
		}
		os.Exit(1)
	}

	fmt.Printf("%sCode scan completed for - %s%s\n", ColorGreen, filePath, ColorReset)
}
