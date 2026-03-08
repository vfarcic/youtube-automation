package video_test

import (
	"devopstoolkit/youtube-automation/internal/aspect"
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
			manager := video.NewManager(filePathFunc, nil)
			phase := manager.GetVideoPhase(tc.videoIndex)
			assert.Equal(t, tc.expectedPhase, phase, fmt.Sprintf("Test case %s failed", tc.name))
		})
	}
}

func TestCalculateDefinePhaseCompletion(t *testing.T) {
	manager := video.NewManager(nil, aspect.NewService()) // filePathFunc is not needed for this method

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
			expectedTotal:     10, // Titles, Description, Tags, DescriptionTags, Tweet, Animations, Shorts, Members, RequestThumbnail, RequestEdit
		},
		{
			name: "Some tasks complete",
			video: storage.Video{
				Titles:           []storage.TitleVariant{{Index: 1, Text: "Test Title"}},
				Description:      "Test Description",
				Tweet:            "A tweet",
				RequestThumbnail: true,
			},
			expectedCompleted: 4,
			expectedTotal:     10,
		},
		{
			name: "All Definition tasks complete",
			video: storage.Video{
				Titles:           []storage.TitleVariant{{Index: 1, Text: "Complete Title"}},
				Description:      "Complete Description",
				Tags:             "tag1,tag2",
				DescriptionTags:  "desc_tag1",
				Tweet:            "Final Tweet",
				Animations:       "Script for animations",
				Shorts:           []storage.Short{{ID: "short1", Title: "Short"}},
				Members:          "member1",
				RequestThumbnail: true,
				RequestEdit:      true,
			},
			expectedCompleted: 10, // All 10 Definition fields complete
			expectedTotal:     10,
		},
		{
			name: "All Definition tasks complete with Gist (Gist should not affect Definition count)",
			video: storage.Video{
				Titles:           []storage.TitleVariant{{Index: 1, Text: "Complete Title"}},
				Description:      "Complete Description",
				Tags:             "tag1,tag2",
				DescriptionTags:  "desc_tag1",
				Tweet:            "Final Tweet",
				Animations:       "Script for animations",
				Shorts:           []storage.Short{{ID: "short1", Title: "Short"}},
				Members:          "member1",
				RequestThumbnail: true,
				RequestEdit:      true,
				Gist:             "path/to/my/gist.md", // This should NOT affect Definition phase count
			},
			expectedCompleted: 10,
			expectedTotal:     10,
		},
		{
			name: "Edge case - empty strings not counted",
			video: storage.Video{
				Titles:           []storage.TitleVariant{}, // Empty array
				Description:      "-",                      // Dash, not counted
				Tags:             "",
				DescriptionTags:  "",
				Tweet:            "",
				Animations:       "",
				RequestThumbnail: false, // Boolean false
				Gist:             "",    // Empty Gist (but doesn't matter for Definition phase)
			},
			expectedCompleted: 0,
			expectedTotal:     10,
		},
		{
			name: "Edge case - string with only spaces not counted",
			video: storage.Video{
				Titles:           []storage.TitleVariant{{Index: 1, Text: "   "}}, // Spaces only
				Description:      "Valid Description",
				RequestThumbnail: true,
				Gist:             "  valid/gist/path.md  ", // Gist doesn't affect Definition phase count
			},
			expectedCompleted: 2, // Description, RequestThumbnail (title with spaces not counted)
			expectedTotal:     10,
		},
		{
			name: "Multiple titles - at least one valid counts as complete",
			video: storage.Video{
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Valid Title"},
					{Index: 2, Text: "Another Title"},
					{Index: 3, Text: "Third Title"},
				},
				Description:      "Test Description",
				RequestThumbnail: true,
			},
			expectedCompleted: 3, // Titles (1), Description (1), RequestThumbnail (1)
			expectedTotal:     10,
		},
		{
			name: "Titles array with empty text not counted",
			video: storage.Video{
				Titles: []storage.TitleVariant{
					{Index: 1, Text: ""},
					{Index: 2, Text: "-"},
				},
				Description: "Test Description",
			},
			expectedCompleted: 1, // Only Description
			expectedTotal:     10,
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

func TestCalculateVideoPhase(t *testing.T) {
	testCases := []struct {
		name          string
		videoData     storage.Video
		expectedPhase int
		description   string
	}{
		{
			name:          "PhaseDelayed",
			videoData:     storage.Video{Delayed: true},
			expectedPhase: workflow.PhaseDelayed,
			description:   "Video marked as delayed should return PhaseDelayed",
		},
		{
			name:          "PhaseSponsoredBlocked",
			videoData:     storage.Video{Sponsorship: storage.Sponsorship{Blocked: "Waiting for sponsor"}},
			expectedPhase: workflow.PhaseSponsoredBlocked,
			description:   "Video with sponsorship block should return PhaseSponsoredBlocked",
		},
		{
			name:          "PhasePublished",
			videoData:     storage.Video{Repo: "github.com/some/repo"},
			expectedPhase: workflow.PhasePublished,
			description:   "Video with repository should return PhasePublished",
		},
		{
			name:          "PhasePublishPending_UploadAndTweet",
			videoData:     storage.Video{UploadVideo: "youtube.com/id", Tweet: "Check out my new video!"},
			expectedPhase: workflow.PhasePublishPending,
			description:   "Video with upload and tweet should return PhasePublishPending",
		},
		{
			name:          "PhaseEditRequested",
			videoData:     storage.Video{RequestEdit: true},
			expectedPhase: workflow.PhaseEditRequested,
			description:   "Video with edit request should return PhaseEditRequested",
		},
		{
			name:          "PhaseMaterialDone_AllMaterials",
			videoData:     storage.Video{Code: true, Screen: true, Head: true, Diagrams: true},
			expectedPhase: workflow.PhaseMaterialDone,
			description:   "Video with all materials should return PhaseMaterialDone",
		},
		{
			name:          "PhaseStarted",
			videoData:     storage.Video{Date: "2023-01-01T10:00"},
			expectedPhase: workflow.PhaseStarted,
			description:   "Video with date should return PhaseStarted",
		},
		{
			name:          "PhaseIdeas_Default",
			videoData:     storage.Video{},
			expectedPhase: workflow.PhaseIdeas,
			description:   "Empty video should return PhaseIdeas (default)",
		},
		// Priority order tests - higher priority phases should override lower ones
		{
			name: "Priority_SponsoredBlocked_Over_Delayed",
			videoData: storage.Video{
				Delayed:     true,
				Sponsorship: storage.Sponsorship{Blocked: "Sponsor issue"},
			},
			expectedPhase: workflow.PhaseSponsoredBlocked,
			description:   "Sponsored blocked should override delayed",
		},
		{
			name: "Priority_Published_Over_Pending",
			videoData: storage.Video{
				Repo:        "github.com/x/y",
				UploadVideo: "youtube.com/id",
				Tweet:       "Tweet!",
			},
			expectedPhase: workflow.PhasePublished,
			description:   "Published should override publish pending",
		},
		{
			name: "Priority_EditRequested_Over_MaterialDone",
			videoData: storage.Video{
				RequestEdit: true,
				Code:        true,
				Screen:      true,
				Head:        true,
				Diagrams:    true,
			},
			expectedPhase: workflow.PhaseEditRequested,
			description:   "Edit requested should override material done",
		},
		{
			name: "Priority_MaterialDone_Over_Started",
			videoData: storage.Video{
				Code:     true,
				Screen:   true,
				Head:     true,
				Diagrams: true,
				Date:     "2023-01-01T10:00",
			},
			expectedPhase: workflow.PhaseMaterialDone,
			description:   "Material done should override started",
		},
		{
			name: "Priority_Started_Over_Ideas",
			videoData: storage.Video{
				Date: "2023-01-01T12:00",
			},
			expectedPhase: workflow.PhaseStarted,
			description:   "Started should override ideas",
		},
		// Edge cases
		{
			name: "PhaseMaterialDone_Partial_Materials",
			videoData: storage.Video{
				Code:   true,
				Screen: true,
				// Missing Head and Diagrams
			},
			expectedPhase: workflow.PhaseIdeas, // Should be PhaseIdeas (7) since materials aren't complete
			description:   "Partial materials should not trigger PhaseMaterialDone, should return PhaseIdeas",
		},
		{
			name: "PhasePublishPending_OnlyUpload",
			videoData: storage.Video{
				UploadVideo: "youtube.com/id",
				// Missing Tweet
			},
			expectedPhase: workflow.PhaseIdeas, // Should be PhaseIdeas (7) since both conditions aren't met
			description:   "Only upload video without tweet should not trigger PhasePublishPending, should return PhaseIdeas",
		},
		{
			name: "PhasePublishPending_OnlyTweet",
			videoData: storage.Video{
				Tweet: "Check out my video!",
				// Missing UploadVideo
			},
			expectedPhase: workflow.PhaseIdeas, // Should be PhaseIdeas (7) since both conditions aren't met
			description:   "Only tweet without upload video should not trigger PhasePublishPending, should return PhaseIdeas",
		},
		{
			name: "PhaseSponsoredBlocked_EmptyBlocked",
			videoData: storage.Video{
				Sponsorship: storage.Sponsorship{Blocked: ""},
			},
			expectedPhase: workflow.PhaseIdeas,
			description:   "Empty blocked string should not trigger PhaseSponsoredBlocked",
		},
		{
			name: "PhasePublished_EmptyRepo",
			videoData: storage.Video{
				Repo: "",
			},
			expectedPhase: workflow.PhaseIdeas,
			description:   "Empty repo should not trigger PhasePublished",
		},
		// Complex scenario tests
		{
			name: "Complex_All_Flags_Set_Priority_Test",
			videoData: storage.Video{
				Delayed:     true,
				Sponsorship: storage.Sponsorship{Blocked: "Sponsor issue"},
				Repo:        "github.com/test/repo",
				UploadVideo: "youtube.com/video",
				Tweet:       "Tweet content",
				RequestEdit: true,
				Code:        true,
				Screen:      true,
				Head:        true,
				Diagrams:    true,
				Date:        "2023-01-01T10:00",
			},
			expectedPhase: workflow.PhaseSponsoredBlocked, // Highest priority
			description:   "When all conditions are met, highest priority (SponsoredBlocked) should win",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			phase := video.CalculateVideoPhase(tc.videoData)
			assert.Equal(t, tc.expectedPhase, phase,
				fmt.Sprintf("Test case %s failed: %s. Expected phase %d, got %d",
					tc.name, tc.description, tc.expectedPhase, phase))
		})
	}
}

