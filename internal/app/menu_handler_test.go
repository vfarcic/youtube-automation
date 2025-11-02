package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/cli"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/ui"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/charmbracelet/huh" // For later if we mock forms, or for []huh.Field type
	"github.com/stretchr/testify/assert"
)

// MockConfirmAction is a helper for testing confirmations
var mockConfirmAction func(message string) bool

type mockAlwaysYesConfirmer struct{}

func (m *mockAlwaysYesConfirmer) Confirm(message string) bool {
	if mockConfirmAction != nil {
		return mockConfirmAction(message)
	}
	return true // Default to yes for tests not focused on confirmation logic
}

// TestChooseCreateVideoAndHandleError_GetFieldsError tests the scenario where cli.GetCreateVideoFields returns an error.
func TestChooseCreateVideoAndHandleError_GetFieldsError(t *testing.T) {
	// Arrange
	fs := filesystem.NewOperations() // Real filesystem, consider mock for more isolation
	mh := &MenuHandler{
		confirmer:    &mockAlwaysYesConfirmer{},
		uiRenderer:   ui.NewRenderer(),
		filesystem:   fs,
		videoManager: video.NewManager(fs.GetFilePath),
	}

	// Override the cli.GetCreateVideoFields function for this test
	originalGetCreateVideoFields := cli.GetCreateVideoFields
	defer func() { cli.GetCreateVideoFields = originalGetCreateVideoFields }() // Restore original func after test

	cli.GetCreateVideoFields = func(name, category, date *string, save *bool) ([]huh.Field, error) {
		return nil, errors.New("simulated GetCreateVideoFields error")
	}

	// Act
	_, err := mh.ChooseCreateVideoAndHandleError()

	// Assert
	if err == nil {
		t.Fatal("Expected an error when GetCreateVideoFields fails, but got nil") // Use t.Fatal to stop test on critical failure
	}

	// Check if the error is correctly wrapped and contains the expected message.
	// The implementation in menu_handler.go is: fmt.Errorf("error getting video fields: %w", err)
	if !strings.Contains(err.Error(), "error getting video fields") {
		t.Errorf("Error message '%s' does not contain expected prefix 'error getting video fields'", err.Error())
	}

	underlyingErr := errors.Unwrap(err)
	if underlyingErr == nil {
		t.Fatal("Expected a wrapped error, but errors.Unwrap(err) was nil")
	}
	if underlyingErr.Error() != "simulated GetCreateVideoFields error" {
		t.Errorf("Expected underlying error to be 'simulated GetCreateVideoFields error', but got '%s'", underlyingErr.Error())
	}
}

// TODO: Add test for ChooseIndex when form.Run fails
// TODO: Add test for ChooseCreateVideoAndHandleError when form.Run fails (would need huh form mocking/testing strategy)
// TODO: Add test for ChooseCreateVideoAndHandleError for file/dir operation failures (would need filesystem mocking)

func TestGetEditPhaseOptionText(t *testing.T) {
	mh := &MenuHandler{
		greenStyle:  ui.GreenStyle,
		orangeStyle: ui.OrangeStyle,
	}

	tests := []struct {
		name      string
		phaseName string
		completed int
		total     int
		expected  string
	}{
		{"CompletedPhase", "Initial", 5, 5, mh.greenStyle.Render("Initial (5/5)")},
		{"IncompletePhase", "Work", 2, 5, mh.orangeStyle.Render("Work (2/5)")},
		{"ZeroTotal", "Define", 0, 0, mh.orangeStyle.Render("Define (0/0)")},      // Should be orange if total is 0
		{"CompletedZeroTotal", "Post", 0, 0, mh.orangeStyle.Render("Post (0/0)")}, // Even if completed is also 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.getEditPhaseOptionText(tt.phaseName, tt.completed, tt.total)
			if got != tt.expected {
				t.Errorf("getEditPhaseOptionText(%q, %d, %d) = %q, want %q", tt.phaseName, tt.completed, tt.total, got, tt.expected)
			}
		})
	}
}

