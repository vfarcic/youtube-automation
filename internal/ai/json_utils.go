package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExtractJSONFromResponse attempts to extract JSON from AI response text.
// It handles both direct JSON and JSON wrapped in markdown code blocks.
//
// Returns the extracted JSON string, or the original content if no extraction needed.
func ExtractJSONFromResponse(content string) string {
	// Try to find JSON in markdown code blocks first
	jsonContent := extractJSONFromMarkdown(content)
	if jsonContent != "" {
		return jsonContent
	}

	// If no code blocks found, return original content (might be direct JSON)
	return strings.TrimSpace(content)
}

// ParseJSONResponse extracts and parses JSON from an AI response into the provided target.
//
// Parameters:
//   - response: Raw AI response (might contain JSON in markdown code blocks)
//   - target: Pointer to struct/slice to unmarshal JSON into
//
// Returns error if JSON extraction or parsing fails.
func ParseJSONResponse(response string, target interface{}) error {
	// Extract JSON (handles markdown code blocks)
	jsonContent := ExtractJSONFromResponse(response)

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(jsonContent), target); err != nil {
		// If parsing fails, return error with excerpt
		excerpt := jsonContent
		if len(excerpt) > 200 {
			excerpt = excerpt[:200] + "..."
		}
		return fmt.Errorf("could not parse JSON from AI response. Response starts with: %s", excerpt)
	}

	return nil
}

// extractJSONFromMarkdown attempts to extract JSON from markdown code blocks.
// Looks for ```json or ``` code blocks and returns the content inside.
//
// Returns empty string if no code blocks found.
func extractJSONFromMarkdown(content string) string {
	// Look for ```json or ``` code blocks
	startMarkers := []string{"```json\n", "```\n"}
	endMarker := "```"

	for _, startMarker := range startMarkers {
		startIdx := strings.Index(content, startMarker)
		if startIdx == -1 {
			continue
		}

		startIdx += len(startMarker)
		endIdx := strings.Index(content[startIdx:], endMarker)
		if endIdx == -1 {
			continue
		}

		return strings.TrimSpace(content[startIdx : startIdx+endIdx])
	}

	return ""
}
