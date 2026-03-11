package ai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/storage"

	"gopkg.in/yaml.v3"
)

func TestHasABData(t *testing.T) {
	tests := []struct {
		name   string
		titles []storage.TitleVariant
		want   bool
	}{
		{
			name:   "nil titles",
			titles: nil,
			want:   false,
		},
		{
			name:   "empty titles",
			titles: []storage.TitleVariant{},
			want:   false,
		},
		{
			name: "single title",
			titles: []storage.TitleVariant{
				{Index: 1, Text: "Only Title", Share: 100},
			},
			want: false,
		},
		{
			name: "two titles both zero share",
			titles: []storage.TitleVariant{
				{Index: 1, Text: "Title A", Share: 0},
				{Index: 2, Text: "Title B", Share: 0},
			},
			want: false,
		},
		{
			name: "two titles one with share",
			titles: []storage.TitleVariant{
				{Index: 1, Text: "Title A", Share: 55.2},
				{Index: 2, Text: "Title B", Share: 0},
			},
			want: true,
		},
		{
			name: "three titles all with shares",
			titles: []storage.TitleVariant{
				{Index: 1, Text: "Title A", Share: 42.1},
				{Index: 2, Text: "Title B", Share: 35.5},
				{Index: 3, Text: "Title C", Share: 22.4},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasABData(tt.titles)
			if got != tt.want {
				t.Errorf("HasABData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadIndex(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		index, err := readIndex("/nonexistent/path/index.yaml")
		if err != nil {
			t.Fatalf("expected no error for missing file, got: %v", err)
		}
		if index != nil {
			t.Errorf("expected nil index for missing file, got: %v", index)
		}
	})

	t.Run("valid index file", func(t *testing.T) {
		dir := t.TempDir()
		indexPath := filepath.Join(dir, "index.yaml")
		entries := []storage.VideoIndex{
			{Name: "video-one", Category: "ai"},
			{Name: "video-two", Category: "kubernetes"},
		}
		data, _ := yaml.Marshal(entries)
		os.WriteFile(indexPath, data, 0644)

		index, err := readIndex(indexPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(index) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(index))
		}
		if index[0].Name != "video-one" || index[0].Category != "ai" {
			t.Errorf("unexpected first entry: %+v", index[0])
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		indexPath := filepath.Join(dir, "index.yaml")
		os.WriteFile(indexPath, []byte("not: valid: yaml: ["), 0644)

		_, err := readIndex(indexPath)
		if err == nil {
			t.Fatal("expected error for invalid yaml")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		indexPath := filepath.Join(dir, "index.yaml")
		os.WriteFile(indexPath, []byte(""), 0644)

		index, err := readIndex(indexPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(index) != 0 {
			t.Errorf("expected empty index, got %d entries", len(index))
		}
	})
}

func TestReadVideoFile(t *testing.T) {
	t.Run("valid video file", func(t *testing.T) {
		dir := t.TempDir()
		videoPath := filepath.Join(dir, "test-video.yaml")
		video := storage.Video{
			VideoId: "abc123",
			Date:    "2025-06-15T10:00",
			Titles: []storage.TitleVariant{
				{Index: 1, Text: "Title One", Share: 55.0},
				{Index: 2, Text: "Title Two", Share: 45.0},
			},
		}
		data, _ := yaml.Marshal(video)
		os.WriteFile(videoPath, data, 0644)

		got, err := readVideoFile(videoPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.VideoId != "abc123" {
			t.Errorf("expected videoId abc123, got %s", got.VideoId)
		}
		if len(got.Titles) != 2 {
			t.Errorf("expected 2 titles, got %d", len(got.Titles))
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := readVideoFile("/nonexistent/video.yaml")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

// setupTestData creates a temporary directory structure with index files and video YAML files.
func setupTestData(t *testing.T, videos map[string]storage.Video, currentIndex, archiveIndex []storage.VideoIndex) (indexPath, dataDir, manuscriptDir string) {
	t.Helper()
	dataDir = t.TempDir()
	manuscriptDir = filepath.Join(dataDir, "manuscript")

	// Write current index
	indexPath = filepath.Join(dataDir, "index.yaml")
	if currentIndex != nil {
		data, _ := yaml.Marshal(currentIndex)
		os.WriteFile(indexPath, data, 0644)
	}

	// Write archive index
	if archiveIndex != nil {
		archiveDir := filepath.Join(dataDir, "index")
		os.MkdirAll(archiveDir, 0755)
		data, _ := yaml.Marshal(archiveIndex)
		previousYear := time.Now().Year() - 1
		archivePath := filepath.Join(archiveDir, fmt.Sprintf("%d.yaml", previousYear))
		os.WriteFile(archivePath, data, 0644)
	}

	// Write video files
	for key, video := range videos {
		catAndName := splitCategoryName(key)
		dir := filepath.Join(manuscriptDir, catAndName[0])
		os.MkdirAll(dir, 0755)
		data, _ := yaml.Marshal(video)
		os.WriteFile(filepath.Join(dir, catAndName[1]+".yaml"), data, 0644)
	}

	return indexPath, dataDir, manuscriptDir
}

// splitCategoryName splits "category/name" into [category, name].
// Fallback: returns {key, key} if no '/' found (shouldn't happen with valid index entries).
func splitCategoryName(key string) [2]string {
	for i, c := range key {
		if c == '/' {
			return [2]string{key[:i], key[i+1:]}
		}
	}
	return [2]string{key, key}
}

func TestLoadVideosWithABData(t *testing.T) {
	t.Run("filters to only videos with AB data and videoId", func(t *testing.T) {
		videos := map[string]storage.Video{
			"ai/video-with-ab": {
				VideoId: "vid1",
				Date:    "2026-01-15T10:00",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title A", Share: 55.0},
					{Index: 2, Text: "Title B", Share: 45.0},
				},
			},
			"ai/video-no-ab": {
				VideoId: "vid2",
				Date:    "2026-02-10T10:00",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Only Title"},
				},
			},
			"ai/video-no-videoid": {
				Date: "2026-03-01T10:00",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title X", Share: 60.0},
					{Index: 2, Text: "Title Y", Share: 40.0},
				},
			},
			"kubernetes/video-ab-zero-shares": {
				VideoId: "vid4",
				Date:    "2026-01-20T14:00",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title M", Share: 0},
					{Index: 2, Text: "Title N", Share: 0},
				},
			},
		}

		currentIndex := []storage.VideoIndex{
			{Name: "video-with-ab", Category: "ai"},
			{Name: "video-no-ab", Category: "ai"},
			{Name: "video-no-videoid", Category: "ai"},
			{Name: "video-ab-zero-shares", Category: "kubernetes"},
		}

		indexPath, dataDir, manuscriptDir := setupTestData(t, videos, currentIndex, nil)

		result, err := LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 video with AB data, got %d", len(result))
		}

		if result[0].VideoID != "vid1" {
			t.Errorf("expected videoId vid1, got %s", result[0].VideoID)
		}
		if result[0].Category != "ai" {
			t.Errorf("expected category ai, got %s", result[0].Category)
		}
		if result[0].DayOfWeek != "Thursday" {
			t.Errorf("expected Thursday, got %s", result[0].DayOfWeek)
		}
	})

	t.Run("combines current and archive indexes", func(t *testing.T) {
		videos := map[string]storage.Video{
			"ai/current-video": {
				VideoId: "vid1",
				Date:    "2026-02-10T10:00",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Current A", Share: 52.0},
					{Index: 2, Text: "Current B", Share: 48.0},
				},
			},
			"ai/archive-video": {
				VideoId: "vid2",
				Date:    "2025-06-15T14:00",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Archive A", Share: 60.0},
					{Index: 2, Text: "Archive B", Share: 40.0},
				},
			},
		}

		currentIndex := []storage.VideoIndex{
			{Name: "current-video", Category: "ai"},
		}
		archiveIndex := []storage.VideoIndex{
			{Name: "archive-video", Category: "ai"},
		}

		indexPath, dataDir, manuscriptDir := setupTestData(t, videos, currentIndex, archiveIndex)

		result, err := LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("expected 2 videos, got %d", len(result))
		}
	})

	t.Run("deduplicates entries across indexes", func(t *testing.T) {
		videos := map[string]storage.Video{
			"ai/shared-video": {
				VideoId: "vid1",
				Date:    "2026-01-10T10:00",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title A", Share: 55.0},
					{Index: 2, Text: "Title B", Share: 45.0},
				},
			},
		}

		entry := storage.VideoIndex{Name: "shared-video", Category: "ai"}
		indexPath, dataDir, manuscriptDir := setupTestData(t, videos, []storage.VideoIndex{entry}, []storage.VideoIndex{entry})

		result, err := LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("expected 1 video (deduplicated), got %d", len(result))
		}
	})

	t.Run("missing current index returns error", func(t *testing.T) {
		dir := t.TempDir()
		// Create an unreadable file to trigger an error (not just missing)
		indexPath := filepath.Join(dir, "index.yaml")
		os.WriteFile(indexPath, []byte("not: valid: ["), 0644)

		_, err := LoadVideosWithABData(indexPath, dir, filepath.Join(dir, "manuscript"))
		if err == nil {
			t.Fatal("expected error for invalid index")
		}
	})

	t.Run("missing archive index is not an error", func(t *testing.T) {
		videos := map[string]storage.Video{
			"ai/video-one": {
				VideoId: "vid1",
				Date:    "2026-02-01T10:00",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "A", Share: 50},
					{Index: 2, Text: "B", Share: 50},
				},
			},
		}

		currentIndex := []storage.VideoIndex{
			{Name: "video-one", Category: "ai"},
		}

		indexPath, dataDir, manuscriptDir := setupTestData(t, videos, currentIndex, nil)

		result, err := LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 video, got %d", len(result))
		}
	})

	t.Run("skips videos with unreadable files", func(t *testing.T) {
		currentIndex := []storage.VideoIndex{
			{Name: "missing-video", Category: "ai"},
		}

		indexPath, dataDir, manuscriptDir := setupTestData(t, nil, currentIndex, nil)

		result, err := LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 videos, got %d", len(result))
		}
	})

	t.Run("empty indexes return empty result", func(t *testing.T) {
		indexPath, dataDir, manuscriptDir := setupTestData(t, nil, []storage.VideoIndex{}, nil)

		result, err := LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected 0 videos, got %d", len(result))
		}
	})

	t.Run("parses date-only format", func(t *testing.T) {
		videos := map[string]storage.Video{
			"ai/date-only": {
				VideoId: "vid1",
				Date:    "2026-01-06",
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "A", Share: 50},
					{Index: 2, Text: "B", Share: 50},
				},
			},
		}

		currentIndex := []storage.VideoIndex{
			{Name: "date-only", Category: "ai"},
		}

		indexPath, dataDir, manuscriptDir := setupTestData(t, videos, currentIndex, nil)

		result, err := LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 video, got %d", len(result))
		}
		if result[0].DayOfWeek != "Tuesday" {
			t.Errorf("expected Tuesday for 2026-01-06, got %s", result[0].DayOfWeek)
		}
	})
}