func TestColorTitleSponsoredEmails(t *testing.T) {
	mh := &MenuHandler{
		aspectService: aspect.NewService(),
		greenStyle:    ui.GreenStyle,
		orangeStyle:   ui.OrangeStyle,
	}
	tests := []struct {
		name            string
		title           string
		sponsoredAmount string
		sponsoredEmails string
		expected        string
	}{
		{"AmountEmptyEmailsEmpty", "Sponsor", "", "", mh.greenStyle.Render("Sponsor")},
		{"AmountEmptyEmailsExist", "Sponsor", "", "test@example.com", mh.greenStyle.Render("Sponsor")},
		{"AmountNAEmailsExist", "Sponsor", "N/A", "test@example.com", mh.greenStyle.Render("Sponsor")},
		{"AmountDashEmailsExist", "Sponsor", "-", "test@example.com", mh.greenStyle.Render("Sponsor")},
		{"AmountExistsEmailsEmpty", "Sponsor", "$100", "", mh.orangeStyle.Render("Sponsor")},
		{"AmountAndEmailsExist", "Sponsor", "$100", "test@example.com", mh.greenStyle.Render("Sponsor")},
		{"AllEmpty", "Sponsor", "", "", mh.greenStyle.Render("Sponsor")}, // Amount empty -> green
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.colorTitleSponsoredEmails(tt.title, tt.sponsoredAmount, tt.sponsoredEmails)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestColorTitleStringInverse(t *testing.T) {
	mh := &MenuHandler{
		aspectService: aspect.NewService(),
		greenStyle:    ui.GreenStyle,
		orangeStyle:   ui.OrangeStyle,
	}
	tests := []struct {
		name     string
		title    string
		value    string
		expected string
	}{
		{"EmptyValue", "TitleInv1", "", mh.greenStyle.Render("TitleInv1")},
		{"NonEmptyValue", "TitleInv2", "SomeValue", mh.orangeStyle.Render("TitleInv2")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.colorTitleStringInverse(tt.title, tt.value)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestColorTitleBoolInverse(t *testing.T) {
	mh := &MenuHandler{
		aspectService: aspect.NewService(),
		greenStyle:    ui.GreenStyle,
		orangeStyle:   ui.OrangeStyle,
	}
	tests := []struct {
		name     string
		title    string
		value    bool
		expected string
	}{
		{"TrueValueInv", "TitleBoolInvT", true, mh.orangeStyle.Render("TitleBoolInvT")},
		{"FalseValueInv", "TitleBoolInvF", false, mh.greenStyle.Render("TitleBoolInvF")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.colorTitleBoolInverse(tt.title, tt.value)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestColorTitleSponsorshipAmount(t *testing.T) {
	mh := &MenuHandler{
		aspectService: aspect.NewService(),
		greenStyle:    ui.GreenStyle,  // Initialize unexported field
		orangeStyle:   ui.OrangeStyle, // Initialize unexported field
	}

	tests := []struct {
		name     string
		title    string
		value    string
		expected string // Expected to be a styled string
	}{
		{"Amount Present", "Sponsorship Amount", "100", mh.greenStyle.Render("Sponsorship Amount")},
		{"Amount Present N/A", "Sponsorship Amount", "N/A", mh.greenStyle.Render("Sponsorship Amount")},
		{"Amount Present Dash", "Sponsorship Amount", "-", mh.greenStyle.Render("Sponsorship Amount")},
		{"Amount Empty", "Sponsorship Amount", "", mh.orangeStyle.Render("Sponsorship Amount")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := mh.colorTitleSponsorshipAmount(tc.title, tc.value)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCountCompletedTasks(t *testing.T) {
	mh := &MenuHandler{} // countCompletedTasks doesn't depend on MenuHandler fields

	tests := []struct {
		name              string
		fields            []interface{}
		expectedCompleted int
		expectedTotal     int
	}{
		{
			name:              "EmptyList",
			fields:            []interface{}{},
			expectedCompleted: 0,
			expectedTotal:     0,
		},
		{
			name:              "NoCompletedTasks",
			fields:            []interface{}{false, false, false},
			expectedCompleted: 0,
			expectedTotal:     3,
		},
		{
			name:              "SomeCompletedTasks",
			fields:            []interface{}{true, false, true, false},
			expectedCompleted: 2,
			expectedTotal:     4,
		},
		{
			name:              "AllTasksCompleted",
			fields:            []interface{}{true, true, true},
			expectedCompleted: 3,
			expectedTotal:     3,
		},
		{
			name:              "MixedWithNonBooleanAtEnd",
			fields:            []interface{}{true, false, "not a boolean"},
			expectedCompleted: 2,
			expectedTotal:     3,
		},
		{
			name:              "MixedWithNonBooleanAtStart",
			fields:            []interface{}{"not a boolean", true, false},
			expectedCompleted: 2,
			expectedTotal:     3,
		},
		{
			name:              "OnlyNonBooleans",
			fields:            []interface{}{"string1", 123, "string2"},
			expectedCompleted: 2,
			expectedTotal:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completed, total := mh.countCompletedTasks(tt.fields)
			assert.Equal(t, tt.expectedCompleted, completed, "Completed count mismatch")
			assert.Equal(t, tt.expectedTotal, total, "Total count mismatch")
		})
	}
}

func TestGetPhaseText(t *testing.T) {
	mh := &MenuHandler{
		greenStyle:  ui.GreenStyle,
		orangeStyle: ui.OrangeStyle,
	}
	tests := []struct {
		name      string
		text      string
		completed int
		total     int
		expected  string
	}{
		{"CompletedTasks", "Phase 1", 5, 5, mh.greenStyle.Render("Phase 1 (5/5)")},
		{"IncompleteTasks", "Phase 2", 2, 5, mh.orangeStyle.Render("Phase 2 (2/5)")},
		{"ZeroTotalTasks", "Phase 3", 0, 0, mh.orangeStyle.Render("Phase 3 (0/0)")},
		{"CompletedZeroTotal", "Phase 4", 0, 0, mh.orangeStyle.Render("Phase 4 (0/0)")},
		{"MoreCompletedThanTotal", "Phase 5", 6, 5, mh.orangeStyle.Render("Phase 5 (6/5)")}, // Should still be orange as it's not '== total'
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.GetPhaseText(tt.text, tt.completed, tt.total)
			if got != tt.expected {
				t.Errorf("GetPhaseText(%q, %d, %d) = %q, want %q", tt.text, tt.completed, tt.total, got, tt.expected)
			}
		})
	}
}

func TestColorTitleString(t *testing.T) {
	mh := &MenuHandler{
		aspectService: aspect.NewService(),
		greenStyle:    ui.GreenStyle,
		orangeStyle:   ui.OrangeStyle,
	}
	tests := []struct {
		name     string
		title    string
		value    string
		expected string
	}{
		{"EmptyValue", "TitleStr1", "", mh.orangeStyle.Render("TitleStr1")},
		{"NonEmptyValue", "TitleStr2", "SomeValue", mh.greenStyle.Render("TitleStr2")},
		{"DashValue", "TitleStr3", "-", mh.orangeStyle.Render("TitleStr3")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.colorTitleString(tt.title, tt.value)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestColorTitleBool(t *testing.T) {
	mh := &MenuHandler{
		aspectService: aspect.NewService(),
		greenStyle:    ui.GreenStyle,
		orangeStyle:   ui.OrangeStyle,
	}
	tests := []struct {
		name     string
		title    string
		value    bool
		expected string
	}{
		{"TrueValue", "TitleBool1", true, mh.greenStyle.Render("TitleBool1")},
		{"FalseValue", "TitleBool2", false, mh.orangeStyle.Render("TitleBool2")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.colorTitleBool(tt.title, tt.value)
			assert.Equal(t, tt.expected, got)
		})
	}
}

const (
	phaseTestIdeas = iota
	phaseTestStarted
	phaseTestMaterialDone
	phaseTestEditRequested
	phaseTestPublishPending
	phaseTestPublished
)

// Actual actionReturn from menu_handler.go (usually in an enum or const block like index...)
// For testing, we'll use its typical value. If it changes, this might need an update
// or a more robust way to share it (e.g., exporting it if it were a package-level const).
const testHandlerActionReturn = 99

func TestGetPhaseColoredText(t *testing.T) {
	mh := &MenuHandler{
		greenStyle:  ui.GreenStyle,
		orangeStyle: ui.OrangeStyle,
	}

	tests := []struct {
		name     string
		phases   map[int]int
		phase    int
		title    string
		expected string
	}{
		{"IdeasNotEnough", map[int]int{phaseTestIdeas: 2}, phaseTestIdeas, "Ideas", mh.orangeStyle.Render("Ideas (2)")},
		{"IdeasEnough", map[int]int{phaseTestIdeas: 3}, phaseTestIdeas, "Ideas", mh.greenStyle.Render("Ideas (3)")},
		{"StartedNotEnough", map[int]int{phaseTestStarted: 2}, phaseTestStarted, "Started", mh.orangeStyle.Render("Started (2)")},
		{"StartedEnough", map[int]int{phaseTestStarted: 3}, phaseTestStarted, "Started", mh.greenStyle.Render("Started (3)")},
		{"MaterialDoneNotEnough", map[int]int{phaseTestMaterialDone: 2}, phaseTestMaterialDone, "Material Done", mh.orangeStyle.Render("Material Done (2)")},
		{"MaterialDoneEnough", map[int]int{phaseTestMaterialDone: 3}, phaseTestMaterialDone, "Material Done", mh.greenStyle.Render("Material Done (3)")},
		{"EditRequestedZero", map[int]int{phaseTestEditRequested: 0}, phaseTestEditRequested, "Edit Requested", mh.orangeStyle.Render("Edit Requested (0)")},
		{"EditRequestedPositive", map[int]int{phaseTestEditRequested: 1}, phaseTestEditRequested, "Edit Requested", mh.greenStyle.Render("Edit Requested (1)")},
		{"PublishPendingZero", map[int]int{phaseTestPublishPending: 0}, phaseTestPublishPending, "Publish Pending", mh.orangeStyle.Render("Publish Pending (0)")},
		{"PublishPendingPositive", map[int]int{phaseTestPublishPending: 1}, phaseTestPublishPending, "Publish Pending", mh.greenStyle.Render("Publish Pending (1)")},
		{"Published", map[int]int{phaseTestPublished: 5}, phaseTestPublished, "Published", mh.greenStyle.Render("Published (5)")},
		{"ActionReturn", map[int]int{}, testHandlerActionReturn, "Return", "Return"},
		{"UnknownPhase", map[int]int{100: 1}, 100, "Unknown", mh.orangeStyle.Render("Unknown (1)")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.getPhaseColoredText(tt.phases, tt.phase, tt.title)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetPhaseColoredTextWithCount(t *testing.T) {
	mh := &MenuHandler{
		greenStyle:  ui.GreenStyle,
		orangeStyle: ui.OrangeStyle,
	}

	tests := []struct {
		name          string
		phases        map[int]int
		phase         int
		title         string
		expectedText  string
		expectedCount int
	}{
		{"IdeasEnough", map[int]int{phaseTestIdeas: 3}, phaseTestIdeas, "Ideas", mh.greenStyle.Render("Ideas (3)"), 3},
		{"EditRequestedZero", map[int]int{phaseTestEditRequested: 0}, phaseTestEditRequested, "Edit Requested", "Edit Requested", 0}, // Special case: if count is 0, title is not formatted
		{"PublishedWithCount", map[int]int{phaseTestPublished: 5}, phaseTestPublished, "Published", mh.greenStyle.Render("Published (5)"), 5},
		{"ActionReturnNoCount", map[int]int{}, testHandlerActionReturn, "Return", "Return", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotCount := mh.getPhaseColoredTextWithCount(tt.phases, tt.phase, tt.title)
			assert.Equal(t, tt.expectedText, gotText, "Text mismatch")
			assert.Equal(t, tt.expectedCount, gotCount, "Count mismatch")
		})
	}
}

func TestGetVideoTitleForDisplay(t *testing.T) {
	mh := &MenuHandler{
		greenStyle:     ui.GreenStyle,
		orangeStyle:    ui.OrangeStyle,
		farFutureStyle: ui.FarFutureStyle, // Assuming this is how farFutureStyle is initialized
	}

	// Define a reference time for consistent date comparisons
	// For IsFarFutureDate, a date is "far future" if it's more than 7 days from referenceTime.
	referenceTime := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	today := referenceTime.Format("2006-01-02T15:04")
	farFutureDate := referenceTime.Add(10 * 24 * time.Hour).Format("2006-01-02T15:04")   // 10 days in future
	notFarFutureDate := referenceTime.Add(3 * 24 * time.Hour).Format("2006-01-02T15:04") // 3 days in future

	tests := []struct {
		name          string
		video         storage.Video
		currentPhase  int // Using the same local constants as above for phases
		referenceTime time.Time
		expected      string
	}{
		{
			name:          "BasicTitleNoSpecialConditions",
			video:         storage.Video{Name: "Test Video 1"},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      "Test Video 1",
		},
		{
			name:          "SponsoredNotBlocked",
			video:         storage.Video{Name: "Sponsored Vid", Sponsorship: storage.Sponsorship{Amount: "$100"}},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      mh.orangeStyle.Render("Sponsored Vid") + " (S)",
		},
		{
			name:          "BlockedSponsor",
			video:         storage.Video{Name: "Blocked Vid", Sponsorship: storage.Sponsorship{Amount: "$100", Blocked: "Legal"}},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      "Blocked Vid (Legal)", // Blocked takes precedence for main styling, (S) is not added if blocked
		},
		{
			name:          "BlockedNoSponsorAmount",
			video:         storage.Video{Name: "Blocked Vid 2", Sponsorship: storage.Sponsorship{Blocked: "Waiting"}},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      "Blocked Vid 2 (Waiting)",
		},
		{
			name:          "StartedPhaseFarFutureDate",
			video:         storage.Video{Name: "Future Video", Date: farFutureDate},
			currentPhase:  phaseTestStarted,
			referenceTime: referenceTime,
			expected:      mh.farFutureStyle.Render("Future Video") + fmt.Sprintf(" (%s)", farFutureDate),
		},
		{
			name:          "StartedPhaseNotFarFutureDate",
			video:         storage.Video{Name: "Soon Video", Date: notFarFutureDate},
			currentPhase:  phaseTestStarted, // Not far future, no special style from this rule
			referenceTime: referenceTime,
			expected:      fmt.Sprintf("Soon Video (%s)", notFarFutureDate),
		},
		{
			name:          "SponsoredAndFarFutureStarted (Sponsored wins for color)",
			video:         storage.Video{Name: "Sponsored Future", Sponsorship: storage.Sponsorship{Amount: "$50"}, Date: farFutureDate},
			currentPhase:  phaseTestStarted,
			referenceTime: referenceTime,
			expected:      mh.orangeStyle.Render("Sponsored Future") + fmt.Sprintf(" (%s) (S)", farFutureDate),
		},
		{
			name:          "BlockedAndFarFutureStarted (Blocked wins)",
			video:         storage.Video{Name: "Blocked Future", Sponsorship: storage.Sponsorship{Blocked: "Hold"}, Date: farFutureDate},
			currentPhase:  phaseTestStarted,
			referenceTime: referenceTime,
			expected:      "Blocked Future (Hold)", // Corrected: Date is NOT shown when blocked
		},
		{
			name:          "AMACategory",
			video:         storage.Video{Name: "AMA Time", Category: "ama"},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      "AMA Time (AMA)",
		},
		{
			name:          "AMACategoryWithDateAndSponsor",
			video:         storage.Video{Name: "Complex AMA", Category: "ama", Date: today, Sponsorship: storage.Sponsorship{Amount: "$10"}},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      mh.orangeStyle.Render("Complex AMA") + fmt.Sprintf(" (%s) (S) (AMA)", today),
		},
		{
			name:          "DatePresentNoSponsorNoBlock",
			video:         storage.Video{Name: "Dated Video", Date: today},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      fmt.Sprintf("Dated Video (%s)", today),
		},
		{
			name:          "SponsoredNoAmount (treated as not sponsored for S tag)",
			video:         storage.Video{Name: "Free Sponsor", Sponsorship: storage.Sponsorship{Amount: ""}, Date: today},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      fmt.Sprintf("Free Sponsor (%s)", today),
		},
		{
			name:          "SponsoredAmountNA (treated as not sponsored for S tag)",
			video:         storage.Video{Name: "NA Sponsor", Sponsorship: storage.Sponsorship{Amount: "N/A"}, Date: today},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      fmt.Sprintf("NA Sponsor (%s)", today),
		},
		{
			name:          "SponsoredAmountDash (treated as not sponsored for S tag)",
			video:         storage.Video{Name: "Dash Sponsor", Sponsorship: storage.Sponsorship{Amount: "-"}, Date: today},
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      fmt.Sprintf("Dash Sponsor (%s)", today),
		},
		{
			name:          "BlockedReasonEmptyString (not blocked)",
			video:         storage.Video{Name: "Blocked Reason Empty", Sponsorship: storage.Sponsorship{Blocked: ""}}, // Empty string means NOT blocked
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      "Blocked Reason Empty", // No (B), no date, no (S) as Amount is also empty
		},
		{
			name:          "BlockedReasonDash",
			video:         storage.Video{Name: "Blocked Reason Dash", Sponsorship: storage.Sponsorship{Blocked: "-"}}, // Dash means blocked, uses (B)
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      "Blocked Reason Dash (B)",
		},
		{
			name:          "BlockedReasonNA",
			video:         storage.Video{Name: "Blocked Reason NA", Sponsorship: storage.Sponsorship{Blocked: "N/A"}}, // N/A means blocked, uses (B)
			currentPhase:  phaseTestIdeas,
			referenceTime: referenceTime,
			expected:      "Blocked Reason NA (B)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure utils.IsFarFutureDate is available or mock it if it becomes problematic for CI/CD
			got := mh.getVideoTitleForDisplay(tt.video, tt.currentPhase, tt.referenceTime)
			assert.Equal(t, tt.expected, got)
		})
	}
}
