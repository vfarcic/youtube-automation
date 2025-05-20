package utils

import (
	"fmt"
	"time"
)

// IsFarFutureDate checks if the given dateStr is on a day that is more than 3 calendar months
// after the day of the referenceTime.
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
	// This check uses the original precision of referenceTime and the HH:MM:00 precision of parsedDate.
	if !parsedDate.After(referenceTime) {
		return false, nil
	}

	// Calculate the boundary: exactly 3 calendar months from the referenceTime.
	threeMonthsBoundaryDate := referenceTime.AddDate(0, 3, 0)

	// Truncate both the parsed date and the boundary date to the beginning of their respective days.
	// This makes the comparison based on whole days, aligning with "calendar months".
	parsedDateDayStart := parsedDate.Truncate(24 * time.Hour)
	threeMonthsBoundaryDayStart := threeMonthsBoundaryDate.Truncate(24 * time.Hour)

	// If the day of the parsed date is strictly after the day of the 3-month boundary,
	// then it's considered "far future".
	return parsedDateDayStart.After(threeMonthsBoundaryDayStart), nil
}
