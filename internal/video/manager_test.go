package video_test

import (
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// mockFilePathFunc creates a temporary YAML file with the given video data
// and returns the path to it. It handles cleanup of the temp file.
func mockFilePathFunc(t *testing.T, videoData storage.Video, uniqueID string) func(category, name, extension string) string {
	tempDir := t.TempDir()
	return func(category, name, extension string) string {
		// Use uniqueID from test case name to ensure distinct file for each sub-test
		fileName := fmt.Sprintf("%s-%s-%s.%s", category, name, uniqueID, extension)
		filePath := filepath.Join(tempDir, fileName)

		data, errYaml := yaml.Marshal(&videoData)
		if errYaml != nil {
			t.Fatalf("Failed to marshal video data for mock: %v", errYaml)
		}
		errWrite := os.WriteFile(filePath, data, 0644)
		if errWrite != nil {
			t.Fatalf("Failed to write mock video file: %v", errWrite)
		}
		return filePath
	}
}

func TestGetVideoPhase(t *testing.T) {
	testCases := []struct {
		name          string
		videoIndex    storage.VideoIndex
		videoData     storage.Video // Data to be written to the mock YAML file
		expectedPhase int
	}{
		{
			name:          "PhaseDelayed",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "delayed_video"},
			videoData:     storage.Video{Delayed: true},
			expectedPhase: workflow.PhaseDelayed,
		},
		{
			name:          "PhaseSponsoredBlocked",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "blocked_video"},
			videoData:     storage.Video{Sponsorship: storage.Sponsorship{Blocked: "Waiting for sponsor"}},
			expectedPhase: workflow.PhaseSponsoredBlocked,
		},
		{
			name:          "PhasePublished",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "published_video"},
			videoData:     storage.Video{Repo: "github.com/some/repo"},
			expectedPhase: workflow.PhasePublished,
		},
		{
			name:          "PhasePublishPending",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "pending_video"},
			videoData:     storage.Video{UploadVideo: "youtube.com/id", Tweet: "Check out my new video!"},
			expectedPhase: workflow.PhasePublishPending,
		},
		{
			name:          "PhaseEditRequested",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "edit_req_video"},
			videoData:     storage.Video{RequestEdit: true},
			expectedPhase: workflow.PhaseEditRequested,
		},
		{
			name:          "PhaseMaterialDone",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "material_done_video"},
			videoData:     storage.Video{Code: true, Screen: true, Head: true, Diagrams: true},
			expectedPhase: workflow.PhaseMaterialDone,
		},
		{
			name:          "PhaseStarted",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "started_video"},
			videoData:     storage.Video{Date: "2023-01-01T10:00"},
			expectedPhase: workflow.PhaseStarted,
		},
		{
			name:          "PhaseIdeas_Default",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "idea_video"},
			videoData:     storage.Video{}, // Empty video
			expectedPhase: workflow.PhaseIdeas,
		},
		{
			name:          "Order_SponsoredBlocked_Overrides_Delayed",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "blocked_vs_delayed"},
			videoData:     storage.Video{Delayed: true, Sponsorship: storage.Sponsorship{Blocked: "Sponsor issue"}},
			expectedPhase: workflow.PhaseSponsoredBlocked,
		},
		{
			name:          "Order_Published_Overrides_Pending",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "published_vs_pending"},
			videoData:     storage.Video{Repo: "github.com/x/y", UploadVideo: "youtube.com/id", Tweet: "Tweet!"},
			expectedPhase: workflow.PhasePublished,
		},
		{
			name:          "Order_EditRequested_Overrides_MaterialDone",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "edit_vs_material"},
			videoData:     storage.Video{RequestEdit: true, Code: true, Screen: true, Head: true, Diagrams: true},
			expectedPhase: workflow.PhaseEditRequested,
		},
		{
			name:          "Order_MaterialDone_Overrides_Started",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "material_vs_started"},
			videoData:     storage.Video{Code: true, Screen: true, Head: true, Diagrams: true, Date: "2023-01-01T10:00"},
			expectedPhase: workflow.PhaseMaterialDone,
		},
		{
			name:          "Order_Started_Overrides_Ideas",
			videoIndex:    storage.VideoIndex{Category: "testcat", Name: "started_vs_ideas"},
			videoData:     storage.Video{Date: "2023-01-01T12:00"}, // Only date set
			expectedPhase: workflow.PhaseStarted,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePathFunc := mockFilePathFunc(t, tc.videoData, tc.name)
			manager := video.NewManager(filePathFunc)
			phase := manager.GetVideoPhase(tc.videoIndex)
			assert.Equal(t, tc.expectedPhase, phase, fmt.Sprintf("Test case %s failed", tc.name))
		})
	}
}

func TestCalculateDefinePhaseCompletion(t *testing.T) {
	manager := video.NewManager(nil) // filePathFunc is not needed for this method

	testCases := []struct {
		name              string
		video             storage.Video
		expectedCompleted int
		expectedTotal     int
	}{
		{
			name:              "All tasks incomplete",
			video:             storage.Video{},
			expectedCompleted: 0,
			expectedTotal:     9, // Title, Description, Highlight, Tags, DescriptionTags, Tweet, Animations, RequestThumbnail, Gist
		},
		{
			name: "Some tasks complete",
			video: storage.Video{
				Title:            "Test Title",
				Description:      "Test Description",
				Tweet:            "A tweet",
				RequestThumbnail: true,
			},
			expectedCompleted: 4,
			expectedTotal:     9,
		},
		{
			name: "All tasks (excluding Gist for this case) complete", // Updated name for clarity
			video: storage.Video{
				Title:            "Complete Title",
				Description:      "Complete Description",
				Highlight:        "Complete Highlight",
				Tags:             "tag1,tag2",
				DescriptionTags:  "desc_tag1",
				Tweet:            "Final Tweet",
				Animations:       "Script for animations",
				RequestThumbnail: true,
			},
			expectedCompleted: 8, // Gist is not set here
			expectedTotal:     9,
		},
		{
			name: "All 9 tasks complete (including Gist)",
			video: storage.Video{
				Title:            "Complete Title",
				Description:      "Complete Description",
				Highlight:        "Complete Highlight",
				Tags:             "tag1,tag2",
				DescriptionTags:  "desc_tag1",
				Tweet:            "Final Tweet",
				Animations:       "Script for animations",
				RequestThumbnail: true,
				Gist:             "path/to/my/gist.md",
			},
			expectedCompleted: 9,
			expectedTotal:     9,
		},
		{
			name: "Edge case - empty strings not counted",
			video: storage.Video{
				Title:            "",  // Empty
				Description:      "-", // Dash, not counted
				Highlight:        "Real Highlight",
				Tags:             "",
				DescriptionTags:  "",
				Tweet:            "",
				Animations:       "",
				RequestThumbnail: false, // Boolean false
				Gist:             "",    // Empty Gist
			},
			expectedCompleted: 1, // Only Highlight
			expectedTotal:     9,
		},
		{
			name: "Edge case - string with only spaces not counted, Gist complete",
			video: storage.Video{
				Title:            "   ", // Spaces only
				Description:      "Valid Description",
				RequestThumbnail: true,
				Gist:             "  valid/gist/path.md  ", // Will be trimmed and counted
			},
			expectedCompleted: 3, // Description, RequestThumbnail, Gist
			expectedTotal:     9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completed, total := manager.CalculateDefinePhaseCompletion(tc.video)
			assert.Equal(t, tc.expectedCompleted, completed, "Completed count mismatch")
			assert.Equal(t, tc.expectedTotal, total, "Total count mismatch")
		})
	}
}
