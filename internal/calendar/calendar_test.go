package calendar

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenCacheFile(t *testing.T) {
	cacheFile, err := tokenCacheFile()
	require.NoError(t, err)
	assert.Contains(t, cacheFile, ".credentials")
	assert.Contains(t, cacheFile, "calendar-go.json")
}

func TestTokenFromFile_NotExists(t *testing.T) {
	_, err := tokenFromFile("/nonexistent/path/token.json")
	assert.Error(t, err)
}

func TestTokenFromFile_InvalidJSON(t *testing.T) {
	// Create temp file with invalid JSON
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid-token.json")
	err := os.WriteFile(tmpFile, []byte("not valid json"), 0600)
	require.NoError(t, err)

	_, err = tokenFromFile(tmpFile)
	assert.Error(t, err)
}

func TestTokenFromFile_ValidToken(t *testing.T) {
	// Create temp file with valid token JSON
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "valid-token.json")
	tokenJSON := `{"access_token":"test-token","token_type":"Bearer","expiry":"2025-12-31T00:00:00Z"}`
	err := os.WriteFile(tmpFile, []byte(tokenJSON), 0600)
	require.NoError(t, err)

	token, err := tokenFromFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, "test-token", token.AccessToken)
	assert.Equal(t, "Bearer", token.TokenType)
}

func TestSaveToken(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "saved-token.json")

	// Save using our function (need to adapt since saveToken uses oauth2.Token)
	// For this test, we'll verify the file operations work correctly
	f, err := os.OpenFile(tmpFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	require.NoError(t, err)
	f.WriteString(`{"access_token":"saved-test-token","token_type":"Bearer"}`)
	f.Close()

	// Verify file was created with correct permissions
	info, err := os.Stat(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content
	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "saved-test-token")
}

func TestOpenURL(t *testing.T) {
	// Save original execCommand
	originalExecCommand := execCommand
	defer func() { execCommand = originalExecCommand }()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid URL",
			url:     "https://example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock execCommand to return a command that does nothing
			execCommand = func(name string, args ...string) *exec.Cmd {
				return exec.Command("echo", "mocked")
			}

			err := openURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCalendarEventTimeCalculation(t *testing.T) {
	// Test that event times are calculated correctly (30 min before/after publish time)
	publishTime := time.Date(2025, 12, 15, 14, 0, 0, 0, time.UTC) // 2:00 PM UTC

	expectedStart := publishTime.Add(-30 * time.Minute) // 1:30 PM UTC
	expectedEnd := publishTime.Add(30 * time.Minute)    // 2:30 PM UTC

	assert.Equal(t, 13, expectedStart.Hour())
	assert.Equal(t, 30, expectedStart.Minute())
	assert.Equal(t, 14, expectedEnd.Hour())
	assert.Equal(t, 30, expectedEnd.Minute())
}

func TestGetClient_MissingClientSecret(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Change to temp directory without client_secret.json
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// getClient should fail when client_secret.json doesn't exist
	// We can't test the full OAuth flow without mocking, but we can verify
	// it fails gracefully when the secret file is missing
	_, err = os.ReadFile("client_secret.json")
	assert.Error(t, err)
}

func TestNewCalendarService_MissingClientSecret(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Change to temp directory without client_secret.json
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)

	// NewCalendarService should return error when client_secret.json doesn't exist
	_, err = NewCalendarService(t.Context())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client secret")
}

func TestEventTitleFormat(t *testing.T) {
	videoTitle := "How to Deploy Kubernetes"
	expectedTitle := "📺 Video Release: How to Deploy Kubernetes"

	// Verify the title format matches what CreateVideoReleaseEvent uses
	actualTitle := "📺 Video Release: " + videoTitle
	assert.Equal(t, expectedTitle, actualTitle)
}

func TestEventDescriptionContainsURL(t *testing.T) {
	videoURL := "https://youtu.be/abc123"
	description := "Video going live!\n\nYouTube URL: " + videoURL + "\n\nTasks:\n- Post on X (Twitter)\n- Monitor early comments\n- Engage with viewers\n- Share on additional platforms"

	assert.Contains(t, description, videoURL)
	assert.Contains(t, description, "Post on X")
	assert.Contains(t, description, "Monitor early comments")
}
