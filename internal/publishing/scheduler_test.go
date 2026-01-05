package publishing

import (
	"testing"
	"time"
)

func TestCalculateShortsSchedule(t *testing.T) {
	// Fixed reference date for testing
	mainVideoDate := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		mainVideoDate time.Time
		count         int
		wantLen       int
	}{
		{
			name:          "schedule 3 shorts",
			mainVideoDate: mainVideoDate,
			count:         3,
			wantLen:       3,
		},
		{
			name:          "schedule 1 short",
			mainVideoDate: mainVideoDate,
			count:         1,
			wantLen:       1,
		},
		{
			name:          "schedule 5 shorts",
			mainVideoDate: mainVideoDate,
			count:         5,
			wantLen:       5,
		},
		{
			name:          "zero count returns empty",
			mainVideoDate: mainVideoDate,
			count:         0,
			wantLen:       0,
		},
		{
			name:          "negative count returns empty",
			mainVideoDate: mainVideoDate,
			count:         -1,
			wantLen:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateShortsSchedule(tt.mainVideoDate, tt.count)
			if len(got) != tt.wantLen {
				t.Errorf("CalculateShortsSchedule() returned %d schedules, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestCalculateShortsSchedule_DayIntervals(t *testing.T) {
	mainVideoDate := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	count := 3

	schedules := CalculateShortsSchedule(mainVideoDate, count)

	// Check that each short is scheduled on the correct day
	for i, schedule := range schedules {
		expectedDay := mainVideoDate.AddDate(0, 0, i+1).Day()
		expectedMonth := mainVideoDate.AddDate(0, 0, i+1).Month()
		expectedYear := mainVideoDate.AddDate(0, 0, i+1).Year()

		if schedule.Day() != expectedDay || schedule.Month() != expectedMonth || schedule.Year() != expectedYear {
			t.Errorf("Short %d scheduled on %v, expected day %d/%d/%d",
				i+1, schedule, expectedYear, expectedMonth, expectedDay)
		}
	}
}

func TestCalculateShortsSchedule_RandomizedTimes(t *testing.T) {
	mainVideoDate := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	count := 10

	schedules := CalculateShortsSchedule(mainVideoDate, count)

	// Check that hours and minutes are within valid ranges
	for i, schedule := range schedules {
		if schedule.Hour() < 0 || schedule.Hour() > 23 {
			t.Errorf("Short %d has invalid hour: %d", i+1, schedule.Hour())
		}
		if schedule.Minute() < 0 || schedule.Minute() > 59 {
			t.Errorf("Short %d has invalid minute: %d", i+1, schedule.Minute())
		}
		// Seconds should always be 0
		if schedule.Second() != 0 {
			t.Errorf("Short %d has non-zero seconds: %d", i+1, schedule.Second())
		}
	}
}

func TestCalculateShortsSchedule_PreservesTimezone(t *testing.T) {
	// Test with a specific timezone
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("Could not load America/New_York timezone")
	}

	mainVideoDate := time.Date(2025, 1, 15, 10, 0, 0, 0, loc)
	schedules := CalculateShortsSchedule(mainVideoDate, 2)

	for i, schedule := range schedules {
		if schedule.Location() != loc {
			t.Errorf("Short %d timezone mismatch: got %v, want %v", i+1, schedule.Location(), loc)
		}
	}
}

func TestCalculateShortsSchedule_MonthBoundary(t *testing.T) {
	// Test scheduling across month boundary
	mainVideoDate := time.Date(2025, 1, 30, 10, 0, 0, 0, time.UTC)
	schedules := CalculateShortsSchedule(mainVideoDate, 3)

	// Short 1: Jan 31
	if schedules[0].Month() != time.January || schedules[0].Day() != 31 {
		t.Errorf("Short 1 expected Jan 31, got %v", schedules[0])
	}

	// Short 2: Feb 1
	if schedules[1].Month() != time.February || schedules[1].Day() != 1 {
		t.Errorf("Short 2 expected Feb 1, got %v", schedules[1])
	}

	// Short 3: Feb 2
	if schedules[2].Month() != time.February || schedules[2].Day() != 2 {
		t.Errorf("Short 3 expected Feb 2, got %v", schedules[2])
	}
}

func TestFormatScheduleISO(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "UTC time",
			time: time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
			want: "2025-01-15T14:30:00Z",
		},
		{
			name: "midnight",
			time: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			want: "2025-01-15T00:00:00Z",
		},
		{
			name: "end of day",
			time: time.Date(2025, 1, 15, 23, 59, 0, 0, time.UTC),
			want: "2025-01-15T23:59:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatScheduleISO(tt.time)
			if got != tt.want {
				t.Errorf("FormatScheduleISO() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatScheduleISO_ConvertsToUTC(t *testing.T) {
	// Test that non-UTC times are converted to UTC
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("Could not load America/New_York timezone")
	}

	// 10:00 AM EST = 15:00 UTC (during standard time)
	estTime := time.Date(2025, 1, 15, 10, 0, 0, 0, loc)
	result := FormatScheduleISO(estTime)

	// Should be converted to UTC
	if result != "2025-01-15T15:00:00Z" {
		t.Errorf("FormatScheduleISO() = %v, expected UTC conversion to 2025-01-15T15:00:00Z", result)
	}
}