func TestEnrichWithAnalytics(t *testing.T) {
	t.Run("matches by video ID", func(t *testing.T) {
		videos := []VideoABData{
			{VideoID: "vid1", Category: "ai"},
			{VideoID: "vid2", Category: "kubernetes"},
			{VideoID: "vid3", Category: "devops"},
		}

		analytics := []publishing.VideoAnalytics{
			{
				VideoID:           "vid1",
				FirstWeekViews:    15000,
				FirstWeekCTR:      8.2,
				FirstWeekLikes:    890,
				FirstWeekComments: 145,
			},
			{
				VideoID:           "vid3",
				FirstWeekViews:    5000,
				FirstWeekCTR:      5.1,
				FirstWeekLikes:    200,
				FirstWeekComments: 30,
			},
		}

		result := EnrichWithAnalytics(videos, analytics)

		if len(result) != 3 {
			t.Fatalf("expected 3 videos, got %d", len(result))
		}

		// vid1 should have analytics
		if !result[0].HasAnalytics {
			t.Error("vid1 should have analytics")
		}
		if result[0].FirstWeekViews != 15000 {
			t.Errorf("vid1 FirstWeekViews = %d, want 15000", result[0].FirstWeekViews)
		}
		// vid2 should NOT have analytics
		if result[1].HasAnalytics {
			t.Error("vid2 should not have analytics")
		}

		// vid3 should have analytics with computed engagement
		if !result[2].HasAnalytics {
			t.Error("vid3 should have analytics")
		}
		expectedEngagement := float64(200+30) / float64(5000) * 100 // 4.6%
		if result[2].FirstWeekEngagement != expectedEngagement {
			t.Errorf("vid3 engagement = %f, want %f", result[2].FirstWeekEngagement, expectedEngagement)
		}
	})

	t.Run("empty analytics", func(t *testing.T) {
		videos := []VideoABData{
			{VideoID: "vid1"},
		}

		result := EnrichWithAnalytics(videos, nil)
		if len(result) != 1 {
			t.Fatalf("expected 1 video, got %d", len(result))
		}
		if result[0].HasAnalytics {
			t.Error("should not have analytics")
		}
	})

	t.Run("zero views gives zero engagement", func(t *testing.T) {
		videos := []VideoABData{
			{VideoID: "vid1"},
		}
		analytics := []publishing.VideoAnalytics{
			{VideoID: "vid1", FirstWeekViews: 0, FirstWeekLikes: 10},
		}

		result := EnrichWithAnalytics(videos, analytics)
		if result[0].FirstWeekEngagement != 0 {
			t.Errorf("expected 0 engagement for zero views, got %f", result[0].FirstWeekEngagement)
		}
	})

	t.Run("does not mutate original slice", func(t *testing.T) {
		videos := []VideoABData{
			{VideoID: "vid1"},
		}
		analytics := []publishing.VideoAnalytics{
			{VideoID: "vid1", FirstWeekViews: 1000, FirstWeekLikes: 50},
		}

		result := EnrichWithAnalytics(videos, analytics)
		if videos[0].HasAnalytics {
			t.Error("original slice should not be mutated")
		}
		if !result[0].HasAnalytics {
			t.Error("result should have analytics")
		}
	})
}

