package app

import (
	"fmt"
	"math/rand"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"
)

// GetWeekBoundaries returns the Monday (start) and Sunday (end) of the week containing the given date.
// Week is defined as Monday-Sunday.
// Monday is set to 00:00:00, Sunday is set to 23:59:59.
func GetWeekBoundaries(date time.Time) (monday, sunday time.Time) {
	// Normalize to start of day
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	// Calculate days since Monday
	// In Go, Sunday = 0, Monday = 1, ..., Saturday = 6
	// We want Monday = 0, Tuesday = 1, ..., Sunday = 6
	weekday := int(date.Weekday())
	daysSinceMonday := (weekday + 6) % 7 // Convert Sunday=0 to Sunday=6

	// Calculate Monday (subtract days since Monday)
	monday = date.AddDate(0, 0, -daysSinceMonday)

	// Calculate Sunday (add remaining days to get to Sunday)
	daysUntilSunday := 6 - daysSinceMonday
	sunday = date.AddDate(0, 0, daysUntilSunday)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())

	return monday, sunday
}

// ApplyRandomTiming picks a random timing recommendation and calculates the new publish date
// within the same week (Monday-Sunday) as the current date.
//
// Parameters:
//   - currentDateStr: Current video date in format "YYYY-MM-DDTHH:MM" (e.g., "2025-12-02T16:00")
//   - recommendations: List of timing recommendations from settings.yaml
//
// Returns:
//   - newDateStr: New date string in same format
//   - selectedRec: The randomly selected recommendation (for showing reasoning to user)
//   - error: If parsing fails or no recommendations available
func ApplyRandomTiming(currentDateStr string, recommendations []configuration.TimingRecommendation) (string, configuration.TimingRecommendation, error) {
	// Validate input
	if len(recommendations) == 0 {
		return "", configuration.TimingRecommendation{}, fmt.Errorf("no timing recommendations available in settings.yaml")
	}

	// Parse current date
	currentDate, err := time.Parse("2006-01-02T15:04", currentDateStr)
	if err != nil {
		return "", configuration.TimingRecommendation{}, fmt.Errorf("invalid date format '%s': %w", currentDateStr, err)
	}

	// Pick random recommendation
	selectedRec := recommendations[rand.Intn(len(recommendations))]

	// Parse target time from recommendation (e.g., "16:00")
	targetTime, err := time.Parse("15:04", selectedRec.Time)
	if err != nil {
		return "", configuration.TimingRecommendation{}, fmt.Errorf("invalid time format in recommendation '%s': %w", selectedRec.Time, err)
	}

	// Calculate new date within same week
	newDate, err := calculateDateInSameWeek(currentDate, selectedRec.Day, targetTime.Hour(), targetTime.Minute())
	if err != nil {
		return "", configuration.TimingRecommendation{}, err
	}

	// Format back to YYYY-MM-DDTHH:MM
	newDateStr := newDate.Format("2006-01-02T15:04")

	return newDateStr, selectedRec, nil
}

// calculateDateInSameWeek finds the target day/time within the same week as the current date.
// Week is defined as Monday-Sunday.
func calculateDateInSameWeek(currentDate time.Time, targetDay string, targetHour, targetMinute int) (time.Time, error) {
	// Get week boundaries
	monday, sunday := GetWeekBoundaries(currentDate)

	// Map day name to weekday offset from Monday
	dayOffsets := map[string]int{
		"Monday":    0,
		"Tuesday":   1,
		"Wednesday": 2,
		"Thursday":  3,
		"Friday":    4,
		"Saturday":  5,
		"Sunday":    6,
	}

	offset, ok := dayOffsets[targetDay]
	if !ok {
		return time.Time{}, fmt.Errorf("invalid day name '%s'", targetDay)
	}

	// Calculate target date
	targetDate := monday.AddDate(0, 0, offset)
	targetDate = time.Date(
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		targetHour,
		targetMinute,
		0,
		0,
		currentDate.Location(),
	)

	// Validate target date is within week boundaries
	if targetDate.Before(monday) || targetDate.After(sunday) {
		return time.Time{}, fmt.Errorf("calculated date %s is outside week boundaries (%s to %s)",
			targetDate.Format("2006-01-02"), monday.Format("2006-01-02"), sunday.Format("2006-01-02"))
	}

	return targetDate, nil
}