func TestCalculateOverallProgress(t *testing.T) {
	manager := video.NewManager(nil, aspect.NewService()) // filePathFunc not needed for progress calculations

	testCases := []struct {
		name              string
		video             storage.Video
		expectedCompleted int
		expectedTotal     int
		description       string
	}{
		{
			name:              "Empty_video",
			video:             storage.Video{},
			expectedCompleted: -1, // Will be verified dynamically
			expectedTotal:     -1,
			description:       "Empty video should have minimal completed tasks",
		},
		{
			name: "Partially_complete_video",
			video: storage.Video{
				// Initial Details phase
				ProjectName: "Test Project",
				Date:        "2023-01-01",

				// Work Progress phase
				Code:   true,
				Screen: true,

				// Definition phase
				Titles:      []storage.TitleVariant{{Index: 1, Text: "Test Title"}},
				Description: "Test Description",

				// Post-Production phase
				VideoFile: "video.mp4",

				// Publishing phase
				UploadVideo: "youtube.com/video",

				// Post-Publish phase
				SlackPosted: true,
			},
			expectedCompleted: -1,
			expectedTotal:     -1,
			description:       "Partially complete video should count all completed fields",
		},
		{
			name: "Fully_complete_video",
			video: storage.Video{
				// Initial Details
				ProjectName: "Complete Project",
				ProjectURL:  "https://example.com",
				Gist:        "path/to/gist",
				Date:        "2023-01-01",
				Sponsorship: storage.Sponsorship{Amount: "1000"},

				// Work Progress
				Code:          true,
				Head:          true,
				Screen:        true,
				RelatedVideos: "video1,video2",
				Thumbnails:    true,
				Diagrams:      true,
				Screenshots:   true,
				Location:      "Conference Room",
				Tagline:       "Great tagline",
				TaglineIdeas:  "Some ideas",
				OtherLogos:    "some_logo.png",

				// Definition
				Titles:           []storage.TitleVariant{{Index: 1, Text: "Complete Title"}},
				Description:      "Complete Description",
				Tags:             "tag1,tag2",
				DescriptionTags:  "desc_tag1",
				Tweet:            "Final Tweet",
				Animations:       "Animation script",
				Shorts:           []storage.Short{{ID: "short1", Title: "Short"}},
				Members:          "member1",
				RequestThumbnail: true,
				RequestEdit:      true,

				// Post-Production
				ThumbnailVariants: []storage.ThumbnailVariant{{Path: "thumb.jpg"}},
				VideoFile:         "video.mp4",
				Slides:            true,
				Timecodes:         "00:00 Intro, 05:00 Main", // No FIXME

				// Publishing
				UploadVideo: "youtube.com/video",
				VideoId:     "abc123",
				HugoPath:    "/path/to/hugo",

				// Post-Publish
				DOTPosted:           true,
				BlueSkyPosted:       true,
				LinkedInPosted:      true,
				SlackPosted:         true,
				YouTubeHighlight:    true,
				YouTubeComment:      true,
				YouTubeCommentReply: true,
				GDE:                 true,
				Repo:                "github.com/repo",
				NotifiedSponsors:    true,
			},
			expectedCompleted: -1,
			expectedTotal:     -1,
			description:       "Fully complete video should have almost all tasks completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completed, total := manager.CalculateOverallProgress(tc.video)

			// Verify overall equals sum of individual phases
			initC, initT := manager.CalculateInitialDetailsProgress(tc.video)
			workC, workT := manager.CalculateWorkProgressProgress(tc.video)
			defineC, defineT := manager.CalculateDefinePhaseCompletion(tc.video)
			editC, editT := manager.CalculatePostProductionProgress(tc.video)
			publishC, publishT := manager.CalculatePublishingProgress(tc.video)
			postPublishC, postPublishT := manager.CalculatePostPublishProgress(tc.video)

			expectedTotal := initT + workT + defineT + editT + publishT + postPublishT
			expectedCompleted := initC + workC + defineC + editC + publishC + postPublishC

			assert.Equal(t, expectedCompleted, completed, "Completed count should equal sum of phases")
			assert.Equal(t, expectedTotal, total, "Total count should equal sum of phases")
			assert.GreaterOrEqual(t, completed, 0, "Completed should be non-negative")
			assert.GreaterOrEqual(t, total, completed, "Total should be >= completed")
		})
	}
}

