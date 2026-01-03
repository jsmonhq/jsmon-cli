package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/api"
)

// SecretOutput represents the output format for secrets (without source and occurrences, with reordered fields)
type SecretOutput struct {
	ModuleName  string `json:"moduleName"`
	MatchedWord string `json:"matchedWord"`
	Severity    string `json:"severity"`
	CreatedAt   string `json:"createdAt"`
}

// SecretsResponseOutput represents the output format for secrets response
type SecretsResponseOutput struct {
	Data       []SecretOutput `json:"data"`
	Pagination struct {
		CurrentPage     int  `json:"currentPage"`
		TotalItems      int  `json:"totalItems"`
		TotalPages      int  `json:"totalPages"`
		ItemsPerPage    int  `json:"itemsPerPage"`
		HasNextPage     bool `json:"hasNextPage"`
		HasPreviousPage bool `json:"hasPreviousPage"`
	} `json:"pagination"`
}

// HandleSecrets displays secrets for a workspace in JSON format
func HandleSecrets(workspaceID, apiKey string, headers map[string]string, page int, runID, lastScannedOn, formDate, toDate, limit, search string) {
	client := api.NewClient(apiKey, headers)

	response, err := client.GetSecrets(workspaceID, page, runID, lastScannedOn, formDate, toDate, limit, search)
	if err != nil {
		// Check for authentication error (wrong or missing API key)
		if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
			fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%sError fetching secrets: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Convert to output format (exclude source and occurrences, reorder fields)
	output := SecretsResponseOutput{
		Data: make([]SecretOutput, 0, len(response.Data)),
	}
	output.Pagination = response.Pagination

	for _, secret := range response.Data {
		output.Data = append(output.Data, SecretOutput{
			ModuleName:  secret.ModuleName,
			MatchedWord: secret.MatchedWord,
			Severity:    secret.Severity,
			CreatedAt:   secret.CreatedAt,
		})
	}

	// Display secrets in JSON format
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}
