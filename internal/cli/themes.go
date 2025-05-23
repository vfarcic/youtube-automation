package cli

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// GetCustomHuhTheme creates a custom theme for huh forms.
// This theme allows pre-rendered styles for unselected options to show through
// and applies a distinct style for selected options.
func GetCustomHuhTheme() *huh.Theme {
	theme := huh.ThemeCharm() // Start with a copy of the Charm theme

	// For UNSELECTED options:
	// Apply an empty style. This allows pre-rendered styles (like cyan for far-future videos)
	// from getVideoTitleForDisplay to be visible in the resting state.
	theme.Focused.UnselectedOption = lipgloss.NewStyle()

	// For SELECTED options (when hovered/active):
	// Apply a style that gives clear visual feedback and overrides pre-rendered styles.
	// Reverse(true) swaps the item's foreground and background colors.
	theme.Focused.SelectedOption = lipgloss.NewStyle().Reverse(true)

	// Ensure the selector (e.g., '>') remains styled as per Charm theme (fuchsia).
	// This might be inherited correctly, but re-asserting to be safe.
	theme.Focused.SelectSelector = lipgloss.NewStyle().Foreground(lipgloss.Color("#F780E2")).SetString("> ")

	return theme
}
