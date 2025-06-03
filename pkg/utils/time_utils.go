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

	// Convert both times to UTC to ensure consistent timezone handling
	referenceUTC := referenceTime.UTC()
	parsedUTC := parsedDate.UTC()

	// If the parsed date is not even after the reference time, it can't be far future.
	if !parsedUTC.After(referenceUTC) {
		return false, nil
	}

	// Get the date components in UTC to avoid timezone issues
	refYear, refMonth, refDay := referenceUTC.Date()
	parsedYear, parsedMonth, parsedDay := parsedUTC.Date()

	// Create start of day in UTC for both dates
	referenceDayStart := time.Date(refYear, refMonth, refDay, 0, 0, 0, 0, time.UTC)
	parsedDateDayStart := time.Date(parsedYear, parsedMonth, parsedDay, 0, 0, 0, 0, time.UTC)

	// Calculate the threshold: the start of the day that is (3 months + 1 day) after referenceDayStart.
	// A date is "far future" if it is on or after this threshold day.
	farFutureThresholdDayStart := referenceDayStart.AddDate(0, 3, 1)

	// Check if parsedDateDayStart is on or after farFutureThresholdDayStart.
	return !parsedDateDayStart.Before(farFutureThresholdDayStart), nil
}
