package ui

// Renderer handles UI coloring and styling operations
type Renderer struct{}

// NewRenderer creates a new UI renderer
func NewRenderer() *Renderer {
	return &Renderer{}
}

// ColorFromSponsoredEmails returns colored title based on sponsored email conditions
func (r *Renderer) ColorFromSponsoredEmails(title, sponsored string, sponsoredEmails string) (string, bool) {
	if len(sponsored) == 0 || sponsored == "N/A" || sponsored == "-" || len(sponsoredEmails) > 0 {
		return GreenStyle.Render(title), true
	}
	return RedStyle.Render(title), false
}

// ColorFromString returns colored title based on string value presence
func (r *Renderer) ColorFromString(title, value string) string {
	if len(value) > 0 {
		return GreenStyle.Render(title)
	}
	return RedStyle.Render(title)
}

// ColorFromStringInverse returns colored title with inverse logic
func (r *Renderer) ColorFromStringInverse(title, value string) string {
	if len(value) > 0 {
		return RedStyle.Render(title)
	}
	return GreenStyle.Render(title)
}

// ColorFromBool returns colored title based on boolean value
func (r *Renderer) ColorFromBool(title string, value bool) string {
	if value {
		return GreenStyle.Render(title)
	}
	return RedStyle.Render(title)
}