package platform

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockConfirmationStyle is a mock for the confirmationStyle interface
type MockConfirmationStyle struct {
	Called       bool
	ReceivedArgs []string
	RenderOutput string // What this mock's Render should return
}

func (m *MockConfirmationStyle) Render(args ...string) string {
	m.Called = true
	m.ReceivedArgs = args
	return m.RenderOutput // Or simply return args[0] if no transformation is needed by the mock
}

func TestPostTechnologyConversations(t *testing.T) {
	title := "Test Title"
	description := "Test Description"
	videoId := "TestVideoID"
	gist := "TestGist"
	projectName := "TestProject"
	projectURL := "http://test.project.com"
	relatedVideos := "TestRelatedVideos"

	additionalInfoResult := "Mocked Additional Info"
	mockGetAdditionalInfo := func(g, pn, pu, rv string) string {
		assert.Equal(t, gist, g, "getAdditionalInfo called with wrong gist")
		assert.Equal(t, projectName, pn, "getAdditionalInfo called with wrong projectName")
		assert.Equal(t, projectURL, pu, "getAdditionalInfo called with wrong projectURL")
		assert.Equal(t, relatedVideos, rv, "getAdditionalInfo called with wrong relatedVideos")
		return additionalInfoResult
	}

	mockStyle := &MockConfirmationStyle{RenderOutput: "Styled Message"}

	// As PostTechnologyConversations uses `println`, we can't directly assert its output easily.
	// We will assert that the mocks were called correctly and that the message passed to Render is correct.
	PostTechnologyConversations(title, description, videoId, gist, projectName, projectURL, relatedVideos, mockGetAdditionalInfo, mockStyle)

	assert.True(t, mockStyle.Called, "confirmationStyle.Render should have been called")
	assert.Len(t, mockStyle.ReceivedArgs, 1, "Render should be called with one argument")

	expectedMessagePart1 := "Use the following information to post it to https://wordpress.com/posts/technologyconversations.com manually."
	expectedMessagePart2 := fmt.Sprintf("Title:\n%s", title)
	expectedMessagePart3 := fmt.Sprintf("Description:\n%s", description)
	expectedMessagePart4 := fmt.Sprintf("Video ID:\n%s", videoId)
	expectedMessagePart5 := fmt.Sprintf("Additional info:\n%s", additionalInfoResult)

	receivedMessage := mockStyle.ReceivedArgs[0]
	assert.True(t, strings.Contains(receivedMessage, expectedMessagePart1), "Message should contain the manual posting instruction")
	assert.True(t, strings.Contains(receivedMessage, expectedMessagePart2), "Message should contain the title")
	assert.True(t, strings.Contains(receivedMessage, expectedMessagePart3), "Message should contain the description")
	assert.True(t, strings.Contains(receivedMessage, expectedMessagePart4), "Message should contain the video ID")
	assert.True(t, strings.Contains(receivedMessage, expectedMessagePart5), "Message should contain the additional info")
}
