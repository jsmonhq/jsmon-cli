package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
)

// ParamValue represents a parameter value with url and parameters
type ParamValue struct {
	URL        string                   `json:"url"`
	Parameters []map[string]interface{} `json:"parameters"`
}

// DomainStatusValue represents a domain status value with domainName, status, and expiryDate
type DomainStatusValue struct {
	DomainName string `json:"domainName"`
	Status     string `json:"status"`
	ExpiryDate string `json:"expiryDate"`
}

// HandleJSIntelligence displays reconnaissance data for a workspace in raw JSON format
func HandleJSIntelligence(workspaceID, apiKey string, headers map[string]string, field string, page int, limit int) {
	client := api.NewClient(apiKey, headers)

	// Get raw JSON response from API
	rawJSON, err := client.GetJSIntelligenceRaw(workspaceID, field, page, "", "", "", limit)
	if err != nil {
		// Check for authentication error (wrong or missing API key)
		if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
			fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "%sError fetching reconnaissance data: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Parse JSON to extract only values
	var response map[string]interface{}
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		fmt.Fprintf(os.Stderr, "%sError parsing JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Check if this is a field with object values (not strings)
	isParamField := strings.ToLower(field) == "param"
	isDomainStatusField := strings.ToLower(field) == "activedomains" || strings.ToLower(field) == "inactivedomains"
	isAwsAssetsField := strings.ToLower(field) == "awsassets" || strings.ToLower(field) == "allawsassets"

	var output interface{}
	if isParamField {
		// For param field, extract the full value objects and ensure url comes before parameters
		var paramValues []ParamValue
		if data, ok := response["data"].([]interface{}); ok {
			for _, item := range data {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if value, exists := itemMap["value"]; exists && value != nil {
						if valueMap, ok := value.(map[string]interface{}); ok {
							paramValue := ParamValue{}
							if url, exists := valueMap["url"]; exists {
								if urlStr, ok := url.(string); ok {
									paramValue.URL = urlStr
								}
							}
							if params, exists := valueMap["parameters"]; exists {
								if paramsArray, ok := params.([]interface{}); ok {
									paramValue.Parameters = make([]map[string]interface{}, 0, len(paramsArray))
									for _, p := range paramsArray {
										if paramMap, ok := p.(map[string]interface{}); ok {
											paramValue.Parameters = append(paramValue.Parameters, paramMap)
										}
									}
								}
							}
							paramValues = append(paramValues, paramValue)
						}
					}
				}
			}
		}
		output = paramValues
	} else if isDomainStatusField {
		// For domain status fields, extract the full value objects
		var domainValues []DomainStatusValue
		if data, ok := response["data"].([]interface{}); ok {
			for _, item := range data {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if value, exists := itemMap["value"]; exists && value != nil {
						if valueMap, ok := value.(map[string]interface{}); ok {
							domainValue := DomainStatusValue{}
							if domainName, exists := valueMap["domainName"]; exists {
								if domainNameStr, ok := domainName.(string); ok {
									domainValue.DomainName = domainNameStr
								}
							}
							if status, exists := valueMap["status"]; exists {
								if statusStr, ok := status.(string); ok {
									domainValue.Status = statusStr
								}
							}
							if expiryDate, exists := valueMap["expiryDate"]; exists {
								// expiryDate can be a string or null, handle both
								if expiryDateStr, ok := expiryDate.(string); ok {
									domainValue.ExpiryDate = expiryDateStr
								} else if expiryDate == nil {
									domainValue.ExpiryDate = "NULL"
								}
							}
							domainValues = append(domainValues, domainValue)
						}
					}
				}
			}
		}
		output = domainValues
	} else if isAwsAssetsField {
		// For awsassets field, extract the full value objects (AWS assets are objects)
		var awsAssets []map[string]interface{}
		if data, ok := response["data"].([]interface{}); ok {
			for _, item := range data {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if value, exists := itemMap["value"]; exists && value != nil {
						if valueMap, ok := value.(map[string]interface{}); ok {
							awsAssets = append(awsAssets, valueMap)
						} else if valueStr, ok := value.(string); ok && valueStr != "" {
							// If value is a string, wrap it in a map
							awsAssets = append(awsAssets, map[string]interface{}{"value": valueStr})
						}
					}
				}
			}
		}
		output = awsAssets
	} else {
		// For other fields, extract string values
		var values []string
		if data, ok := response["data"].([]interface{}); ok {
			for _, item := range data {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if value, ok := itemMap["value"].(string); ok && value != "" {
						values = append(values, value)
					}
				}
			}
		}

		// Check if this is a GraphQL field that needs \n replacement
		isGraphQLField := strings.ToLower(field) == "gqlqueries" ||
			strings.ToLower(field) == "gqlmutaions" ||
			strings.ToLower(field) == "gqlmutations" ||
			strings.ToLower(field) == "gqlfragments"

		// Replace \n with actual newlines for GraphQL fields
		if isGraphQLField {
			for i := range values {
				values[i] = strings.ReplaceAll(values[i], "\\n", "\n")
			}
		}

		output = values
	}

	// Output as a JSON array so jq can parse it
	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
