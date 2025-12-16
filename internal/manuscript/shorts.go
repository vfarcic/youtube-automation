package manuscript

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"devopstoolkit/youtube-automation/internal/storage"
)

// MarkerFormat defines the TODO marker format for Shorts
const (
	// StartMarkerFormat is the format for the start marker: TODO: Short (id: %s) (start)
	StartMarkerFormat = `TODO: Short (id: %s) (start)`
	// EndMarkerFormat is the format for the end marker: TODO: Short (id: %s) (end)
	EndMarkerFormat = `TODO: Short (id: %s) (end)`
)

// insertion represents a marker to be inserted at a specific position
type insertion struct {
	position int
	text     string
	isStart  bool // true for start marker, false for end marker
}

// InsertShortMarkers inserts TODO markers into a manuscript file for the given shorts.
// Each short's text segment will be wrapped with start and end markers.
//
// Marker format:
//   - Start: TODO: Short (id: short1) (start)
//   - End: TODO: Short (id: short1) (end)
//
// The function reads the manuscript, finds each short's text, inserts markers,
// and writes the updated content back to the file.
//
// Returns an error if the file cannot be read/written, or a warning message
// if some shorts' text could not be found in the manuscript.
func InsertShortMarkers(manuscriptPath string, shorts []storage.Short) error {
	if len(shorts) == 0 {
		return nil // Nothing to do
	}

	// Read manuscript content
	content, err := os.ReadFile(manuscriptPath)
	if err != nil {
		return fmt.Errorf("failed to read manuscript: %w", err)
	}

	manuscriptText := string(content)
	var notFound []string
	var insertions []insertion

	// Find positions for each short and prepare insertions
	for _, short := range shorts {
		start, end, found := findTextPosition(manuscriptText, short.Text)
		if !found {
			notFound = append(notFound, short.ID)
			continue
		}

		// Create start marker (inserted before the text)
		startMarker := fmt.Sprintf(StartMarkerFormat, short.ID)
		insertions = append(insertions, insertion{
			position: start,
			text:     startMarker + "\n\n",
			isStart:  true,
		})

		// Create end marker (inserted after the text)
		endMarker := fmt.Sprintf(EndMarkerFormat, short.ID)
		insertions = append(insertions, insertion{
			position: end,
			text:     "\n\n" + endMarker,
			isStart:  false,
		})
	}

	if len(insertions) == 0 {
		return fmt.Errorf("no short segments found in manuscript (tried: %v)", notFound)
	}

	// Sort insertions by position in reverse order (to avoid offset shifts)
	sort.Slice(insertions, func(i, j int) bool {
		// If same position, end markers come before start markers
		if insertions[i].position == insertions[j].position {
			return !insertions[i].isStart && insertions[j].isStart
		}
		return insertions[i].position > insertions[j].position
	})

	// Apply insertions (in reverse order so positions remain valid)
	result := manuscriptText
	for _, ins := range insertions {
		result = result[:ins.position] + ins.text + result[ins.position:]
	}

	// Write updated manuscript
	if err := os.WriteFile(manuscriptPath, []byte(result), 0644); err != nil {
		return fmt.Errorf("failed to write manuscript: %w", err)
	}

	// Return warning if some shorts were not found
	if len(notFound) > 0 {
		return fmt.Errorf("markers inserted, but %d short(s) not found in manuscript: %v", len(notFound), notFound)
	}

	return nil
}

// findTextPosition finds the position of text within content.
// It first tries an exact match, then falls back to normalized matching.
// Returns the start position, end position, and whether the text was found.
func findTextPosition(content, text string) (start, end int, found bool) {
	// Try exact match first
	idx := strings.Index(content, text)
	if idx >= 0 {
		return idx, idx + len(text), true
	}

	// Try normalized matching (collapse whitespace)
	normalizedContent := normalizeWhitespace(content)
	normalizedText := normalizeWhitespace(text)

	idx = strings.Index(normalizedContent, normalizedText)
	if idx < 0 {
		return 0, 0, false
	}

	// Map normalized position back to original content position
	start, end = mapNormalizedPosition(content, normalizedContent, idx, len(normalizedText))
	return start, end, true
}

// normalizeWhitespace collapses all whitespace sequences into single spaces
// and trims leading/trailing whitespace.
func normalizeWhitespace(s string) string {
	// Replace all whitespace sequences with single space
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// mapNormalizedPosition maps a position in normalized text back to original text.
// This is needed because normalization changes character positions.
func mapNormalizedPosition(original, normalized string, normStart, normLen int) (origStart, origEnd int) {
	// Count characters in original text, tracking normalized position
	normPos := 0
	origStart = -1
	inWhitespace := false

	for i := 0; i < len(original); i++ {
		c := original[i]
		isSpace := c == ' ' || c == '\t' || c == '\n' || c == '\r'

		// Skip leading whitespace in original
		if origStart == -1 && isSpace {
			continue
		}

		if isSpace {
			if !inWhitespace {
				// First whitespace character after non-whitespace
				if normPos >= normStart && origStart == -1 {
					origStart = i
				}
				normPos++ // Count as single space in normalized
				inWhitespace = true
			}
			// Skip additional whitespace characters
		} else {
			inWhitespace = false
			if normPos >= normStart && origStart == -1 {
				origStart = i
			}
			if normPos >= normStart+normLen {
				return origStart, i
			}
			normPos++
		}
	}

	// If we reach the end, return end of original
	return origStart, len(original)
}

// RemoveShortMarkers removes all TODO Short markers from a manuscript.
// Useful for re-analyzing or cleaning up a manuscript.
func RemoveShortMarkers(manuscriptPath string) error {
	content, err := os.ReadFile(manuscriptPath)
	if err != nil {
		return fmt.Errorf("failed to read manuscript: %w", err)
	}

	text := string(content)

	// Remove start markers: TODO: Short (id: ...) (start)
	startPattern := regexp.MustCompile(`TODO: Short \(id: [^)]+\) \(start\)\n*`)
	text = startPattern.ReplaceAllString(text, "")

	// Remove end markers: TODO: Short (id: ...) (end)
	endPattern := regexp.MustCompile(`\n*TODO: Short \(id: [^)]+\) \(end\)`)
	text = endPattern.ReplaceAllString(text, "")

	if err := os.WriteFile(manuscriptPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write manuscript: %w", err)
	}

	return nil
}

// ExtractShortText extracts the text between start and end markers for a given short ID.
// Returns the extracted text and whether it was found.
func ExtractShortText(manuscriptPath string, shortID string) (string, error) {
	content, err := os.ReadFile(manuscriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read manuscript: %w", err)
	}

	text := string(content)

	// Build patterns for this specific short ID
	startPattern := fmt.Sprintf(`TODO: Short \(id: %s\) \(start\)\n*`, regexp.QuoteMeta(shortID))
	endPattern := fmt.Sprintf(`\n*TODO: Short \(id: %s\) \(end\)`, regexp.QuoteMeta(shortID))

	startRe := regexp.MustCompile(startPattern)
	endRe := regexp.MustCompile(endPattern)

	// Find start marker
	startMatch := startRe.FindStringIndex(text)
	if startMatch == nil {
		return "", fmt.Errorf("start marker not found for short ID: %s", shortID)
	}

	// Find end marker after start
	endMatch := endRe.FindStringIndex(text[startMatch[1]:])
	if endMatch == nil {
		return "", fmt.Errorf("end marker not found for short ID: %s", shortID)
	}

	// Extract text between markers
	extracted := text[startMatch[1] : startMatch[1]+endMatch[0]]
	return strings.TrimSpace(extracted), nil
}
