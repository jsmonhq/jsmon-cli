package handlers

import (
	"fmt"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
)

func printScanQueued(label, target string, response *api.ScanSubmitResponse) {
	fmt.Printf("%s%s queued for - %s%s\n", ColorGreen, label, target, ColorReset)
	if response == nil {
		return
	}

	if strings.TrimSpace(response.RunID) != "" {
		fmt.Printf("Run ID: %s\n", response.RunID)
	}
	if response.Version > 0 {
		fmt.Printf("Version: %d\n", response.Version)
	}
	if strings.TrimSpace(response.Status) != "" {
		fmt.Printf("Status: %s\n", response.Status)
	}
}
