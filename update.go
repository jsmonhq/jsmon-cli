package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// Version is the current version of the CLI
	Version = "1.0.0"
	// GitHubRepo is the repository for checking updates
	GitHubRepo = "jsmonhq/jsmon-cli"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
}

// checkForUpdate checks if a newer version is available on GitHub
func checkForUpdate() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Fetch latest release from GitHub API
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return // Silently fail if we can't create request
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return // Silently fail if network request fails
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return // Silently fail if API returns error
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return // Silently fail if we can't read response
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return // Silently fail if we can't parse JSON
	}

	// Compare versions (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Version, "v")

	// Simple version comparison (works for semantic versioning)
	if compareVersions(latestVersion, currentVersion) > 0 {
		fmt.Fprintf(os.Stderr, "\n[INF] A new version is available: %s (current: %s)\n", latestVersion, currentVersion)
		fmt.Fprintf(os.Stderr, "[INF] Update with: go install github.com/jsmonhq/jsmon-cli@latest\n\n")
	}
}

// compareVersions compares two version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var v1Part, v2Part int
		if i < len(v1Parts) {
			fmt.Sscanf(v1Parts[i], "%d", &v1Part)
		}
		if i < len(v2Parts) {
			fmt.Sscanf(v2Parts[i], "%d", &v2Part)
		}

		if v1Part > v2Part {
			return 1
		}
		if v1Part < v2Part {
			return -1
		}
	}

	return 0
}
