package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/storage"

	"gopkg.in/yaml.v3"
)

// VideoABData represents a video with A/B test title data enriched with analytics.
type VideoABData struct {
	Category            string
	Date                string
	DayOfWeek           string
	VideoID             string
	Titles              []storage.TitleVariant
	FirstWeekViews      int64
	FirstWeekLikes      int64
	FirstWeekComments   int64
	FirstWeekEngagement float64
	HasAnalytics        bool
}

// HasABData returns true if the title variants contain valid A/B test data:
// at least 2 variants with at least one having a non-zero share value.
func HasABData(titles []storage.TitleVariant) bool {
	if len(titles) < 2 {
		return false
	}
	for _, t := range titles {
		if t.Share > 0 {
			return true
		}
	}
	return false
}

// readIndex reads a YAML index file and returns the entries.
// Returns an empty slice (not an error) if the file doesn't exist.
func readIndex(path string) ([]storage.VideoIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read index file %s: %w", path, err)
	}
	var index []storage.VideoIndex
	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index from %s: %w", path, err)
	}
	return index, nil
}

// readVideoFile reads a single video YAML file and returns the Video.
func readVideoFile(path string) (storage.Video, error) {
	var video storage.Video
	data, err := os.ReadFile(path)
	if err != nil {
		return video, fmt.Errorf("failed to read video file %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &video); err != nil {
		return video, fmt.Errorf("failed to unmarshal video from %s: %w", path, err)
	}
	return video, nil
}

// sanitizeName applies the same sanitization as filesystem.Operations.SanitizeName
func sanitizeName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "-")
}

// LoadVideosWithABData loads videos from the current year index and previous year
// archive index, reads each video's YAML file, and returns only those with valid
// A/B test data (2+ title variants, at least one share > 0) and a videoId.
//
// Parameters:
//   - indexPath: path to the current year index.yaml
//   - dataDir: root data directory (archive indexes are at dataDir/index/{year}.yaml)
//   - manuscriptDir: base directory for video YAML files (e.g., dataDir/manuscript)
func LoadVideosWithABData(indexPath, dataDir, manuscriptDir string) ([]VideoABData, error) {
	// Load current year index
	currentIndex, err := readIndex(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read current index: %w", err)
	}

	// Load previous year archive index
	previousYear := time.Now().Year() - 1
	archivePath := filepath.Join(dataDir, "index", fmt.Sprintf("%d.yaml", previousYear))
	archiveIndex, err := readIndex(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read archive index: %w", err)
	}

	// Combine indexes, deduplicating by name+category
	seen := make(map[string]bool)
	var allEntries []storage.VideoIndex

	for _, entry := range currentIndex {
		key := entry.Category + "/" + entry.Name
		if !seen[key] {
			seen[key] = true
			allEntries = append(allEntries, entry)
		}
	}
	for _, entry := range archiveIndex {
		key := entry.Category + "/" + entry.Name
		if !seen[key] {
			seen[key] = true
			allEntries = append(allEntries, entry)
		}
	}

	var result []VideoABData
	for _, entry := range allEntries {
		name := sanitizeName(entry.Name)
		category := strings.ReplaceAll(strings.ToLower(entry.Category), " ", "-")
		videoPath := filepath.Join(manuscriptDir, category, name+".yaml")

		video, err := readVideoFile(videoPath)
		if err != nil {
			// Skip videos whose files can't be read (deleted, moved, etc.)
			continue
		}

		// Must have a videoId (published videos only)
		if video.VideoId == "" {
			continue
		}

		// Must have valid A/B data
		if !HasABData(video.Titles) {
			continue
		}

		dayOfWeek := ""
		if video.Date != "" {
			if t, err := time.Parse("2006-01-02T15:04", video.Date); err == nil {
				dayOfWeek = t.Weekday().String()
			} else if t, err := time.Parse("2006-01-02", video.Date); err == nil {
				dayOfWeek = t.Weekday().String()
			}
		}

		result = append(result, VideoABData{
			Category:  category,
			Date:      video.Date,
			DayOfWeek: dayOfWeek,
			VideoID:   video.VideoId,
			Titles:    video.Titles,
		})
	}

	return result, nil
}

// EnrichWithAnalytics joins VideoABData with YouTube Analytics data by video ID,
// populating first-week metrics. Videos without matching analytics keep HasAnalytics=false.
func EnrichWithAnalytics(videos []VideoABData, analytics []publishing.VideoAnalytics) []VideoABData {
	analyticsMap := make(map[string]publishing.VideoAnalytics, len(analytics))
	for _, a := range analytics {
		analyticsMap[a.VideoID] = a
	}

	enriched := make([]VideoABData, len(videos))
	for i, v := range videos {
		enriched[i] = v
		if a, ok := analyticsMap[v.VideoID]; ok {
			enriched[i].FirstWeekViews = a.FirstWeekViews
			enriched[i].FirstWeekLikes = a.FirstWeekLikes
			enriched[i].FirstWeekComments = a.FirstWeekComments
			enriched[i].HasAnalytics = true

			if a.FirstWeekViews > 0 {
				enriched[i].FirstWeekEngagement = float64(a.FirstWeekLikes+a.FirstWeekComments) / float64(a.FirstWeekViews) * 100
			}
		}
	}
	return enriched
}

// FormatABDataForPrompt formats the A/B test dataset as markdown suitable for
// inclusion in the AI analysis prompt, following the format from PRD section 4.3.
func FormatABDataForPrompt(videos []VideoABData) string {
	if len(videos) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("## Data Legend\n")
	sb.WriteString("- **A/B test share**: Watch-time share percentage per title variant. Higher share = that title kept viewers watching longer vs other variants in the same test. This is the primary quality signal.\n")
	sb.WriteString("- **First-week metrics** (days 0-7 after publish, eliminates age bias):\n")
	sb.WriteString("  - **views**: Total views in first week\n")
	sb.WriteString("  - **likes**: Total likes in first week\n")
	sb.WriteString("  - **comments**: Total comments in first week\n")
	sb.WriteString("  - **engagement**: (likes + comments) / views × 100\n")
	sb.WriteString("\n## A/B Test Results\n\n")

	for _, v := range videos {
		sb.WriteString(fmt.Sprintf("### Video: %s | %s\n", v.Category, v.DayOfWeek))

		if v.HasAnalytics {
			sb.WriteString(fmt.Sprintf("First-week: views=%d | likes=%d | comments=%d | engagement=%.1f%%\n",
				v.FirstWeekViews, v.FirstWeekLikes, v.FirstWeekComments, v.FirstWeekEngagement))
		} else {
			sb.WriteString("First-week: (analytics unavailable)\n")
		}

		sb.WriteString("Titles:\n")
		for _, t := range v.Titles {
			sb.WriteString(fmt.Sprintf("- %q (share: %.1f%%)\n", t.Text, t.Share))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
