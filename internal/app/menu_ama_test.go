package app

import (
	"testing"
	"time"
)

func TestExtractDateFromISO(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid RFC3339 format",
			input:    "2025-12-06T15:30:00Z",
			expected: "2025-12-06",
		},
		{
			name:     "valid RFC3339 with timezone offset",
			input:    "2025-01-15T10:00:00+02:00",
			expected: "2025-01-15",
		},
		{
			name:     "valid RFC3339 with milliseconds",
			input:    "2024-06-20T08:45:30.123Z",
			expected: "2024-06-20",
		},
		{
			name:     "date-only string (fallback extraction)",
			input:    "2025-03-10",
			expected: "2025-03-10",
		},
		{
			name:     "empty string returns today",
			input:    "",
			expected: time.Now().UTC().Format("2006-01-02"),
		},
		{
			name:     "invalid format returns today",
			input:    "not-a-date",
			expected: time.Now().UTC().Format("2006-01-02"),
		},
		{
			name:     "partial date returns today",
			input:    "2025",
			expected: time.Now().UTC().Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDateFromISO(tt.input)
			if result != tt.expected {
				t.Errorf("extractDateFromISO(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
