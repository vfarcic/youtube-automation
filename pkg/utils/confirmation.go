package utils

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
)

// ConfirmAction prompts the user for confirmation.
// message: The confirmation message to display
// Returns: Whether the action is confirmed
func ConfirmAction(message string) bool {
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

	// Run the form
	err := form.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error displaying confirmation prompt: %v\n", err)
		return false
	}

	return confirmed
}
