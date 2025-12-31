package handlers

import (
	"fmt"
	"os"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleURLUpload uploads a URL for scanning
func HandleURLUpload(jsURL, workspaceID, apiKey string, headers map[string]string) {
	fmt.Printf("Scanning started for - %s\n", jsURL)

	client := api.NewClient(apiKey, headers)

	err := client.UploadURL(jsURL, workspaceID)
	if err != nil {
		// Check if it's an APIError to extract URL and message
		if apiErr, ok := err.(*api.APIError); ok {
			// Check for insufficient limits error
			if apiErr.IsInsufficientLimitsError() {
				fmt.Fprintf(os.Stderr, "%sInsufficient scan limits%s\n", ColorRed, ColorReset)
				fmt.Fprintf(os.Stderr, "%sPlease add scan limits and try again%s\n", ColorRed, ColorReset)
				os.Exit(1)
			}

			// Check for rate limit error
			if apiErr.IsRateLimitError() {
				fmt.Fprintf(os.Stderr, "%sRate limit reached%s\n", ColorRed, ColorReset)
				fmt.Fprintf(os.Stderr, "%sPlease try again later%s\n", ColorRed, ColorReset)
				os.Exit(1)
			}

			reason := strings.TrimSpace(apiErr.Message)
			// Clean up the message - remove "URL Scan failed - " prefix if present
			if strings.HasPrefix(reason, "URL Scan failed - ") {
				reason = strings.TrimPrefix(reason, "URL Scan failed - ")
				reason = strings.TrimSpace(reason)
			}
			// Also check for old "JS Scan failed - " format for backward compatibility
			if strings.HasPrefix(reason, "JS Scan failed - ") {
				reason = strings.TrimPrefix(reason, "JS Scan failed - ")
				reason = strings.TrimSpace(reason)
			}
			// Remove "Please provide a valid URL." from the end if present
			reason = strings.TrimSuffix(reason, " Please provide a valid URL.")
			reason = strings.TrimSuffix(reason, "Please provide a valid URL.")
			// Also check for old "JS file URL" format for backward compatibility
			reason = strings.TrimSuffix(reason, " Please provide a valid JS file URL.")
			reason = strings.TrimSuffix(reason, "Please provide a valid JS file URL.")
			reason = strings.TrimSpace(reason)
			fmt.Fprintf(os.Stderr, "%sError uploading url: %s - %s%s\n", ColorRed, apiErr.URL, reason, ColorReset)
		} else {
			// Fallback for other errors
			fmt.Fprintf(os.Stderr, "%sError uploading url: %s - %v%s\n", ColorRed, jsURL, err, ColorReset)
		}
		os.Exit(1)
	}

	fmt.Printf("%sURL scan completed for - %s%s\n", ColorGreen, jsURL, ColorReset)
}