func TestCalculateInitialDetailsProgress(t *testing.T) {
	manager := video.NewManager(nil, aspect.NewService())

	testCases := []struct {
		name              string
		video             storage.Video
		expectedCompleted int
		expectedTotal     int
		description       string
	}{
		{
			name:              "Empty_video",
			video:             storage.Video{},
			expectedCompleted: 3, // Sponsorship.Emails (conditional, no amount), Sponsorship.Blocked (empty_or_filled), Delayed (false_only)
			expectedTotal:     10,
			description:       "Empty video should have 3 completed conditions",
		},
		{
			name: "All_general_fields_complete",
			video: storage.Video{
				ProjectName: "Test Project",
				ProjectURL:  "https://example.com",
				Gist:        "path/to/gist",
				Date:        "2023-01-01",
			},
			expectedCompleted: 7, // 4 general + emails(conditional) + blocked(empty_or_filled) + delayed(false_only)
			expectedTotal:     10,
			description:       "All general fields should be counted",
		},
		{
			name: "With_sponsorship_amount",
			video: storage.Video{
				Sponsorship: storage.Sponsorship{Amount: "1000"},
			},
			expectedCompleted: 3, // Amount(filled_only) + Blocked(empty_or_filled) + Delayed(false_only)
			expectedTotal:     10,
			description:       "Sponsorship amount should be counted",
		},
		{
			name: "With_sponsorship_emails",
			video: storage.Video{
				Sponsorship: storage.Sponsorship{
					Amount: "1000",
					Emails: "sponsor@example.com",
				},
			},
			expectedCompleted: 4, // Amount + Emails + Blocked(empty_or_filled) + Delayed(false_only)
			expectedTotal:     10,
			description:       "Sponsorship emails should be counted when amount is set",
		},
		{
			name: "With_sponsorship_blocked",
			video: storage.Video{
				Sponsorship: storage.Sponsorship{Blocked: "Some reason"},
			},
			expectedCompleted: 2, // Emails(conditional, no amount=pass) + Delayed(false_only)
			expectedTotal:     10,
			description:       "Sponsorship blocked should fail the blocked condition",
		},
		{
			name: "With_delayed_true",
			video: storage.Video{
				Delayed: true,
			},
			expectedCompleted: 2, // Emails(conditional) + Blocked(empty_or_filled)
			expectedTotal:     10,
			description:       "Delayed video should fail the delayed condition",
		},
		{
			name: "Sponsorship_amount_NA",
			video: storage.Video{
				Sponsorship: storage.Sponsorship{Amount: "N/A"},
			},
			expectedCompleted: 4, // Amount + Emails(conditional, N/A=pass) + Blocked(empty_or_filled) + Delayed(false_only)
			expectedTotal:     10,
			description:       "N/A sponsorship amount should pass emails condition",
		},
		{
			name: "Sponsorship_amount_dash",
			video: storage.Video{
				Sponsorship: storage.Sponsorship{Amount: "-"},
			},
			expectedCompleted: 3, // Emails(conditional, dash=pass) + Blocked(empty_or_filled) + Delayed(false_only); Amount("-") fails filled_only
			expectedTotal:     10,
			description:       "Dash sponsorship amount should pass emails condition",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completed, total := manager.CalculateInitialDetailsProgress(tc.video)
			assert.Equal(t, tc.expectedCompleted, completed, "Completed count mismatch for %s", tc.description)
			assert.Equal(t, tc.expectedTotal, total, "Total count mismatch for %s", tc.description)
		})
	}
}

