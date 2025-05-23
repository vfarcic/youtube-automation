package slack

import (
	"os"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateSlackPostStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "slack_status_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		initialPosted bool
		newPosted     bool
		videoPath     string // if empty, will create a temp file
		expectError   bool
		setupFile     func(filePath string) *storage.Video // Function to set up the initial file state
	}{
		{
			name:          "Set SlackPosted to true",
			initialPosted: false,
			newPosted:     true,
			expectError:   false,
			setupFile: func(filePath string) *storage.Video {
				v := &storage.Video{Name: "Test Video", SlackPosted: false, Path: filePath}
				yamlWriter := storage.YAML{IndexPath: ""}
				require.NoError(t, yamlWriter.WriteVideo(*v, filePath))
				return v
			},
		},
		{
			name:          "Set SlackPosted to false",
			initialPosted: true,
			newPosted:     false,
			expectError:   false,
			setupFile: func(filePath string) *storage.Video {
				v := &storage.Video{Name: "Test Video", SlackPosted: true, Path: filePath}
				yamlWriter := storage.YAML{IndexPath: ""}
				require.NoError(t, yamlWriter.WriteVideo(*v, filePath))
				return v
			},
		},
		{
			name:          "No change to SlackPosted (already true)",
			initialPosted: true,
			newPosted:     true,
			expectError:   false,
			setupFile: func(filePath string) *storage.Video {
				v := &storage.Video{Name: "Test Video", SlackPosted: true, Path: filePath}
				yamlWriter := storage.YAML{IndexPath: ""}
				require.NoError(t, yamlWriter.WriteVideo(*v, filePath))
				return v
			},
		},
		{
			name:          "Error writing to invalid path",
			initialPosted: false,
			newPosted:     true,
			videoPath:     "/this/path/should/not/exist/video.yaml",
			expectError:   true,
			setupFile: func(filePath string) *storage.Video {
				// No file setup needed as we expect a write error to a non-existent directory
				return &storage.Video{Name: "Test Video", SlackPosted: false, Path: filePath}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			videoFilePath := tt.videoPath
			if videoFilePath == "" {
				videoFile, err := os.CreateTemp(tempDir, "video-*.yaml")
				require.NoError(t, err)
				videoFilePath = videoFile.Name()
				videoFile.Close() // Close the file so WriteVideo can open it
			}

			video := tt.setupFile(videoFilePath)
			video.Path = videoFilePath // Ensure path is set for UpdateSlackPostStatus

			err := UpdateSlackPostStatus(video, tt.newPosted, videoFilePath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Read the file back and check the SlackPosted status
				yamlReader := storage.YAML{IndexPath: ""} // IndexPath might not be needed for GetVideo if path is absolute
				updatedVideo := yamlReader.GetVideo(videoFilePath)
				assert.Equal(t, tt.newPosted, updatedVideo.SlackPosted, "SlackPosted status should be updated in the file")
				assert.Equal(t, video.Name, updatedVideo.Name) // Sanity check other field
			}
		})
	}
}
