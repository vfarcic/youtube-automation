package publishing

import (
	"math/rand"
	"time"
)

// CalculateShortsSchedule returns scheduled publish times for shorts.
// Starting from the day after the main video's publish date, each short
// is scheduled 1 day apart with randomized hour (0-23) and minute (0-59).
//
// Parameters:
//   - mainVideoDate: The main video's scheduled publish date
//   - count: Number of shorts to schedule
//
// Returns:
//   - []time.Time: Slice of scheduled times, one per short
func CalculateShortsSchedule(mainVideoDate time.Time, count int) []time.Time {
	if count <= 0 {
		return []time.Time{}
	}

	schedules := make([]time.Time, count)

	for i := 0; i < count; i++ {
		// Start from day after main video, then add i days
		daysAfter := i + 1
		scheduledDate := mainVideoDate.AddDate(0, 0, daysAfter)

		// Randomize hour (0-23) and minute (0-59)
		randomHour := rand.Intn(24)
		randomMinute := rand.Intn(60)

		// Create the scheduled time with randomized hour/minute, keeping the date
		schedules[i] = time.Date(
			scheduledDate.Year(),
			scheduledDate.Month(),
			scheduledDate.Day(),
			randomHour,
			randomMinute,
			0, // seconds
			0, // nanoseconds
			scheduledDate.Location(),
		)
	}

	return schedules
}

// FormatScheduleISO formats a time as ISO 8601 string for YouTube API
// Format: 2006-01-02T15:04:00Z
func FormatScheduleISO(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:00Z")
}