func TestFormatABDataForPrompt(t *testing.T) {
	t.Run("empty videos returns empty string", func(t *testing.T) {
		result := FormatABDataForPrompt(nil)
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("formats video with analytics", func(t *testing.T) {
		videos := []VideoABData{
			{
				Category:            "ai",
				DayOfWeek:           "Monday",
				HasAnalytics:        true,
				FirstWeekViews:      15230,
				FirstWeekLikes:      890,
				FirstWeekComments:   145,
				FirstWeekEngagement: 6.8,
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Why I Changed My Mind About Cursor", Share: 42.1},
					{Index: 2, Text: "Top 10 AI Coding Tools in 2025", Share: 35.5},
					{Index: 3, Text: "AI Coding Is Broken (Here's the Fix)", Share: 22.4},
				},
			},
		}

		result := FormatABDataForPrompt(videos)

		// Check legend is present
		if !strings.Contains(result, "## Data Legend") {
			t.Error("missing Data Legend section")
		}
		if !strings.Contains(result, "A/B test share") {
			t.Error("missing A/B test share explanation")
		}

		// Check video header
		if !strings.Contains(result, "### Video: ai | Monday") {
			t.Error("missing video header")
		}

		// Check first-week metrics
		if !strings.Contains(result, "views=15230") {
			t.Error("missing views")
		}
		// Check titles
		if !strings.Contains(result, `"Why I Changed My Mind About Cursor" (share: 42.1%)`) {
			t.Error("missing first title with share")
		}
		if !strings.Contains(result, `"AI Coding Is Broken (Here's the Fix)" (share: 22.4%)`) {
			t.Error("missing third title with share")
		}
	})

	t.Run("formats video without analytics", func(t *testing.T) {
		videos := []VideoABData{
			{
				Category:     "kubernetes",
				DayOfWeek:    "Thursday",
				HasAnalytics: false,
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Stop Using Helm Charts!", Share: 51.2},
					{Index: 2, Text: "Why Helm Is Dead in 2025", Share: 48.8},
				},
			},
		}

		result := FormatABDataForPrompt(videos)

		if !strings.Contains(result, "(analytics unavailable)") {
			t.Error("missing analytics unavailable marker")
		}
		if !strings.Contains(result, `"Stop Using Helm Charts!" (share: 51.2%)`) {
			t.Error("missing title")
		}
	})

	t.Run("formats multiple videos", func(t *testing.T) {
		videos := []VideoABData{
			{
				Category:     "ai",
				DayOfWeek:    "Monday",
				HasAnalytics: true,
				FirstWeekViews: 1000,
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title A", Share: 60},
					{Index: 2, Text: "Title B", Share: 40},
				},
			},
			{
				Category:     "devops",
				DayOfWeek:    "Friday",
				HasAnalytics: true,
				FirstWeekViews: 2000,
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title C", Share: 55},
					{Index: 2, Text: "Title D", Share: 45},
				},
			},
		}

		result := FormatABDataForPrompt(videos)

		if !strings.Contains(result, "### Video: ai | Monday") {
			t.Error("missing first video header")
		}
		if !strings.Contains(result, "### Video: devops | Friday") {
			t.Error("missing second video header")
		}
	})
}

