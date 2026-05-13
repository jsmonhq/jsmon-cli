package handlers

import (
	"fmt"
	"os"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleCreateWorkspace creates a new workspace
func HandleCreateWorkspace(workspaceName, apiKey string, headers map[string]string) {
	workspaceName = strings.TrimSpace(workspaceName)
	if workspaceName == "" {
		fmt.Fprintf(os.Stderr, "%sError: Workspace name cannot be empty%s\n", ColorRed, ColorReset)
		os.Exit(1)
	}

	client := api.NewClient(apiKey, headers)

	workspaceID, err := client.CreateWorkspace(workspaceName)
	if err != nil {
		// Check for authentication error (wrong or missing API key)
		if apiErr, ok := err.(*api.APIError); ok && apiErr.IsAuthError() {
			fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error creating workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s✓ Workspace created successfully%s\n", ColorGreen, ColorReset)
	fmt.Printf("Workspace Name: %s\n", workspaceName)
	fmt.Printf("Workspace ID: %s\n", workspaceID)
}
