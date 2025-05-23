package utils

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/huh"
)

// ConfirmAction prompts the user for confirmation.
// message: The confirmation message to display
// inputReader: Optional reader for testability. If nil, defaults to os.Stdin behavior.
// Returns: Whether the action is confirmed
func ConfirmAction(message string, inputReader io.Reader) bool {
	var confirmed bool

	// Create confirmation prompt using huh library
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(message).
				Affirmative("Yes").
				Negative("No").
				Value(&confirmed),
		),
	)

	if inputReader != nil {
		form = form.WithInput(inputReader)
	}

	// Run the form
	err := form.Run()
	if err != nil {
		// In tests with controlled input, an error might mean the input was exhausted
		// or the form was aborted by specific sequences. Non-interactive huh.ErrUserAborted
		// might not apply, but other errors can occur if input is malformed or incomplete.
		fmt.Fprintf(os.Stderr, "Error displaying confirmation prompt: %v\n", err)
		return false // Default to false on form error
	}

	return confirmed
}
