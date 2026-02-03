package publishing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devopstoolkit/youtube-automation/internal/configuration"
)

const (
	// SponsorAnalyticsStartMarker marks the beginning of the analytics section
	SponsorAnalyticsStartMarker = "<!-- SPONSOR_ANALYTICS_START -->"
	// SponsorAnalyticsEndMarker marks the end of the analytics section
	SponsorAnalyticsEndMarker = "<!-- SPONSOR_ANALYTICS_END -->"
)

// GetSponsorPagePath returns the full path to the Hugo sponsor page
func GetSponsorPagePath() string {
	return filepath.Join(configuration.GlobalSettings.Hugo.Path, "content", "sponsor", "_index.md")
}

// ReadSponsorPage reads the sponsor page content from Hugo
func ReadSponsorPage() (string, error) {
	path := GetSponsorPagePath()
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read sponsor page at %s: %w", path, err)
	}
	return string(content), nil
}

// UpdateSponsorPageAnalytics replaces the analytics section between markers.
// If markers don't exist, appends the section at the end of the file.
func UpdateSponsorPageAnalytics(newSection string) error {
	content, err := ReadSponsorPage()
	if err != nil {
		return err
	}

	updatedContent := updateContentBetweenMarkers(content, SponsorAnalyticsStartMarker, SponsorAnalyticsEndMarker, newSection)

	path := GetSponsorPagePath()
	if err := os.WriteFile(path, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write sponsor page at %s: %w", path, err)
	}

	return nil
}

// updateContentBetweenMarkers performs the actual marker-based replacement.
// If markers exist, replaces everything between them (inclusive of markers).
// If markers don't exist, appends the section at the end.
func updateContentBetweenMarkers(content, startMarker, endMarker, newSection string) string {
	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	// If both markers exist and are in correct order, replace content between them
	if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
		// Replace everything from start marker to end of end marker
		endIdx += len(endMarker)
		return content[:startIdx] + newSection + content[endIdx:]
	}

	// Markers don't exist or are malformed - append at end
	if content == "" {
		return newSection
	}

	// Ensure there's a newline before appending
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + "\n" + newSection
}
