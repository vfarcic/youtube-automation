package app

import (
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"
)

func TestGetWeekBoundaries(t *testing.T) {
	tests := []struct {
		name          string
		inputDate     string
		wantMonday    string
		wantSunday    string
		wantMondayTime string
		wantSundayTime string
	}{
		{
			name:           "Monday of week",
			inputDate:      "2025-12-01", // Monday
			wantMonday:     "2025-12-01",
			wantSunday:     "2025-12-07",
			wantMondayTime: "00:00:00",
			wantSundayTime: "23:59:59",
		},
		{
			name:           "Tuesday of week",
			inputDate:      "2025-12-02", // Tuesday
			wantMonday:     "2025-12-01",
			wantSunday:     "2025-12-07",
			wantMondayTime: "00:00:00",
			wantSundayTime: "23:59:59",
		},
		{
			name:           "Thursday of week",
			inputDate:      "2025-12-04", // Thursday
			wantMonday:     "2025-12-01",
			wantSunday:     "2025-12-07",
			wantMondayTime: "00:00:00",
			wantSundayTime: "23:59:59",
		},
		{
			name:           "Sunday of week",
			inputDate:      "2025-12-07", // Sunday
			wantMonday:     "2025-12-01",
			wantSunday:     "2025-12-07",
			wantMondayTime: "00:00:00",
			wantSundayTime: "23:59:59",
		},
		{
			name:           "Saturday of week",
			inputDate:      "2025-12-06", // Saturday
			wantMonday:     "2025-12-01",
			wantSunday:     "2025-12-07",
			wantMondayTime: "00:00:00",
			wantSundayTime: "23:59:59",
		},
		{
			name:           "Different week - Monday",
			inputDate:      "2025-12-08", // Monday of next week
			wantMonday:     "2025-12-08",
			wantSunday:     "2025-12-14",
			wantMondayTime: "00:00:00",
			wantSundayTime: "23:59:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputDate, err := time.Parse("2006-01-02", tt.inputDate)
			if err != nil {
				t.Fatalf("Failed to parse input date: %v", err)
			}

			gotMonday, gotSunday := GetWeekBoundaries(inputDate)

			// Check Monday date
			if gotMonday.Format("2006-01-02") != tt.wantMonday {
				t.Errorf("Monday date = %s, want %s", gotMonday.Format("2006-01-02"), tt.wantMonday)
			}

			// Check Sunday date
			if gotSunday.Format("2006-01-02") != tt.wantSunday {
				t.Errorf("Sunday date = %s, want %s", gotSunday.Format("2006-01-02"), tt.wantSunday)
			}

			// Check Monday time
			if gotMonday.Format("15:04:05") != tt.wantMondayTime {
				t.Errorf("Monday time = %s, want %s", gotMonday.Format("15:04:05"), tt.wantMondayTime)
			}

			// Check Sunday time
			if gotSunday.Format("15:04:05") != tt.wantSundayTime {
				t.Errorf("Sunday time = %s, want %s", gotSunday.Format("15:04:05"), tt.wantSundayTime)
			}

			// Verify Sunday is after Monday
			if !gotSunday.After(gotMonday) {
				t.Errorf("Sunday %s should be after Monday %s", gotSunday, gotMonday)
			}
		})
	}
}

