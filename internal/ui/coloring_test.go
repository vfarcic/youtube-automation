package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRenderer(t *testing.T) {
	renderer := NewRenderer()
	assert.NotNil(t, renderer, "NewRenderer should return a non-nil Renderer struct")
}

func TestColorFromSponsoredEmails(t *testing.T) {
	renderer := NewRenderer()
	title := "Test Title"

	tests := []struct {
		name            string
		sponsored       string
		sponsoredEmails string
		expectGreen     bool
	}{
		{"no sponsorship", "", "", true},
		{"sponsored N/A", "N/A", "", true},
		{"sponsored hyphen", "-", "", true},
		{"sponsored with emails", "yes", "emails@example.com", true},
		{"sponsored no emails", "yes", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, isGreen := renderer.ColorFromSponsoredEmails(title, tt.sponsored, tt.sponsoredEmails)
			assert.Equal(t, tt.expectGreen, isGreen)
			// We can't easily assert the exact styled string without more complex lipgloss mocking
			// or inspecting private fields, so we focus on the boolean logic.
		})
	}
}

func TestColorFromString(t *testing.T) {
	renderer := NewRenderer()
	title := "Test Title"

	assert.Equal(t, GreenStyle.Render(title), renderer.ColorFromString(title, "some value"))
	assert.Equal(t, RedStyle.Render(title), renderer.ColorFromString(title, ""))
}

func TestColorFromStringInverse(t *testing.T) {
	renderer := NewRenderer()
	title := "Test Title"

	assert.Equal(t, RedStyle.Render(title), renderer.ColorFromStringInverse(title, "some value"))
	assert.Equal(t, GreenStyle.Render(title), renderer.ColorFromStringInverse(title, ""))
}

func TestColorFromBool(t *testing.T) {
	renderer := NewRenderer()
	title := "Test Title"

	assert.Equal(t, GreenStyle.Render(title), renderer.ColorFromBool(title, true))
	assert.Equal(t, RedStyle.Render(title), renderer.ColorFromBool(title, false))
}
