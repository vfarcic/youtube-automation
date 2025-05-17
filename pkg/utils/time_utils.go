package utils

import (
	"fmt"
	"time"
)

// IsFarFutureDate checks if the given dateStr is more than 3 calendar months after the referenceTime.
// dateStr: The date string to check.
// layout: The layout of the dateStr (e.g., "2006-01-02T15:04").
// referenceTime: The time to compare against (typically time.Now()).
func IsFarFutureDate(dateStr string, layout string, referenceTime time.Time) (bool, error) {
	if dateStr == "" {
		return false, fmt.Errorf("date string cannot be empty")
	}

	parsedDate, err := time.Parse(layout, dateStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse date string '%s' with layout '%s': %w", dateStr, layout, err)
	}

	// If the parsed date is before or exactly at the reference time, it's not in the future at all.
	if !parsedDate.After(referenceTime) {
		return false, nil
	}

	// Calculate the point in time that is exactly 3 calendar months from the referenceTime.
	threeMonthsFromReference := referenceTime.AddDate(0, 3, 0)

	// If the parsed date is after this point, then it's considered "far future".
	return parsedDate.After(threeMonthsFromReference), nil
}
