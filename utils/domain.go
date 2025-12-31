package utils

import (
	"net/url"
	"strings"
)

// ExtractDomain extracts the domain from a URL or returns the input if it's already a domain
func ExtractDomain(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}

	// If it doesn't start with http:// or https://, assume it's already a domain
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		// Remove any trailing slashes
		return strings.TrimSuffix(input, "/")
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		// If parsing fails, try to extract domain manually
		// Remove http:// or https://
		domain := strings.TrimPrefix(input, "http://")
		domain = strings.TrimPrefix(domain, "https://")
		// Remove path and query
		if idx := strings.Index(domain, "/"); idx >= 0 {
			domain = domain[:idx]
		}
		if idx := strings.Index(domain, "?"); idx >= 0 {
			domain = domain[:idx]
		}
		// Remove port
		if idx := strings.Index(domain, ":"); idx >= 0 {
			domain = domain[:idx]
		}
		return strings.TrimSpace(domain)
	}

	// Extract hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		// Fallback: use the path if hostname is empty
		hostname = strings.TrimPrefix(parsedURL.Path, "/")
		if idx := strings.Index(hostname, "/"); idx >= 0 {
			hostname = hostname[:idx]
		}
	}

	return strings.TrimSpace(hostname)
}