func TestCalculateDateInSameWeek(t *testing.T) {
	tests := []struct {
		name         string
		currentDate  string
		targetDay    string
		targetHour   int
		targetMinute int
		wantDate     string
		wantErr      bool
	}{
		{
			name:         "Monday to Thursday same week",
			currentDate:  "2025-12-01T16:00", // Monday
			targetDay:    "Thursday",
			targetHour:   13,
			targetMinute: 0,
			wantDate:     "2025-12-04T13:00",
			wantErr:      false,
		},
		{
			name:         "Thursday to Tuesday same week (backward)",
			currentDate:  "2025-12-04T16:00", // Thursday
			targetDay:    "Tuesday",
			targetHour:   9,
			targetMinute: 0,
			wantDate:     "2025-12-02T09:00",
			wantErr:      false,
		},
		{
			name:         "Wednesday to Wednesday same week",
			currentDate:  "2025-12-03T10:00", // Wednesday
			targetDay:    "Wednesday",
			targetHour:   14,
			targetMinute: 30,
			wantDate:     "2025-12-03T14:30",
			wantErr:      false,
		},
		{
			name:         "Friday to Sunday same week",
			currentDate:  "2025-12-05T16:00", // Friday
			targetDay:    "Sunday",
			targetHour:   10,
			targetMinute: 0,
			wantDate:     "2025-12-07T10:00",
			wantErr:      false,
		},
		{
			name:         "Sunday to Monday same week",
			currentDate:  "2025-12-07T16:00", // Sunday
			targetDay:    "Monday",
			targetHour:   16,
			targetMinute: 0,
			wantDate:     "2025-12-01T16:00",
			wantErr:      false,
		},
		{
			name:         "Invalid day name",
			currentDate:  "2025-12-02T16:00",
			targetDay:    "NotADay",
			targetHour:   16,
			targetMinute: 0,
			wantDate:     "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentDate, err := time.Parse("2006-01-02T15:04", tt.currentDate)
			if err != nil {
				t.Fatalf("Failed to parse current date: %v", err)
			}

			gotDate, err := calculateDateInSameWeek(currentDate, tt.targetDay, tt.targetHour, tt.targetMinute)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			gotDateStr := gotDate.Format("2006-01-02T15:04")
			if gotDateStr != tt.wantDate {
				t.Errorf("calculateDateInSameWeek() = %s, want %s", gotDateStr, tt.wantDate)
			}

			// Verify date is within same week as current date
			monday, sunday := GetWeekBoundaries(currentDate)
			if gotDate.Before(monday) || gotDate.After(sunday) {
				t.Errorf("Calculated date %s is outside week boundaries (%s to %s)",
					gotDate.Format("2006-01-02"), monday.Format("2006-01-02"), sunday.Format("2006-01-02"))
			}
		})
	}
}

func TestApplyRandomTiming(t *testing.T) {
	tests := []struct {
		name            string
		currentDateStr  string
		recommendations []configuration.TimingRecommendation
		wantErr         bool
		validateFunc    func(t *testing.T, newDateStr string, selectedRec configuration.TimingRecommendation, currentDateStr string)
	}{
		{
			name:           "Valid recommendations - single option",
			currentDateStr: "2025-12-02T16:00", // Tuesday
			recommendations: []configuration.TimingRecommendation{
				{Day: "Thursday", Time: "13:00", Reasoning: "Test reasoning"},
			},
			wantErr: false,
			validateFunc: func(t *testing.T, newDateStr string, selectedRec configuration.TimingRecommendation, currentDateStr string) {
				if newDateStr != "2025-12-04T13:00" {
					t.Errorf("Expected new date 2025-12-04T13:00, got %s", newDateStr)
				}
				if selectedRec.Day != "Thursday" {
					t.Errorf("Expected selected day Thursday, got %s", selectedRec.Day)
				}
			},
		},
		{
			name:           "Valid recommendations - multiple options",
			currentDateStr: "2025-12-02T16:00", // Tuesday
			recommendations: []configuration.TimingRecommendation{
				{Day: "Monday", Time: "16:00", Reasoning: "Baseline"},
				{Day: "Tuesday", Time: "09:00", Reasoning: "Morning"},
				{Day: "Thursday", Time: "13:00", Reasoning: "Afternoon"},
			},
			wantErr: false,
			validateFunc: func(t *testing.T, newDateStr string, selectedRec configuration.TimingRecommendation, currentDateStr string) {
				// Parse to verify it's a valid date
				newDate, err := time.Parse("2006-01-02T15:04", newDateStr)
				if err != nil {
					t.Errorf("Failed to parse new date: %v", err)
				}

				// Verify it's in same week
				currentDate, _ := time.Parse("2006-01-02T15:04", currentDateStr)
				monday, sunday := GetWeekBoundaries(currentDate)
				if newDate.Before(monday) || newDate.After(sunday) {
					t.Errorf("New date %s is outside week boundaries", newDateStr)
				}

				// Verify selected recommendation is one of the input recommendations
				found := false
				for _, rec := range []string{"Monday", "Tuesday", "Thursday"} {
					if selectedRec.Day == rec {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Selected day %s is not in recommendations", selectedRec.Day)
				}
			},
		},
		{
			name:           "Empty recommendations",
			currentDateStr: "2025-12-02T16:00",
			recommendations: []configuration.TimingRecommendation{},
			wantErr:        true,
			validateFunc:   nil,
		},
		{
			name:           "Invalid date format",
			currentDateStr: "invalid-date",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Thursday", Time: "13:00", Reasoning: "Test"},
			},
			wantErr:      true,
			validateFunc: nil,
		},
		{
			name:           "Invalid time format in recommendation",
			currentDateStr: "2025-12-02T16:00",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Thursday", Time: "invalid-time", Reasoning: "Test"},
			},
			wantErr:      true,
			validateFunc: nil,
		},
		{
			name:           "Invalid day name in recommendation",
			currentDateStr: "2025-12-02T16:00",
			recommendations: []configuration.TimingRecommendation{
				{Day: "NotADay", Time: "13:00", Reasoning: "Test"},
			},
			wantErr:      true,
			validateFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newDateStr, selectedRec, err := ApplyRandomTiming(tt.currentDateStr, tt.recommendations)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, newDateStr, selectedRec, tt.currentDateStr)
			}
		})
	}
}

