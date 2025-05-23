package slack

import (
	"fmt"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildHeaderBlock(t *testing.T) {
	tests := []struct {
		name          string
		videoDetails  storage.Video
		expectedTitle string
		expectError   bool
	}{
		{
			name: "Valid video title",
			videoDetails: storage.Video{
				Title: "My Awesome Video",
			},
			expectedTitle: "My Awesome Video",
			expectError:   false,
		},
		{
			name: "Empty video title",
			videoDetails: storage.Video{
				Title: "", // Empty title
			},
			expectedTitle: "",
			expectError:   true,
		},
		{
			name: "Video with other details but valid title",
			videoDetails: storage.Video{
				Title:       "Another Great Video",
				Description: "Some description",
				VideoId:     "video123",
			},
			expectedTitle: "Another Great Video",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headerBlock, err := BuildHeaderBlock(tt.videoDetails)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, headerBlock)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, headerBlock)
				require.NotNil(t, headerBlock.Text)
				assert.Equal(t, slack.MBTHeader, headerBlock.Type)
				assert.Equal(t, slack.PlainTextType, headerBlock.Text.Type)
				assert.Equal(t, tt.expectedTitle, headerBlock.Text.Text)
			}
		})
	}
}

// Placeholder for TestBuildContextBlock - to be implemented in Subtask 4.4
func TestBuildContextBlock(t *testing.T) {
	tests := []struct {
		name                string
		videoDetails        storage.Video
		expectedElements    []string // Expected text content of elements
		expectNilBlock      bool
		expectedNumElements int
	}{
		{
			name: "All fields present",
			videoDetails: storage.Video{
				Date:     "2023-10-26T10:00",
				Category: "Tutorials",
				Tags:     "golang,slack,api",
			},
			expectedElements:    []string{"Oct 26, 2023", "Tutorials", "golang | slack | api"},
			expectNilBlock:      false,
			expectedNumElements: 3,
		},
		{
			name: "Only Date present",
			videoDetails: storage.Video{
				Date: "2024-01-15T14:30",
			},
			expectedElements:    []string{"Jan 15, 2024"},
			expectNilBlock:      false,
			expectedNumElements: 1,
		},
		{
			name: "Only Category present",
			videoDetails: storage.Video{
				Category: "DevOps",
			},
			expectedElements:    []string{"DevOps"},
			expectNilBlock:      false,
			expectedNumElements: 1,
		},
		{
			name: "Only Tags present",
			videoDetails: storage.Video{
				Tags: "cicd,testing",
			},
			expectedElements:    []string{"cicd | testing"},
			expectNilBlock:      false,
			expectedNumElements: 1,
		},
		{
			name: "No relevant fields present",
			videoDetails: storage.Video{
				Date:     "", // Empty date
				Category: "", // Empty category
				Tags:     "", // Empty tags
			},
			expectNilBlock:      true,
			expectedNumElements: 0,
		},
		{
			name: "Date and Tags present, no Category",
			videoDetails: storage.Video{
				Date: "2023-11-01T00:00",
				Tags: "containers,orchestration",
			},
			expectedElements:    []string{"Nov 1, 2023", "containers | orchestration"},
			expectNilBlock:      false,
			expectedNumElements: 2,
		},
		{
			name: "Invalid date format",
			videoDetails: storage.Video{
				Date:     "26-10-2023", // Invalid format
				Category: "News",
			},
			expectedElements:    []string{"News"}, // Date should be skipped
			expectNilBlock:      false,
			expectedNumElements: 1,
		},
		{
			name: "Tags with extra spaces and empty segments",
			videoDetails: storage.Video{
				Tags: " tag1 ,  ,, tag2 ,tag3  ",
			},
			expectedElements:    []string{"tag1 | tag2 | tag3"},
			expectNilBlock:      false,
			expectedNumElements: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			block, err := BuildContextBlock(tc.videoDetails)
			require.NoError(t, err) // BuildContextBlock itself should not error based on content

			if tc.expectNilBlock {
				assert.Nil(t, block)
			} else {
				require.NotNil(t, block)
				assert.Equal(t, slack.MBTContext, block.Type)
				require.NotNil(t, block.ContextElements)
				require.Len(t, block.ContextElements.Elements, tc.expectedNumElements)

				for i, expectedText := range tc.expectedElements {
					mixedElement := block.ContextElements.Elements[i]
					textElement, ok := mixedElement.(*slack.TextBlockObject)
					require.True(t, ok, "Element %d is not a TextBlockObject", i)
					assert.Equal(t, expectedText, textElement.Text)
					// Check type based on what BuildContextBlock sets (Markdown for Date, PlainText for others)
					if strings.Contains(expectedText, ", 202") { // Heuristic for date
						assert.Equal(t, slack.MarkdownType, textElement.Type)
					} else {
						assert.Equal(t, slack.PlainTextType, textElement.Type)
					}
				}
			}
		})
	}
}

