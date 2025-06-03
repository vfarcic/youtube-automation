package utils

import (
	"testing"
	"time"
)

func TestIsFarFutureDate(t *testing.T) {
	// Use a fixed time in UTC to make the test deterministic across timezones
	now := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	layout := "2006-01-02T15:04"

	tests := []struct {
		name      string
		dateStr   string
		want      bool
		expectErr bool
	}{
		{
			name:    "Date exactly 3 months from now",
			dateStr: now.AddDate(0, 3, 0).Format(layout),
			want:    false, // Exactly 3 months is not > 3 months
		},
		{
			name:    "Date slightly less than 3 months from now (e.g., 3 months minus 1 day)",
			dateStr: now.AddDate(0, 3, -1).Format(layout),
			want:    false,
		},
		{
			name:    "Date slightly more than 3 months from now (e.g., 3 months plus 1 day)",
			dateStr: now.AddDate(0, 3, 1).Format(layout),
			want:    true,
		},
		{
			name:    "Date 4 months from now",
			dateStr: now.AddDate(0, 4, 0).Format(layout),
			want:    true,
		},
		{
			name:    "Date in the past",
			dateStr: now.AddDate(0, -1, 0).Format(layout), // 1 month ago
			want:    false,
		},
		{
			name:      "Empty date string",
			dateStr:   "",
			want:      false,
			expectErr: true, // Expecting an error for empty string
		},
		{
			name:      "Invalid date format",
			dateStr:   "2023-12-32T10:00", // Invalid day
			want:      false,
			expectErr: true, // Expecting an error for invalid format
		},
		{
			name:    "Date 2 months from now",
			dateStr: now.AddDate(0, 2, 0).Format(layout),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pass the fixed UTC time to make the test deterministic.
			got, err := IsFarFutureDate(tt.dateStr, layout, now)
			if (err != nil) != tt.expectErr {
				t.Errorf("IsFarFutureDate() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && got != tt.want {
				t.Errorf("IsFarFutureDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
