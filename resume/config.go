package resume

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the resume configuration
type Config struct {
	Type         string   `json:"type"`         // "file" only
	File         string   `json:"file,omitempty"`           // For file scanning
	ProcessedURLs []string `json:"processed_urls,omitempty"` // URLs already processed
	TotalURLs    []string `json:"total_urls,omitempty"`    // All URLs to process
	LastIndex    int      `json:"last_index"`               // Last processed index
}

// Save saves the resume configuration to a file
func Save(config *Config, filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load loads the resume configuration from a file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// Delete removes the resume configuration file
func Delete(filename string) error {
	return os.Remove(filename)
}

