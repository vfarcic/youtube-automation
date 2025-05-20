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

	// If the parsed date is not even after the reference time, it can't be far future.
	// This initial check considers the full precision of referenceTime and parsedDate (from layout).
	if !parsedDate.After(referenceTime) {
		return false, nil
	}

	// Determine the start of the day for referenceTime.
	year, month, day := referenceTime.Date()
	referenceDayStart := time.Date(year, month, day, 0, 0, 0, 0, referenceTime.Location())

	// Calculate the threshold: the start of the day that is (3 months + 1 day) after referenceDayStart.
	// A date is "far future" if it is on or after this threshold day.
	farFutureThresholdDayStart := referenceDayStart.AddDate(0, 3, 1)

	// Truncate the parsed date to the beginning of its day.
	parsedDateDayStart := parsedDate.Truncate(24 * time.Hour)

	// Check if parsedDateDayStart is on or after farFutureThresholdDayStart.
	return !parsedDateDayStart.Before(farFutureThresholdDayStart), nil
}