func TestApplyRandomTiming_Randomness(t *testing.T) {
	// This test verifies that with multiple recommendations, we get different results
	// over multiple runs (probabilistic test)
	currentDateStr := "2025-12-02T16:00"
	recommendations := []configuration.TimingRecommendation{
		{Day: "Monday", Time: "16:00", Reasoning: "Baseline"},
		{Day: "Tuesday", Time: "09:00", Reasoning: "Morning"},
		{Day: "Wednesday", Time: "10:00", Reasoning: "Mid-morning"},
		{Day: "Thursday", Time: "13:00", Reasoning: "Afternoon"},
		{Day: "Friday", Time: "14:00", Reasoning: "Late afternoon"},
	}

	results := make(map[string]bool)
	iterations := 50

	for i := 0; i < iterations; i++ {
		newDateStr, _, err := ApplyRandomTiming(currentDateStr, recommendations)
		if err != nil {
			t.Fatalf("Unexpected error on iteration %d: %v", i, err)
		}
		results[newDateStr] = true
	}

	// With 5 options and 50 iterations, we should see at least 2 different results
	// (extremely unlikely to get the same result 50 times in a row)
	if len(results) < 2 {
		t.Errorf("Expected at least 2 different results from %d iterations with 5 options, got %d unique results",
			iterations, len(results))
	}
}

func TestApplyRandomTiming_PreservesWeekBoundary(t *testing.T) {
	// Test that regardless of which recommendation is picked, the result stays in the same week
	currentDateStr := "2025-12-03T10:00" // Wednesday
	recommendations := []configuration.TimingRecommendation{
		{Day: "Monday", Time: "16:00", Reasoning: "Start of week"},
		{Day: "Sunday", Time: "23:00", Reasoning: "End of week"},
	}

	// Run multiple times to test both recommendations
	for i := 0; i < 20; i++ {
		newDateStr, _, err := ApplyRandomTiming(currentDateStr, recommendations)
		if err != nil {
			t.Fatalf("Unexpected error on iteration %d: %v", i, err)
		}

		// Parse dates
		currentDate, _ := time.Parse("2006-01-02T15:04", currentDateStr)
		newDate, _ := time.Parse("2006-01-02T15:04", newDateStr)

		// Get week boundaries
		monday, sunday := GetWeekBoundaries(currentDate)

		// Verify new date is within boundaries
		if newDate.Before(monday) || newDate.After(sunday) {
			t.Errorf("Iteration %d: New date %s is outside week boundaries (%s to %s)",
				i, newDateStr, monday.Format("2006-01-02"), sunday.Format("2006-01-02"))
		}
	}
}
