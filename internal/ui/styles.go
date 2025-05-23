package ui

import "github.com/charmbracelet/lipgloss"

// Style definitions for the UI
var (
	RedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("1"))

	GreenStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("2"))

	OrangeStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("3"))

	FarFutureStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")) // Cyan

	ConfirmationStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#006E14")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(5).
		PaddingRight(5).
		MarginTop(1)

	ErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FF0000")).
		PaddingTop(1).
		PaddingBottom(1).
		PaddingLeft(5).
		PaddingRight(5).
		MarginTop(1)
)