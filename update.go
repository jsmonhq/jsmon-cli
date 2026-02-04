package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// Version is the current version of the CLI (must match the latest release tag)
	Version = "2.0.0"
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
	}
}

// checkAndUpdateCLI runs on -up: shows current version, fetches latest from GitHub, runs go install if newer. Minimal logs.
func checkAndUpdateCLI() {
	currentVersion := strings.TrimPrefix(Version, "v")
	fmt.Fprintf(os.Stderr, "Current version: %s\n", currentVersion)

	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GitHubRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Fprintf(os.Stderr, "No release found.\n")
		os.Exit(0)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: GitHub API %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	if compareVersions(latestVersion, currentVersion) <= 0 {
		fmt.Fprintf(os.Stderr, "Already on latest version.\n")
		os.Exit(0)
	}

	cmd := exec.Command("go", "install", "github.com/jsmonhq/jsmon-cli@latest")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}

	updateVersionInSource(latestVersion)
	fmt.Fprintf(os.Stderr, "Updated to %s.\n", latestVersion)
	os.Exit(0)
}

// updateVersionInSource sets Version = newVer in update.go in the current directory (when running from repo).
// Returns true if the file was updated.
func updateVersionInSource(newVer string) bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	path := filepath.Join(cwd, "update.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	re := regexp.MustCompile(`(\tVersion\s*=\s*)"[^"]*"`)
	newLine := re.ReplaceAllString(string(data), "${1}\""+newVer+"\"")
	if newLine == string(data) {
		return false
	}
	if err := os.WriteFile(path, []byte(newLine), 0644); err != nil {
		return false
	}
	return true
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
