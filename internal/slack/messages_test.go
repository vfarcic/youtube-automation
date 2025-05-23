package slack

import (
	"errors"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockInternalSlackClient is a mock for our internal slackGoClientInterface.
// It embeds mock.Mock to get the mocking capabilities like On, Called, AssertExpectations.
type mockInternalSlackClient struct {
	mock.Mock
}

// PostMessage is the mocked method.
func (m *mockInternalSlackClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	// This is how testify/mock expects the mocked method to be implemented.
	// It records the call and returns whatever was configured with m.On(...).
	args := m.Called(channelID, options)
	return args.String(0), args.String(1), args.Error(2)
}

func TestPostVideoThumbnail(t *testing.T) {
	const testChannelID = "C0123CHANNEL"

	tests := []struct {
		name          string
		videoDetails  storage.Video
		mockSetup     func(mockClient *mockInternalSlackClient) // Sets expectations on the mock
		expectError   bool
		errorContains string
	}{
		{
			name: "Successful post with valid video details",
			videoDetails: storage.Video{
				VideoId:   "vid001",
				Title:     "Test Video Title",
				Thumbnail: "http://example.com/thumbnail.jpg",
			},
			mockSetup: func(mockClient *mockInternalSlackClient) {
				mockClient.On("PostMessage",
					testChannelID,
					// Simplified matcher: check that exactly two MsgOption arguments are passed.
					// We trust that PostVideoThumbnail constructs them correctly (Text then Attachments).
					mock.MatchedBy(func(options []slack.MsgOption) bool { return len(options) == 2 }),
				).Return(testChannelID, "12345.67890", nil).Once()
			},
			expectError: false,
		},
		{
			name:          "Error when VideoId is empty",
			videoDetails:  storage.Video{Title: "No Video ID", Thumbnail: "http://example.com/image.png"},
			mockSetup:     func(mockClient *mockInternalSlackClient) { /* No PostMessage call expected */ },
			expectError:   true,
			errorContains: "VideoId is empty",
		},
		{
			name:          "Error when Thumbnail is empty",
			videoDetails:  storage.Video{VideoId: "vid002", Title: "No Thumbnail"},
			mockSetup:     func(mockClient *mockInternalSlackClient) { /* No PostMessage call expected */ },
			expectError:   true,
			errorContains: "Thumbnail URL is empty",
		},
		{
			name: "Error when Slack PostMessage itself fails",
			videoDetails: storage.Video{
				VideoId:   "vid003",
				Title:     "API Error Video",
				Thumbnail: "http://example.com/apifail.jpg",
			},
			mockSetup: func(mockClient *mockInternalSlackClient) {
				mockClient.On("PostMessage",
					testChannelID,
					mock.MatchedBy(func(options []slack.MsgOption) bool { return len(options) == 2 }),
				).Return("", "", errors.New("simulated Slack API error")).Once()
			},
			expectError:   true,
			errorContains: "failed to post Slack message",
		},
		{
			name: "Handles empty video title by using default",
			videoDetails: storage.Video{
				VideoId:   "vid004",
				Title:     "", // Empty title
				Thumbnail: "http://example.com/default_title.jpg",
			},
			mockSetup: func(mockClient *mockInternalSlackClient) {
				mockClient.On("PostMessage",
					testChannelID,
					mock.MatchedBy(func(options []slack.MsgOption) bool { return len(options) == 2 }),
				).Return(testChannelID, "12345.00004", nil).Once()
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSlackGo := new(mockInternalSlackClient)
			tc.mockSetup(mockSlackGo)

			dummyAuth := &SlackAuth{Token: "test-token"}
			actualClient, err := NewSlackClient(dummyAuth)
			assert.NoError(t, err, "NewSlackClient failed during test setup")

			actualClient.slackGoClient = mockSlackGo

			err = PostVideoThumbnail(actualClient, testChannelID, tc.videoDetails)

			if tc.expectError {
				assert.Error(t, err, "Expected an error but got none")
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains, "Error message mismatch")
				}
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}

			mockSlackGo.AssertExpectations(t)
		})
	}
}