func TestBuildSectionBlockWithThumbnail(t *testing.T) {
	tests := []struct {
		name                 string
		videoDetails         storage.Video
		expectedText         string
		expectedThumbnailURL string
		expectError          bool
		errorContains        string
	}{
		{
			name: "Valid video with Highlight, Description, and VideoId",
			videoDetails: storage.Video{
				Title:       "Awesome Title",
				Highlight:   "This is the highlight.",
				Description: "This is the full description.",
				VideoId:     "vid123",
			},
			expectedText:         "This is the highlight.",
			expectedThumbnailURL: fmt.Sprintf(thumbnailURLFormat, "vid123"),
			expectError:          false,
		},
		{
			name: "Valid video with Description (no Highlight) and VideoId",
			videoDetails: storage.Video{
				Title:       "Another Title",
				Description: "Only description here.",
				VideoId:     "vid456",
			},
			expectedText:         "Only description here.",
			expectedThumbnailURL: fmt.Sprintf(thumbnailURLFormat, "vid456"),
			expectError:          false,
		},
		{
			name: "Valid video with no Highlight or Description, but with VideoId",
			videoDetails: storage.Video{
				Title:   "Minimal Video",
				VideoId: "vid789",
			},
			expectedText:         defaultSummary,
			expectedThumbnailURL: fmt.Sprintf(thumbnailURLFormat, "vid789"),
			expectError:          false,
		},
		{
			name: "Video with text fields but empty VideoId",
			videoDetails: storage.Video{
				Title:     "Text Only",
				Highlight: "Some highlight.",
				VideoId:   "", // Empty VideoId
			},
			expectError:   true,
			errorContains: "VideoId cannot be empty",
		},
		{
			name: "Video with very long description",
			videoDetails: storage.Video{
				Title:       "Long Desc Video",
				Description: strings.Repeat("a", maxDescriptionLength+100),
				VideoId:     "vidLong",
			},
			expectedText:         strings.Repeat("a", maxDescriptionLength-3) + "...",
			expectedThumbnailURL: fmt.Sprintf(thumbnailURLFormat, "vidLong"),
			expectError:          false,
		},
		{
			name: "Empty video details (should error due to empty VideoId)",
			videoDetails: storage.Video{
				Title:       "",
				Highlight:   "",
				Description: "",
				VideoId:     "",
			},
			expectError:   true,
			errorContains: "VideoId cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := BuildSectionBlockWithThumbnail(tt.videoDetails)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, block)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, block)
				assert.Equal(t, slack.MBTSection, block.Type)

				require.NotNil(t, block.Text)
				assert.Equal(t, slack.MarkdownType, block.Text.Type)
				assert.Equal(t, tt.expectedText, block.Text.Text)

				require.NotNil(t, block.Accessory)
				require.NotNil(t, block.Accessory.ImageElement, "Accessory.ImageElement should not be nil")
				assert.Equal(t, slack.METImage, block.Accessory.ImageElement.Type)
				assert.Equal(t, tt.expectedThumbnailURL, block.Accessory.ImageElement.ImageURL)
				// Use video title as alt text if available, otherwise a default or empty.
				// BuildSectionBlockWithThumbnail sets alt text to video.Title
				expectedAltText := tt.videoDetails.Title
				if expectedAltText == "" && tt.videoDetails.VideoId != "" {
					// If title is empty but we have a thumbnail, a generic alt text might be better than empty.
					// However, current implementation uses title directly.
				}
				assert.Equal(t, expectedAltText, block.Accessory.ImageElement.AltText)
			}
		})
	}
}

