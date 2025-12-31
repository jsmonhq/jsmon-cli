package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
)

// URLItem represents a URL item with url before createdAt
type URLItem struct {
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
}

// filterResourceIdAndReorder recursively removes resourceId and reorders fields in urls array
func filterResourceIdAndReorder(data map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})

	// Process each field
	for key, value := range data {
		if key == "resourceId" {
			continue // Skip resourceId
		}

		// Special handling for urls array - reorder fields
		if key == "urls" {
			if nestedArray, ok := value.([]interface{}); ok {
				urlItems := make([]URLItem, 0, len(nestedArray))
				for _, item := range nestedArray {
					if itemMap, ok := item.(map[string]interface{}); ok {
						urlItem := URLItem{}
						if url, exists := itemMap["url"]; exists {
							if urlStr, ok := url.(string); ok {
								urlItem.URL = urlStr
							}
						}
						if createdAt, exists := itemMap["createdAt"]; exists {
							if createdAtStr, ok := createdAt.(string); ok {
								urlItem.CreatedAt = createdAtStr
							}
						}
						urlItems = append(urlItems, urlItem)
					}
				}
				filtered[key] = urlItems
				continue
			}
		}

		// Recursively process nested maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			filtered[key] = filterResourceIdAndReorder(nestedMap)
		} else if nestedArray, ok := value.([]interface{}); ok {
			// Process arrays that might contain maps
			filteredArray := make([]interface{}, 0, len(nestedArray))
			for _, item := range nestedArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					filteredArray = append(filteredArray, filterResourceIdAndReorder(itemMap))
				} else {
					filteredArray = append(filteredArray, item)
				}
			}
			filtered[key] = filteredArray
		} else {
			filtered[key] = value
		}
	}
	return filtered
}

// HandleReverseSearch displays reverse search results in JSON format
func HandleReverseSearch(workspaceID, apiKey string, headers map[string]string, field, searchValue string) {
	// Convert literal \n to actual newlines for GraphQL queries
	// This handles cases where the shell passes \n as literal characters
	// Also handle other escape sequences that might be in the value
	searchValue = strings.ReplaceAll(searchValue, "\\n", "\n")
	searchValue = strings.ReplaceAll(searchValue, "\\t", "\t")
	searchValue = strings.ReplaceAll(searchValue, "\\r", "\r")

	client := api.NewClient(apiKey, headers)

	response, err := client.ReverseSearch(workspaceID, field, searchValue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError performing reverse search: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Check if data is empty
	if response.Data == nil || len(response.Data) == 0 {
		fmt.Println("[]")
		return
	}

	// Recursively filter out resourceId and reorder fields in urls array
	filteredData := make([]map[string]interface{}, 0, len(response.Data))
	for _, item := range response.Data {
		filteredItem := filterResourceIdAndReorder(item)
		filteredData = append(filteredData, filteredItem)
	}

	// Output the filtered data as JSON
	jsonOutput, err := json.MarshalIndent(filteredData, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
