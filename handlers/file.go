package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jsmonhq/jsmon-cli/api"
	"github.com/jsmonhq/jsmon-cli/resume"
)

// FileScanResult represents the result of a file scan
type FileScanResult struct {
	Success []string                 `json:"success"`
	Failed  []map[string]string      `json:"failed"`
	Summary map[string]interface{}   `json:"summary"`
}

// HandleFileUpload uploads a file for scanning
// The file should contain one JS URL per line
func HandleFileUpload(filePath, workspaceID, apiKey string, resumeFile string, headers map[string]string) {
	var resumeState *ResumeState
	var startIndex int
	var urls []string

	// Check if resuming
	if resumeFile != "" {
		config, err := resume.Load(resumeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError loading resume file: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}
		if config.Type != "file" || config.File != filePath {
			fmt.Fprintf(os.Stderr, "%sError: Resume file does not match current file%s\n", ColorRed, ColorReset)
			os.Exit(1)
		}
		startIndex = config.LastIndex + 1
		urls = config.TotalURLs

		// Create resume state from loaded config to enable updates during scan
		resumeState = &ResumeState{
			Config:   config,
			Filename: resumeFile,
			Saved:    false,
		}
		SetupSignalHandler(resumeState)
		fmt.Printf("Resuming from URL %d/%d\n", startIndex+1, len(urls))
	} else {
		// Read all URLs from file
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError opening file: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			url := strings.TrimSpace(scanner.Text())
			if url != "" {
				urls = append(urls, url)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "%sError reading file: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}

		// Use standard resume filename
		resumeFilename := "resume.cfg"

		// Create resume state (API key and workspace ID are NOT stored in resume file)
		resumeState = NewResumeState(resumeFilename, "file", workspaceID, apiKey)
		resumeState.Config.File = filePath
		resumeState.Config.TotalURLs = urls
		SetupSignalHandler(resumeState)
		startIndex = 0
	}

	fmt.Printf("Scanning started for - %s\n", filePath)

	client := api.NewClient(apiKey, headers)
	successCount := 0
	errorCount := 0
	lastProcessedURL := ""
	
	// Collect results for JSON output
	var successURLs []string
	var failedURLs []map[string]string

	for i := startIndex; i < len(urls); i++ {
		jsURL := urls[i]
		lastProcessedURL = jsURL

		// Update resume state
		if resumeState != nil {
			resumeState.Config.LastIndex = i
			resumeState.Config.ProcessedURLs = append(resumeState.Config.ProcessedURLs, jsURL)
		}

		// Upload URL
		err := client.UploadURL(jsURL, workspaceID)
		if err != nil {
			// Check if it's an APIError
			if apiErr, ok := err.(*api.APIError); ok {
				// Check for authentication error (wrong or missing API key)
				if apiErr.IsAuthError() {
					fmt.Fprintf(os.Stderr, "Error: API key is invalid or not configured. Use -key flag, add to ~/.jsmon/credentials, or set JSMON_API_KEY environment variable\n")
					os.Exit(1)
				}

				// Check if it's an insufficient limits error
				if apiErr.IsInsufficientLimitsError() {
					fmt.Fprintf(os.Stderr, "\n%s[!] Insufficient scan limits. Saving resume state...%s\n", ColorRed, ColorReset)
					if resumeState != nil {
						if saveErr := resumeState.Save(); saveErr != nil {
							fmt.Fprintf(os.Stderr, "%s[!] Failed to save resume state: %v%s\n", ColorRed, saveErr, ColorReset)
						} else {
							fmt.Fprintf(os.Stderr, "%s[!] Resume state saved to: %s%s\n", ColorGreen, resumeState.Filename, ColorReset)
							fmt.Fprintf(os.Stderr, "%s[!] Please add scan limits and resume with: -resume %s%s\n", ColorGreen, resumeState.Filename, ColorReset)
						}
					} else {
						fmt.Fprintf(os.Stderr, "%s[!] Insufficient scan limits%s\n", ColorRed, ColorReset)
						fmt.Fprintf(os.Stderr, "%s[!] Please add scan limits and try again%s\n", ColorRed, ColorReset)
					}
					os.Exit(1)
				}

				// Check if it's a rate limit error
				if apiErr.IsRateLimitError() {
					fmt.Fprintf(os.Stderr, "\n%s[!] Rate limit reached. Saving resume state...%s\n", ColorRed, ColorReset)
					if resumeState != nil {
						if saveErr := resumeState.Save(); saveErr != nil {
							fmt.Fprintf(os.Stderr, "%s[!] Failed to save resume state: %v%s\n", ColorRed, saveErr, ColorReset)
						} else {
							fmt.Fprintf(os.Stderr, "%s[!] Resume state saved to: %s%s\n", ColorGreen, resumeState.Filename, ColorReset)
							fmt.Fprintf(os.Stderr, "%s[!] Resume with: -resume %s%s\n", ColorGreen, resumeState.Filename, ColorReset)
						}
					}
					os.Exit(1)
				}
			}

			// Regular error handling
			// Save resume state on any error (but continue processing other URLs)
			if resumeState != nil {
				// Save progress before showing error
				resumeState.Save()
			}

			var errorMsg string
			if apiErr, ok := err.(*api.APIError); ok {
				reason := strings.TrimSpace(apiErr.Message)
				// Clean up the message - remove "URL Scan failed - " prefix if present
				if strings.HasPrefix(reason, "URL Scan failed - ") {
					reason = strings.TrimPrefix(reason, "URL Scan failed - ")
					reason = strings.TrimSpace(reason)
				}
				// Also check for old "JS Scan failed - " format for backward compatibility
				if strings.HasPrefix(reason, "JS Scan failed - ") {
					reason = strings.TrimPrefix(reason, "JS Scan failed - ")
					reason = strings.TrimSpace(reason)
				}
				// Remove "Please provide a valid URL." from the end if present
				reason = strings.TrimSuffix(reason, " Please provide a valid URL.")
				reason = strings.TrimSuffix(reason, "Please provide a valid URL.")
				// Also check for old "JS file URL" format for backward compatibility
				reason = strings.TrimSuffix(reason, " Please provide a valid JS file URL.")
				reason = strings.TrimSuffix(reason, "Please provide a valid JS file URL.")
				reason = strings.TrimSpace(reason)
				errorMsg = reason
				fmt.Fprintf(os.Stderr, "%sError uploading url: %s - %s%s\n", ColorRed, apiErr.URL, reason, ColorReset)
			} else {
				errorMsg = err.Error()
				fmt.Fprintf(os.Stderr, "%sError uploading url: %s - %v%s\n", ColorRed, jsURL, err, ColorReset)
			}
			errorCount++
			failedURLs = append(failedURLs, map[string]string{
				"url":   jsURL,
				"error": errorMsg,
			})
		} else {
			fmt.Fprintf(os.Stderr, "%sScan Completed: %s%s\n", ColorGreen, jsURL, ColorReset)
			successCount++
			successURLs = append(successURLs, jsURL)
		}

		// Save progress periodically (every 10 URLs)
		if resumeState != nil && (i+1)%10 == 0 {
			resumeState.Save()
		}
	}

	// Save final state and clean up resume file if completed (only when not resuming)
	if resumeState != nil {
		if resumeFile == "" {
			// Save final state before deleting (for non-resume scans)
			resumeState.Save()
		resume.Delete(resumeState.Filename)
		} else {
			// Save final state when resuming (keep the file)
			resumeState.Save()
		}
	}

	// Output results as JSON
	result := FileScanResult{
		Success: successURLs,
		Failed:  failedURLs,
		Summary: map[string]interface{}{
			"total":   len(urls),
			"success": successCount,
			"failed":  errorCount,
			"lastUrl": lastProcessedURL,
		},
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError marshaling JSON: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}