func TestCalculateWorkProgressProgress(t *testing.T) {
	manager := video.NewManager(nil, aspect.NewService())

	testCases := []struct {
		name              string
		video             storage.Video
		expectedCompleted int
		expectedTotal     int
		description       string
	}{
		{
			name:              "Empty_video",
			video:             storage.Video{},
			expectedCompleted: 0,
			expectedTotal:     11, // Code, Head, Screen, RelatedVideos, Thumbnails, Diagrams, Screenshots, Location, Tagline, TaglineIdeas, OtherLogos
			description:       "Empty video should have no work progress",
		},
		{
			name: "Some_booleans_true",
			video: storage.Video{
				Code:        true,
				Screen:      true,
				Thumbnails:  true,
				Diagrams:    true,
				Screenshots: true,
				OtherLogos:  "logo.png",
			},
			expectedCompleted: 6,
			expectedTotal:     11,
			description:       "Boolean fields should be counted when true",
		},
		{
			name: "Some_strings_and_booleans_filled",
			video: storage.Video{
				RelatedVideos: "video1,video2",
				Location:      "Conference Room",
				Tagline:       "Great tagline",
				TaglineIdeas:  "Some ideas",
				Head:          true,
			},
			expectedCompleted: 5,
			expectedTotal:     11,
			description:       "String and boolean fields should be counted when non-empty/true",
		},
		{
			name: "Mixed_fields",
			video: storage.Video{
				Code:          true,
				Head:          false, // Should not count
				Screen:        true,
				RelatedVideos: "video1",
				Thumbnails:    true,
				Diagrams:      false, // Should not count
				Screenshots:   true,
				Location:      "", // Should not count
				Tagline:       "Great tagline",
				TaglineIdeas:  "-", // Should not count (dash)
				OtherLogos:    "company_logo.png",
			},
			expectedCompleted: 7, // Code, Screen, RelatedVideos, Thumbnails, Screenshots, Tagline, OtherLogos
			expectedTotal:     11,
			description:       "Mixed fields should count only completed ones",
		},
		{
			name: "All_complete",
			video: storage.Video{
				Code:          true,
				Head:          true,
				Screen:        true,
				RelatedVideos: "video1,video2",
				Thumbnails:    true,
				Diagrams:      true,
				Screenshots:   true,
				Location:      "Conference Room",
				Tagline:       "Great tagline",
				TaglineIdeas:  "Some ideas",
				OtherLogos:    "logo_files.png",
			},
			expectedCompleted: 11,
			expectedTotal:     11,
			description:       "All work progress fields complete",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completed, total := manager.CalculateWorkProgressProgress(tc.video)
			assert.Equal(t, tc.expectedCompleted, completed, "Completed count mismatch for %s", tc.description)
			assert.Equal(t, tc.expectedTotal, total, "Total count mismatch for %s", tc.description)
		})
	}
}

