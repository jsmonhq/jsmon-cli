package handlers

import (
	"fmt"
	"os"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
	"github.com/jsmonhq/jsmon-cli/utils"
)

// HandleDomainScan scans a domain
func HandleDomainScan(domain, workspaceID, apiKey string, resumeFile string, headers map[string]string) {
	// Extract domain from URL if needed (handles both "example.com" and "https://example.com")
	domain = utils.ExtractDomain(domain)

	// Domain scanning does not support resume functionality
	// Resume is only available for file scanning

	client := api.NewClient(apiKey, headers)
	err := client.ScanDomain(domain, workspaceID)
	if err != nil {
		// Check if it's an APIError
		if apiErr, ok := err.(*api.APIError); ok {
			// Check for authentication error (wrong or missing API key)
			if apiErr.IsAuthError() {
				fmt.Fprintf(os.Stderr, "Error: Invalid API key. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
				os.Exit(1)
			}

			// Check if it's an insufficient limits error
			if apiErr.IsInsufficientLimitsError() {
				fmt.Fprintf(os.Stderr, "%sInsufficient scan limits%s\n", ColorRed, ColorReset)
				fmt.Fprintf(os.Stderr, "%sPlease add scan limits and try again%s\n", ColorRed, ColorReset)
				os.Exit(1)
			}

			// Check if it's a rate limit error
			if apiErr.IsRateLimitError() {
				fmt.Fprintf(os.Stderr, "%sRate limit reached for domain: %s%s\n", ColorRed, domain, ColorReset)
				fmt.Fprintf(os.Stderr, "%sPlease try again later%s\n", ColorRed, ColorReset)
				os.Exit(1)
			}
		}

		// Regular error handling - always show the actual error

		if apiErr, ok := err.(*api.APIError); ok {
			reason := strings.TrimSpace(apiErr.Message)
			// Show full error message including status code if available
			if apiErr.Status > 0 {
				fmt.Fprintf(os.Stderr, "%sError scanning domain: %s - [HTTP %d] %s%s\n", ColorRed, domain, apiErr.Status, reason, ColorReset)
			} else {
				fmt.Fprintf(os.Stderr, "%sError scanning domain: %s - %s%s\n", ColorRed, domain, reason, ColorReset)
			}
		} else {
			fmt.Fprintf(os.Stderr, "%sError scanning domain: %s - %v%s\n", ColorRed, domain, err, ColorReset)
		}
		os.Exit(1)
	}

	fmt.Printf("%sDomain scan completed for - %s%s\n", ColorGreen, domain, ColorReset)
}
