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
// Supported formats:
//   API_KEY='your-api-key-here'
//   API_KEY="your-api-key-here"
//   API_KEY=your-api-key-here
// or, for backwards compatibility, the API key written directly on its own
// line (the first non-empty, non-comment line):
//   your-api-key-here
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
	fallback := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if apiKey, ok := parseAPIKeyLine(line); ok {
			return apiKey, nil
		}

		// Keep the first non-empty, non-comment line as a fallback so the
		// old plain-key format keeps working
		if fallback == "" {
			fallback = line
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading credentials file: %w", err)
	}

	return fallback, nil
}

// parseAPIKeyLine recognizes an "API_KEY=<value>" line, with the value
// optionally wrapped in single or double quotes.
func parseAPIKeyLine(line string) (string, bool) {
	const prefix = "API_KEY="
	if len(line) < len(prefix) || !strings.EqualFold(line[:len(prefix)], prefix) {
		return "", false
	}

	value := strings.TrimSpace(line[len(prefix):])
	value = strings.Trim(value, `'"`)
	return value, true
}