func TestCalculatePostProductionProgress(t *testing.T) {
	manager := video.NewManager(nil, aspect.NewService())

	testCases := []struct {
		name              string
		video             storage.Video
		expectedCompleted int
		expectedTotal     int
		description       string
	}{
		{
			name:              "Empty_video",
			video:             storage.Video{},
			expectedCompleted: 0,
			expectedTotal:     4, // ThumbnailVariants, Timecodes, VideoFile, Slides
			description:       "Empty video should have no post-production progress",
		},
		{
			name: "Basic_fields_complete",
			video: storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{{Path: "thumbnail.jpg"}},
				VideoFile:         "video.mp4",
				Slides:            true,
			},
			expectedCompleted: 3,
			expectedTotal:     4,
			description:       "Basic post-production fields should be counted",
		},
		{
			name: "Timecodes_valid",
			video: storage.Video{
				Timecodes: "00:00 Intro, 05:00 Main content",
			},
			expectedCompleted: 1,
			expectedTotal:     4,
			description:       "Valid timecodes should be counted",
		},
		{
			name: "Timecodes_with_FIXME",
			video: storage.Video{
				Timecodes: "00:00 Intro, FIXME: Add more timecodes",
			},
			expectedCompleted: 0,
			expectedTotal:     4,
			description:       "Timecodes with FIXME should not be counted",
		},
		{
			name: "Timecodes_empty",
			video: storage.Video{
				Timecodes: "",
			},
			expectedCompleted: 0,
			expectedTotal:     4,
			description:       "Empty timecodes should not be counted",
		},
		{
			name: "All_complete",
			video: storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{{Path: "thumbnail.jpg"}},
				VideoFile:         "video.mp4",
				Slides:            true,
				Timecodes:         "00:00 Intro, 05:00 Main, 10:00 Conclusion",
			},
			expectedCompleted: 4,
			expectedTotal:     4,
			description:       "All post-production fields complete",
		},
		{
			name: "Mixed_completion",
			video: storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{{Path: "thumbnail.jpg"}},
				VideoFile:         "video.mp4",
				Slides:            false,                  // False, not counted
				Timecodes:         "FIXME: Add timecodes", // Has FIXME, not counted
			},
			expectedCompleted: 2, // Only ThumbnailVariants and VideoFile
			expectedTotal:     4,
			description:       "Mixed completion should count only valid fields",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completed, total := manager.CalculatePostProductionProgress(tc.video)
			assert.Equal(t, tc.expectedCompleted, completed, "Completed count mismatch for %s", tc.description)
			assert.Equal(t, tc.expectedTotal, total, "Total count mismatch for %s", tc.description)
		})
	}
}