func TestBuildActionsBlock(t *testing.T) {
	tests := []struct {
		name             string
		videoDetails     storage.Video
		expectError      bool
		errorContains    string
		expectNumButtons int
		buttonChecks     []func(t *testing.T, buttons []slack.BlockElement)
	}{
		{
			name: "Valid VideoId, no ProjectURL",
			videoDetails: storage.Video{
				VideoId: "vid123",
			},
			expectError:      false,
			expectNumButtons: 1,
			buttonChecks: []func(t *testing.T, buttons []slack.BlockElement){
				func(t *testing.T, buttons []slack.BlockElement) {
					btn, ok := buttons[0].(*slack.ButtonBlockElement)
					require.True(t, ok)
					assert.Equal(t, "▶️ Watch Video", btn.Text.Text)
					assert.Equal(t, fmt.Sprintf(youtubeWatchURLFormat, "vid123"), btn.URL)
					assert.Equal(t, slack.StylePrimary, btn.Style)
					assert.Equal(t, "watch_video_button", btn.ActionID)
					assert.Equal(t, "vid123", btn.Value)
				},
			},
		},
		{
			name: "Valid VideoId and ProjectURL",
			videoDetails: storage.Video{
				VideoId:    "vid456",
				ProjectURL: "http://example.com/project",
			},
			expectError:      false,
			expectNumButtons: 2,
			buttonChecks: []func(t *testing.T, buttons []slack.BlockElement){
				func(t *testing.T, buttons []slack.BlockElement) { // Watch button
					btn, ok := buttons[0].(*slack.ButtonBlockElement)
					require.True(t, ok)
					assert.Equal(t, "▶️ Watch Video", btn.Text.Text)
					assert.Equal(t, fmt.Sprintf(youtubeWatchURLFormat, "vid456"), btn.URL)
				},
				func(t *testing.T, buttons []slack.BlockElement) { // Project button
					btn, ok := buttons[1].(*slack.ButtonBlockElement)
					require.True(t, ok)
					assert.Equal(t, "Project Details", btn.Text.Text)
					assert.Equal(t, "http://example.com/project", btn.URL)
					assert.Equal(t, "project_details_button", btn.ActionID)
					assert.Equal(t, "http://example.com/project", btn.Value)
				},
			},
		},
		{
			name: "Empty VideoId",
			videoDetails: storage.Video{
				VideoId: "", // Empty VideoId
			},
			expectError:   true,
			errorContains: "VideoId is required",
		},
		{
			name: "No VideoId, with ProjectURL (should still error)",
			videoDetails: storage.Video{
				VideoId:    "",
				ProjectURL: "http://example.com/project",
			},
			expectError:   true,
			errorContains: "VideoId is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := BuildActionsBlock(tt.videoDetails)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, block)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, block)
				assert.Equal(t, slack.MBTAction, block.Type)
				require.Len(t, block.Elements.ElementSet, tt.expectNumButtons)
				for _, check := range tt.buttonChecks {
					check(t, block.Elements.ElementSet)
				}
			}
		})
	}
}

func TestBuildMessage(t *testing.T) {
	tests := []struct {
		name           string
		videoDetails   storage.Video
		expectError    bool
		errorContains  string
		expectedBlocks int // Expected number of non-nil blocks
	}{
		{
			name: "All details valid, expect all blocks",
			videoDetails: storage.Video{
				Title:      "Full Video",
				VideoId:    "full123",
				Highlight:  "Highlight here.",
				Date:       "2023-01-01T10:00",
				Category:   "Test Category",
				Tags:       "tagA,tagB",
				ProjectURL: "http://project.url",
			},
			expectError:    false,
			expectedBlocks: 4, // Header, Section, Context, Actions
		},
		{
			name: "Missing Title, should error from BuildHeaderBlock",
			videoDetails: storage.Video{
				VideoId: "noTitle123",
			},
			expectError:   true,
			errorContains: "video title cannot be empty",
		},
		{
			name: "Missing VideoId, should error from BuildSectionBlockWithThumbnail (or Actions)",
			videoDetails: storage.Video{
				Title: "No Vid ID",
			},
			expectError:   true,
			errorContains: "VideoId cannot be empty", // Could also be from Actions
		},
		{
			name: "Only Title and VideoId, expect Header, Section, Actions",
			videoDetails: storage.Video{
				Title:   "Minimal",
				VideoId: "min123",
			},
			expectError:    false,
			expectedBlocks: 3, // Context block will be nil
		},
		{
			name: "Only Title, VideoId, Date, expect Header, Section, Context (1 element), Actions",
			videoDetails: storage.Video{
				Title:   "Date Video",
				VideoId: "date123",
				Date:    "2023-02-01T12:00",
			},
			expectError:    false,
			expectedBlocks: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks, err := BuildMessage(tt.videoDetails)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, blocks)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, blocks)
				// Count non-nil blocks. The individual build functions might return nil if not applicable.
				// BuildMessage itself doesn't filter nils, it just appends. This is a bit of a simplification for this test.
				// A more robust check would inspect the types of blocks returned.
				activeBlocks := 0
				for _, b := range blocks {
					if b != nil {
						activeBlocks++
					}
				}
				assert.Equal(t, tt.expectedBlocks, activeBlocks)
			}
		})
	}
}
