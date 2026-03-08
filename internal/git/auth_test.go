package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticatedURL(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		token    string
		expected string
	}{
		{
			name:     "with token",
			rawURL:   "https://github.com/user/repo.git",
			token:    "ghp_abc123",
			expected: "https://x-access-token:ghp_abc123@github.com/user/repo.git",
		},
		{
			name:     "without token",
			rawURL:   "https://github.com/user/repo.git",
			token:    "",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "invalid URL falls back",
			rawURL:   "://invalid",
			token:    "tok",
			expected: "://invalid",
		},
		{
			name:     "preserves path and query",
			rawURL:   "https://github.com/org/repo.git?foo=bar",
			token:    "mytoken",
			expected: "https://x-access-token:mytoken@github.com/org/repo.git?foo=bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AuthenticatedURL(tt.rawURL, tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}
