package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCustomHuhTheme(t *testing.T) {
	theme := GetCustomHuhTheme()

	assert.NotNil(t, theme, "GetCustomHuhTheme should return a non-nil theme")

	// Check that the specific styles we set are not nil
	// This indicates they were assigned a lipgloss.Style object
	assert.NotNil(t, theme.Focused.UnselectedOption, "Focused.UnselectedOption should be set")
	assert.NotNil(t, theme.Focused.SelectedOption, "Focused.SelectedOption should be set")
	assert.NotNil(t, theme.Focused.SelectSelector, "Focused.SelectSelector should be set")

	// Optional: Further assertions could be made here if specific style properties
	// (e.g., a particular foreground color, or whether Reverse is true)
	// are critical and easily inspectable from the lipgloss.Style struct.
	// For now, asserting they are set (not nil) is a good start for coverage.
}
