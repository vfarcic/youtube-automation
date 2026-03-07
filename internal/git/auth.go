package git

import (
	"net/url"
	"strings"
)

// AuthenticatedURL injects a token into an HTTPS URL using x-access-token scheme.
// If token is empty or the URL is invalid, the original URL is returned unchanged.
func AuthenticatedURL(rawURL, token string) string {
	if token == "" {
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	parsed.User = url.UserPassword("x-access-token", token)
	return parsed.String()
}

// SanitizeOutput removes tokens from command output to prevent leaking secrets in logs.
func SanitizeOutput(output []byte, token string) string {
	s := string(output)
	if token != "" {
		s = strings.ReplaceAll(s, token, "***")
	}
	return s
}
