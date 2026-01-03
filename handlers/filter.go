package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleFilter displays filtered reconnaissance data for a workspace
func HandleFilter(workspaceID, apiKey string, headers map[string]string, fieldname, keyword string, page int) {
	client := api.NewClient(apiKey, headers)

	// Use GetJSIntelligence with search parameter
	response, err := client.GetJSIntelligence(workspaceID, fieldname, page, "", keyword, "")
	if err != nil {
		// Check for authentication error (wrong or missing API key)
		if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
			fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%sError fetching filtered data: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Output the data as JSON array
	var output []interface{}

	// Check if this is a field with object values (not strings)
	isParamField := strings.ToLower(fieldname) == "param"

	for _, item := range response.Data {
		if isParamField {
			// For param field, parse the value as ParamValue
			var paramValue ParamValue
			// item.Value is a string, try to unmarshal it as JSON
			if err := json.Unmarshal([]byte(item.Value), &paramValue); err == nil {
				output = append(output, paramValue)
			} else {
				// If parsing fails, just add the string value
				output = append(output, item.Value)
			}
		} else {
			// For other fields, check if value needs special handling
			fieldLower := strings.ToLower(fieldname)
			if fieldLower == "gqlqueries" || fieldLower == "gqlmutaions" || fieldLower == "gqlmutations" || fieldLower == "gqlfragments" {
				// Replace \n with actual newlines for GraphQL fields
				valueStr := strings.ReplaceAll(item.Value, "\\n", "\n")
				output = append(output, valueStr)
			} else {
				// For simple string fields, just add the value
				output = append(output, item.Value)
			}
		}
	}

	// Marshal to JSON
	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}