func TestCalculatePublishingProgress(t *testing.T) {
	manager := video.NewManager(nil, aspect.NewService())

	testCases := []struct {
		name              string
		video             storage.Video
		expectedCompleted int
		expectedTotal     int
		description       string
	}{
		{
			name:              "Empty_video",
			video:             storage.Video{},
			expectedCompleted: 0,
			expectedTotal:     2, // VideoId, HugoPath (UploadVideo hidden)
			description:       "Empty video should have no publishing progress",
		},
		{
			name: "UploadVideo_only",
			video: storage.Video{
				UploadVideo: "youtube.com/video",
			},
			expectedCompleted: 0,
			expectedTotal:     2,
			description:       "UploadVideo is hidden, should not affect progress",
		},
		{
			name: "HugoPath_only",
			video: storage.Video{
				HugoPath: "/path/to/hugo/post",
			},
			expectedCompleted: 1,
			expectedTotal:     2,
			description:       "Only hugo path should count as 1",
		},
		{
			name: "All_base_complete",
			video: storage.Video{
				UploadVideo: "youtube.com/video",
				VideoId:     "abc123",
				HugoPath:    "/path/to/hugo/post",
			},
			expectedCompleted: 2,
			expectedTotal:     2,
			description:       "All visible base fields complete should count as 2",
		},
		{
			name: "Empty_strings",
			video: storage.Video{
				UploadVideo: "",
				HugoPath:    "",
			},
			expectedCompleted: 0,
			expectedTotal:     2,
			description:       "Empty strings should not be counted",
		},
		{
			name: "Dash_values",
			video: storage.Video{
				UploadVideo: "-",
				HugoPath:    "-",
			},
			expectedCompleted: 0,
			expectedTotal:     2,
			description:       "Dash values should not be counted",
		},
		{
			name: "With_pending_shorts",
			video: storage.Video{
				UploadVideo: "video.mp4",
				HugoPath:    "/path/to/hugo",
				Shorts: []storage.Short{
					{ID: "short1", Title: "Short 1", YouTubeID: ""},
					{ID: "short2", Title: "Short 2", YouTubeID: ""},
				},
			},
			expectedCompleted: 1,
			expectedTotal:     4, // 2 base fields + 2 shorts
			description:       "Pending shorts should add to total but not completed",
		},
		{
			name: "With_uploaded_shorts",
			video: storage.Video{
				UploadVideo: "video.mp4",
				HugoPath:    "/path/to/hugo",
				Shorts: []storage.Short{
					{ID: "short1", Title: "Short 1", YouTubeID: "abc123"},
					{ID: "short2", Title: "Short 2", YouTubeID: "def456"},
				},
			},
			expectedCompleted: 3,
			expectedTotal:     4, // 2 base fields + 2 uploaded shorts
			description:       "Uploaded shorts should count as completed",
		},
		{
			name: "With_mixed_shorts",
			video: storage.Video{
				UploadVideo: "video.mp4",
				HugoPath:    "/path/to/hugo",
				Shorts: []storage.Short{
					{ID: "short1", Title: "Short 1", YouTubeID: "abc123"},
					{ID: "short2", Title: "Short 2", YouTubeID: ""},
					{ID: "short3", Title: "Short 3", YouTubeID: "ghi789"},
				},
			},
			expectedCompleted: 3, // 1 base (HugoPath) + 2 uploaded shorts
			expectedTotal:     5, // 2 base + 3 shorts
			description:       "Mixed shorts should partially count",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completed, total := manager.CalculatePublishingProgress(tc.video)
			assert.Equal(t, tc.expectedCompleted, completed, "Completed count mismatch for %s", tc.description)
			assert.Equal(t, tc.expectedTotal, total, "Total count mismatch for %s", tc.description)
		})
	}
}

