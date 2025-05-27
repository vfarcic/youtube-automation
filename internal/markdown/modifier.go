package markdown

import (
	"fmt"
	"os"

	// "regexp" // No longer needed for this simple version
	"strings"
)

// ApplyHighlightsInGist reads a Gist (Markdown) file, applies bolding to the specified highlight phrases,
// and writes the modified content back to the file.
// It uses a simple string replacement strategy.
func ApplyHighlightsInGist(gistPath string, highlightsToApply []string) error {
	contentBytes, err := os.ReadFile(gistPath)
	if err != nil {
		return fmt.Errorf("failed to read gist file %s: %w", gistPath, err)
	}
	content := string(contentBytes)

	for _, phrase := range highlightsToApply {
		if strings.TrimSpace(phrase) == "" {
			continue
		}

		// For simplicity, we'll do a direct replacement.
		// A more advanced solution would parse Markdown or use complex regex with lookarounds
		// to perfectly avoid double-bolding or bolding within already styled text.
		boldedPhrase := "**" + phrase + "**"
		content = strings.ReplaceAll(content, phrase, boldedPhrase)

		// Attempt to clean up accidental quadruple asterisks that might result from replacing
		// a phrase that was somehow adjacent to existing bold markers.
		// This handles ****word**** -> **word**.
		// It does not handle cases like *word* becoming ***word***.
		content = strings.ReplaceAll(content, "****", "**")
	}

	err = os.WriteFile(gistPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write modified content to gist file %s: %w", gistPath, err)
	}

	return nil
}
