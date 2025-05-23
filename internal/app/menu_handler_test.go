package app

import (
	"errors"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/cli"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/ui"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/charmbracelet/huh" // For later if we mock forms, or for []huh.Field type
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