func TestCalculatePostPublishProgress(t *testing.T) {
	manager := video.NewManager(nil, aspect.NewService())

	testCases := []struct {
		name              string
		video             storage.Video
		expectedCompleted int
		expectedTotal     int
		description       string
	}{
		{
			name:              "Empty_video",
			video:             storage.Video{},
			expectedCompleted: 1,  // NotifiedSponsors condition passes (no sponsorship amount)
			expectedTotal:     10, // 9 basic fields + 1 NotifiedSponsors condition
			description:       "Empty video should pass NotifiedSponsors condition",
		},
		{
			name: "Basic_booleans_true",
			video: storage.Video{
				DOTPosted:           true,
				BlueSkyPosted:       true,
				LinkedInPosted:      true,
				SlackPosted:         true,
				YouTubeHighlight:    true,
				YouTubeComment:      true,
				YouTubeCommentReply: true,
				GDE:                 true,
			},
			expectedCompleted: 9, // 8 booleans + NotifiedSponsors condition
			expectedTotal:     10,
			description:       "Boolean fields should be counted when true",
		},
		{
			name: "With_repo",
			video: storage.Video{
				Repo: "github.com/user/repo",
			},
			expectedCompleted: 2, // Repo + NotifiedSponsors condition
			expectedTotal:     10,
			description:       "Repo should be counted when non-empty",
		},
		{
			name: "With_sponsorship_amount_needs_notification",
			video: storage.Video{
				Sponsorship:      storage.Sponsorship{Amount: "1000"},
				NotifiedSponsors: false,
			},
			expectedCompleted: 0, // NotifiedSponsors condition fails
			expectedTotal:     10,
			description:       "With sponsorship amount, NotifiedSponsors must be true",
		},
		{
			name: "With_sponsorship_amount_and_notification",
			video: storage.Video{
				Sponsorship:      storage.Sponsorship{Amount: "1000"},
				NotifiedSponsors: true,
			},
			expectedCompleted: 1, // NotifiedSponsors condition passes
			expectedTotal:     10,
			description:       "With sponsorship amount and notification, condition should pass",
		},
		{
			name: "Sponsorship_amount_NA",
			video: storage.Video{
				Sponsorship:      storage.Sponsorship{Amount: "N/A"},
				NotifiedSponsors: false,
			},
			expectedCompleted: 1, // NotifiedSponsors condition passes (N/A amount)
			expectedTotal:     10,
			description:       "N/A sponsorship amount should pass condition regardless of NotifiedSponsors",
		},
		{
			name: "Sponsorship_amount_dash",
			video: storage.Video{
				Sponsorship:      storage.Sponsorship{Amount: "-"},
				NotifiedSponsors: false,
			},
			expectedCompleted: 1, // NotifiedSponsors condition passes (dash amount)
			expectedTotal:     10,
			description:       "Dash sponsorship amount should pass condition regardless of NotifiedSponsors",
		},
		{
			name: "All_complete",
			video: storage.Video{
				DOTPosted:           true,
				BlueSkyPosted:       true,
				LinkedInPosted:      true,
				SlackPosted:         true,
				YouTubeHighlight:    true,
				YouTubeComment:      true,
				YouTubeCommentReply: true,
				GDE:                 true,
				Repo:                "github.com/user/repo",
				NotifiedSponsors:    true,
			},
			expectedCompleted: 10,
			expectedTotal:     10,
			description:       "All post-publish fields complete",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completed, total := manager.CalculatePostPublishProgress(tc.video)
			assert.Equal(t, tc.expectedCompleted, completed, "Completed count mismatch for %s", tc.description)
			assert.Equal(t, tc.expectedTotal, total, "Total count mismatch for %s", tc.description)
		})
	}
}

