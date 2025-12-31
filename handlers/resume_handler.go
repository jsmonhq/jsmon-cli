package handlers

import (
	"fmt"
	"os"

	"github.com/jsmonhq/jsmon-cli/resume"
)

// HandleResume handles resuming from a resume file
// Note: API key and workspace ID must be provided via flags/env vars, not from resume file
func HandleResume(resumeFile string, workspaceID, apiKey string, headers map[string]string) {
	config, err := resume.Load(resumeFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError loading resume file: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	// Only support file type for resume
	if config.Type != "file" {
		fmt.Fprintf(os.Stderr, "%sError: Resume file type must be 'file'. Only file scanning supports resume functionality.%s\n", ColorRed, ColorReset)
		os.Exit(1)
	}

	// Use API key and workspace ID from command line/env vars (not from resume file)
	HandleFileUpload(config.File, workspaceID, apiKey, resumeFile, headers)
}
