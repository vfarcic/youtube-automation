package app

import (
	"errors"
	"strings"
	"testing"

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

	cli.GetCreateVideoFields = func(name, category *string, save *bool) ([]huh.Field, error) {
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
		greenStyle:  ui.GreenStyle,
		orangeStyle: ui.OrangeStyle,
	}
	tests := []struct {
		name            string
		title           string
		sponsoredAmount string
		sponsoredEmails string
		expected        string
	}{
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
		greenStyle:  ui.GreenStyle,
		orangeStyle: ui.OrangeStyle,
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
		greenStyle:  ui.GreenStyle,
		orangeStyle: ui.OrangeStyle,
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
		greenStyle:  ui.GreenStyle,  // Initialize unexported field
		orangeStyle: ui.OrangeStyle, // Initialize unexported field
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
			// Call the unexported method directly as we are in the same package
			assert.Equal(t, tc.expected, mh.colorTitleSponsorshipAmount(tc.title, tc.value))
		})
	}
}

func TestGetPhaseText(t *testing.T) {
	mh := &MenuHandler{
		greenStyle:  ui.GreenStyle,
		orangeStyle: ui.OrangeStyle,
	}
	tests := []struct {
		name     string
		text     string
		task     storage.Tasks
		expected string
	}{
		{"CompletedTasks", "Phase 1", storage.Tasks{Completed: 5, Total: 5}, mh.greenStyle.Render("Phase 1 (5/5)")},
		{"IncompleteTasks", "Phase 2", storage.Tasks{Completed: 2, Total: 5}, mh.orangeStyle.Render("Phase 2 (2/5)")},
		{"ZeroTotalTasks", "Phase 3", storage.Tasks{Completed: 0, Total: 0}, mh.orangeStyle.Render("Phase 3 (0/0)")},
		{"CompletedZeroTotal", "Phase 4", storage.Tasks{Completed: 0, Total: 0}, mh.orangeStyle.Render("Phase 4 (0/0)")},
		{"MoreCompletedThanTotal", "Phase 5", storage.Tasks{Completed: 6, Total: 5}, mh.orangeStyle.Render("Phase 5 (6/5)")}, // Should still be orange as it's not '== total'
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mh.GetPhaseText(tt.text, tt.task)
			if got != tt.expected {
				t.Errorf("GetPhaseText(%q, %+v) = %q, want %q", tt.text, tt.task, got, tt.expected)
			}
		})
	}
}
