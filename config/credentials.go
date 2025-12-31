package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Credentials holds API key (workspace ID is not stored in credentials file)
type Credentials struct {
	APIKey string
}

// ReadCredentials reads API key from ~/.jsmon/credentials file
// Expected format: API key written directly in the file (first non-empty, non-comment line)
// Example:
// your-api-key-here
func ReadCredentials() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	credPath := filepath.Join(homeDir, ".jsmon", "credentials")

	file, err := os.Open(credPath)
	if err != nil {
		// File doesn't exist, return empty string (not an error)
		return "", nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Return the first non-empty, non-comment line as the API key
		return line, nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading credentials file: %w", err)
	}

	return "", nil
}
