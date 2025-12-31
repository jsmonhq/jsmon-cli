package handlers

import (
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/api"
)

// HandleCreateWorkspace creates a new workspace
func HandleCreateWorkspace(workspaceName, apiKey string, headers map[string]string) {
	client := api.NewClient(apiKey, headers)

	workspaceID, err := client.CreateWorkspace(workspaceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating workspace: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s✓ Workspace created successfully%s\n", ColorGreen, ColorReset)
	fmt.Printf("Workspace Name: %s\n", workspaceName)
	fmt.Printf("Workspace ID: %s\n", workspaceID)
}
