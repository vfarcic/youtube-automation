package slack

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"

	"devopstoolkitseries/youtube-automation/internal/storage"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: Config{
				Token:         "xoxb-valid-token",
				DefaultChannel: "general",
				Reactions:     []string{"thumbsup"},
			},
			wantErr: false,
		},
		{
			name: "Missing token",
			config: Config{
				Token:         "",
				DefaultChannel: "general",
				Reactions:     []string{"thumbsup"},
			},
			wantErr: true,
		},
		{
			name: "Missing default channel",
			config: Config{
				Token:         "xoxb-valid-token",
				DefaultChannel: "",
				Reactions:     []string{"thumbsup"},
			},
			wantErr: false, // This is valid but will generate a warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	token := "xoxb-test-token"
	defaultChannel := "general"
	reactions := []string{"thumbsup", "rocket"}

	config := GetConfig(token, defaultChannel, reactions)

	assert.Equal(t, token, config.Token)
	assert.Equal(t, defaultChannel, config.DefaultChannel)
	assert.Equal(t, reactions, config.Reactions)
}

func TestCreateSlackMessageBlocks(t *testing.T) {
	video := storage.Video{
		Title:       "Test Video Title",
		Description: "This is a test video description",
		VideoId:     "test123",
	}

	blocks := createSlackMessageBlocks(video)

	// Check that we have at least 4 blocks (header, divider, description, link)
	assert.GreaterOrEqual(t, len(blocks), 4)

	// Check that the first block is a section with the title
	if sectionBlock, ok := blocks[0].(*slack.SectionBlock); ok {
		assert.Contains(t, sectionBlock.Text.Text, "Test Video Title")
	} else {
		t.Errorf("Expected first block to be a section block")
	}
}

func TestGetSlackSummary(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{
			name:        "Short description",
			description: "This is a short description",
			expected:    "This is a short description",
		},
		{
			name:        "Long description",
			description: string(make([]byte, 400)), // 400 characters
			expected:    string(make([]byte, 297)) + "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSlackSummary(tt.description)
			assert.Equal(t, tt.expected, result)
		})
	}
}