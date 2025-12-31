package handlers

import (
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleWorkspaces displays all workspaces for the user
func HandleWorkspaces(apiKey string, headers map[string]string) {
	client := api.NewClient(apiKey, headers)

	response, err := client.GetWorkspaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError fetching workspaces: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Display workspaces
	if len(response.Workspaces) == 0 {
		fmt.Printf("\n%sNo workspaces found.%s\n", ColorGreen, ColorReset)
		return
	}

	// Find the longest workspace name for alignment
	maxNameLen := 0
	for _, workspace := range response.Workspaces {
		if len(workspace.Name) > maxNameLen {
			maxNameLen = len(workspace.Name)
		}
	}
	// Cap at 40 characters for readability
	if maxNameLen > 30 {
		maxNameLen = 30
	}

	fmt.Println()
	for _, workspace := range response.Workspaces {
		sharedStatus := "No"
		if workspace.IsShared {
			sharedStatus = "Yes"
		}
		// Format with aligned columns
		fmt.Printf("  %-*s  (ID: %s)  -  Shared Workspace: %s\n", maxNameLen, workspace.Name, workspace.WkspID, sharedStatus)
	}
	fmt.Println()
}
