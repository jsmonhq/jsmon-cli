package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleFilter displays filtered reconnaissance data for a workspace
func HandleFilter(workspaceID, apiKey string, headers map[string]string, fieldname, keyword string, page int, limit int) {
	client := api.NewClient(apiKey, headers)

	// param (and other object-value fields) return data.value as object; use raw to avoid unmarshal error
	isParamField := strings.ToLower(fieldname) == "param"

	var output []interface{}

	if isParamField {
		rawJSON, err := client.GetJSIntelligenceRaw(workspaceID, fieldname, page, "", keyword, "", limit)
		if err != nil {
			if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
				fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "%sError fetching filtered data: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}
		var response map[string]interface{}
		if err := json.Unmarshal(rawJSON, &response); err != nil {
			fmt.Fprintf(os.Stderr, "%sError parsing JSON: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}
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
							output = append(output, paramValue)
						}
					}
				}
			}
		}
	} else {
		response, err := client.GetJSIntelligence(workspaceID, fieldname, page, "", keyword, "", limit)
		if err != nil {
			if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
				fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "%sError fetching filtered data: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}
		fieldLower := strings.ToLower(fieldname)
		for _, item := range response.Data {
			if fieldLower == "gqlqueries" || fieldLower == "gqlmutaions" || fieldLower == "gqlmutations" || fieldLower == "gqlfragments" {
				valueStr := strings.ReplaceAll(item.Value, "\\n", "\n")
				output = append(output, valueStr)
			} else {
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