func TestCalculateAnalysisProgress(t *testing.T) {
	manager := video.NewManager(nil, aspect.NewService())

	testCases := []struct {
		name              string
		video             storage.Video
		expectedCompleted int
		expectedTotal     int
		description       string
	}{
		{
			name:              "No_titles",
			video:             storage.Video{},
			expectedCompleted: 0,
			expectedTotal:     0,
			description:       "Video with no titles should return 0/0",
		},
		{
			name: "One_title_no_share",
			video: storage.Video{
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Test Title", Share: 0},
				},
			},
			expectedCompleted: 0,
			expectedTotal:     1,
			description:       "Title without share percentage should not be counted as complete",
		},
		{
			name: "One_title_with_share",
			video: storage.Video{
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Test Title", Share: 45.5},
				},
			},
			expectedCompleted: 1,
			expectedTotal:     1,
			description:       "Title with share percentage should be counted as complete",
		},
		{
			name: "Three_titles_none_complete",
			video: storage.Video{
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title 1", Share: 0},
					{Index: 2, Text: "Title 2", Share: 0},
					{Index: 3, Text: "Title 3", Share: 0},
				},
			},
			expectedCompleted: 0,
			expectedTotal:     3,
			description:       "Three titles without shares should be 0/3",
		},
		{
			name: "Three_titles_some_complete",
			video: storage.Video{
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title 1", Share: 40.0},
					{Index: 2, Text: "Title 2", Share: 35.5},
					{Index: 3, Text: "Title 3", Share: 0},
				},
			},
			expectedCompleted: 2,
			expectedTotal:     3,
			description:       "Two titles with shares should be 2/3",
		},
		{
			name: "Three_titles_all_complete",
			video: storage.Video{
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title 1", Share: 40.0},
					{Index: 2, Text: "Title 2", Share: 35.5},
					{Index: 3, Text: "Title 3", Share: 24.5},
				},
			},
			expectedCompleted: 3,
			expectedTotal:     3,
			description:       "All three titles with shares should be 3/3",
		},
		{
			name: "Two_titles_mixed",
			video: storage.Video{
				Titles: []storage.TitleVariant{
					{Index: 1, Text: "Title 1", Share: 60.0},
					{Index: 2, Text: "Title 2", Share: 0},
				},
			},
			expectedCompleted: 1,
			expectedTotal:     2,
			description:       "One complete out of two should be 1/2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			completed, total := manager.CalculateAnalysisProgress(tc.video)
			assert.Equal(t, tc.expectedCompleted, completed, "Completed count mismatch for %s", tc.description)
			assert.Equal(t, tc.expectedTotal, total, "Total count mismatch for %s", tc.description)
		})
	}
}

func TestGetVideoPhase_ErrorHandling(t *testing.T) {
	// Test the error path in GetVideoPhase
	manager := video.NewManager(func(category, name, extension string) string {
		return "/nonexistent/path.yaml"
	}, nil)

	videoIndex := storage.VideoIndex{Category: "test", Name: "nonexistent"}
	phase := manager.GetVideoPhase(videoIndex)

	// Should return PhaseIdeas as default when file cannot be read
	assert.Equal(t, workflow.PhaseIdeas, phase, "Should return PhaseIdeas when file cannot be read")
}
