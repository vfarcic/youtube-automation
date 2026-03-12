package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestLoadTimingRecommendations tests loading timing recommendations from settings.yaml
func TestLoadTimingRecommendations(t *testing.T) {
	tests := []struct {
		name         string
		yamlContent  string
		want         []TimingRecommendation
		wantErr      bool
		fileNotExist bool // If true, test without creating settings.yaml
	}{
		{
			name: "Load valid recommendations",
			yamlContent: `
email:
  from: "test@example.com"
  thumbnailTo: "thumb@example.com"
  editTo: "edit@example.com"
  financeTo: "finance@example.com"
ai:
  provider: azure
  azure:
    endpoint: "https://ai.example.com"
    key: "ai-key"
    deployment: "gpt-4"
youtube:
  apiKey: "yt-key"
hugo:
  path: "/hugo/path"
timing:
  recommendations:
    - day: "Monday"
      time: "16:00"
      reasoning: "European end-of-workday"
    - day: "Tuesday"
      time: "09:00"
      reasoning: "European morning"
`,
			want: []TimingRecommendation{
				{
					Day:       "Monday",
					Time:      "16:00",
					Reasoning: "European end-of-workday",
				},
				{
					Day:       "Tuesday",
					Time:      "09:00",
					Reasoning: "European morning",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty recommendations section",
			yamlContent: `
email:
  from: "test@example.com"
  thumbnailTo: "thumb@example.com"
  editTo: "edit@example.com"
  financeTo: "finance@example.com"
ai:
  provider: azure
  azure:
    endpoint: "https://ai.example.com"
    key: "ai-key"
    deployment: "gpt-4"
youtube:
  apiKey: "yt-key"
hugo:
  path: "/hugo/path"
timing:
  recommendations: []
`,
			want:    []TimingRecommendation{},
			wantErr: false,
		},
		{
			name: "Missing timing section",
			yamlContent: `
email:
  from: "test@example.com"
  thumbnailTo: "thumb@example.com"
  editTo: "edit@example.com"
  financeTo: "finance@example.com"
ai:
  provider: azure
  azure:
    endpoint: "https://ai.example.com"
    key: "ai-key"
    deployment: "gpt-4"
youtube:
  apiKey: "yt-key"
hugo:
  path: "/hugo/path"
`,
			want:    []TimingRecommendation{},
			wantErr: false,
		},
		{
			name:         "File does not exist - graceful handling",
			fileNotExist: true,
			want:         []TimingRecommendation{},
			wantErr:      false,
		},
		{
			name: "Invalid YAML - malformed",
			yamlContent: `
email:
  from: "test@example.com"
timing:
  recommendations:
    - day: "Monday
      time: "16:00"  # Missing closing quote on day
`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "timing-test-*")
			require.NoError(t, err, "Failed to create temp directory")
			defer os.RemoveAll(tempDir)

			// Save original working directory
			originalWd, err := os.Getwd()
			require.NoError(t, err, "Failed to get current working directory")
			defer func() {
				err = os.Chdir(originalWd)
				require.NoError(t, err, "Failed to restore original working directory")
			}()

			// Change to temp directory
			err = os.Chdir(tempDir)
			require.NoError(t, err, "Failed to change to temp directory")

			// Create settings file unless test wants to test non-existent file
			if !tt.fileNotExist {
				err = os.WriteFile("settings.yaml", []byte(tt.yamlContent), 0644)
				require.NoError(t, err, "Failed to write settings file")
			}

			// Test the function
			got, err := LoadTimingRecommendations("settings.yaml")

			// Check error condition
			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Unexpected error: %v", err)
				assert.Equal(t, tt.want, got, "Recommendations don't match expected")
			}
		})
	}
}

// TestSaveTimingRecommendations tests saving timing recommendations to settings.yaml
func TestSaveTimingRecommendations(t *testing.T) {
	baseYAMLContent := `email:
  from: test@example.com
  thumbnailTo: thumb@example.com
  editTo: edit@example.com
  financeTo: finance@example.com
ai:
  provider: azure
  azure:
    endpoint: https://ai.example.com
    key: ai-key
    deployment: gpt-4
youtube:
  apiKey: yt-key
hugo:
  path: /hugo/path
`

	tests := []struct {
		name                 string
		initialYAML          string
		recommendations      []TimingRecommendation
		wantErr              bool
		validateOtherFields  bool // If true, verify other settings were preserved
		fileNotExist         bool // If true, don't create initial settings.yaml
	}{
		{
			name:        "Save new recommendations to existing file",
			initialYAML: baseYAMLContent,
			recommendations: []TimingRecommendation{
				{
					Day:       "Monday",
					Time:      "16:00",
					Reasoning: "European end-of-workday",
				},
				{
					Day:       "Thursday",
					Time:      "13:00",
					Reasoning: "Mid-week afternoon",
				},
			},
			wantErr:             false,
			validateOtherFields: true,
		},
		{
			name: "Update existing recommendations",
			initialYAML: baseYAMLContent + `timing:
  recommendations:
    - day: "Monday"
      time: "10:00"
      reasoning: "Old timing"
`,
			recommendations: []TimingRecommendation{
				{
					Day:       "Tuesday",
					Time:      "15:00",
					Reasoning: "New timing",
				},
			},
			wantErr:             false,
			validateOtherFields: true,
		},
		{
			name:        "Save empty recommendations",
			initialYAML: baseYAMLContent,
			recommendations: []TimingRecommendation{},
			wantErr:             false,
			validateOtherFields: true,
		},
		{
			name:            "File does not exist - error",
			fileNotExist:    true,
			recommendations: []TimingRecommendation{{Day: "Monday", Time: "16:00", Reasoning: "Test"}},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "timing-save-test-*")
			require.NoError(t, err, "Failed to create temp directory")
			defer os.RemoveAll(tempDir)

			// Save original working directory
			originalWd, err := os.Getwd()
			require.NoError(t, err, "Failed to get current working directory")
			defer func() {
				err = os.Chdir(originalWd)
				require.NoError(t, err, "Failed to restore original working directory")
			}()

			// Change to temp directory
			err = os.Chdir(tempDir)
			require.NoError(t, err, "Failed to change to temp directory")

			// Create initial settings file unless test wants to test non-existent file
			if !tt.fileNotExist {
				err = os.WriteFile("settings.yaml", []byte(tt.initialYAML), 0644)
				require.NoError(t, err, "Failed to write initial settings file")
			}

			// Test the function
			err = SaveTimingRecommendations("settings.yaml", tt.recommendations)

			// Check error condition
			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
				return
			}

			assert.NoError(t, err, "Unexpected error: %v", err)

			// Verify the saved recommendations
			savedRecommendations, err := LoadTimingRecommendations("settings.yaml")
			require.NoError(t, err, "Failed to load saved recommendations")
			assert.Equal(t, tt.recommendations, savedRecommendations, "Saved recommendations don't match")

			// Verify other settings were preserved
			if tt.validateOtherFields {
				yamlFile, err := os.ReadFile("settings.yaml")
				require.NoError(t, err, "Failed to read saved settings file")

				var settings Settings
				err = yaml.Unmarshal(yamlFile, &settings)
				require.NoError(t, err, "Failed to parse saved settings file")

				// Check that original fields are preserved
				assert.Equal(t, "test@example.com", settings.Email.From, "Email.From was not preserved")
				assert.Equal(t, "thumb@example.com", settings.Email.ThumbnailTo, "Email.ThumbnailTo was not preserved")
				assert.Equal(t, "https://ai.example.com", settings.AI.Azure.Endpoint, "AI.Azure.Endpoint was not preserved")
				assert.Equal(t, "/hugo/path", settings.Hugo.Path, "Hugo.Path was not preserved")
			}
		})
	}
}

// TestTimingRecommendationRoundtrip tests that recommendations survive a save/load cycle
func TestTimingRecommendationRoundtrip(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "timing-roundtrip-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	defer func() {
		err = os.Chdir(originalWd)
		require.NoError(t, err, "Failed to restore original working directory")
	}()

	// Change to temp directory
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Create initial settings file
	initialYAML := `email:
  from: test@example.com
  thumbnailTo: thumb@example.com
  editTo: edit@example.com
  financeTo: finance@example.com
ai:
  provider: azure
  azure:
    endpoint: https://ai.example.com
    key: ai-key
    deployment: gpt-4
youtube:
  apiKey: yt-key
hugo:
  path: /hugo/path
`
	err = os.WriteFile("settings.yaml", []byte(initialYAML), 0644)
	require.NoError(t, err, "Failed to write initial settings file")

	// Test recommendations with various characters and formats
	originalRecommendations := []TimingRecommendation{
		{
			Day:       "Monday",
			Time:      "16:00",
			Reasoning: "European end-of-workday (5pm CET) + US mid-day",
		},
		{
			Day:       "Tuesday",
			Time:      "09:00",
			Reasoning: "European morning: 10-11am CET, high engagement",
		},
		{
			Day:       "Thursday",
			Time:      "13:00",
			Reasoning: "Mid-week afternoon - B2B audience active",
		},
		{
			Day:       "Wednesday",
			Time:      "10:00",
			Reasoning: "Global mid-week morning (special chars: ±°!@#)",
		},
	}

	// Save recommendations
	err = SaveTimingRecommendations("settings.yaml", originalRecommendations)
	require.NoError(t, err, "Failed to save recommendations")

	// Load recommendations
	loadedRecommendations, err := LoadTimingRecommendations("settings.yaml")
	require.NoError(t, err, "Failed to load recommendations")

	// Verify they match exactly
	assert.Equal(t, originalRecommendations, loadedRecommendations, "Recommendations changed during roundtrip")

	// Verify the file is valid YAML
	yamlFile, err := os.ReadFile("settings.yaml")
	require.NoError(t, err, "Failed to read settings file")

	var settings Settings
	err = yaml.Unmarshal(yamlFile, &settings)
	require.NoError(t, err, "Saved YAML is not valid")

	// Verify recommendations are in the settings struct
	assert.Equal(t, originalRecommendations, settings.Timing.Recommendations, "Recommendations in settings struct don't match")
}

// TestSaveTimingRecommendationsWriteError tests error handling when file cannot be written
func TestSaveTimingRecommendationsWriteError(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "timing-write-error-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	defer func() {
		err = os.Chdir(originalWd)
		require.NoError(t, err, "Failed to restore original working directory")
	}()

	// Change to temp directory
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Create initial settings file
	initialYAML := `email:
  from: test@example.com
  thumbnailTo: thumb@example.com
  editTo: edit@example.com
  financeTo: finance@example.com
ai:
  provider: azure
  azure:
    endpoint: https://ai.example.com
    key: ai-key
    deployment: gpt-4
youtube:
  apiKey: yt-key
hugo:
  path: /hugo/path
`
	err = os.WriteFile("settings.yaml", []byte(initialYAML), 0644)
	require.NoError(t, err, "Failed to write initial settings file")

	// Make the file read-only to simulate write error
	err = os.Chmod("settings.yaml", 0444)
	require.NoError(t, err, "Failed to change file permissions")

	// Try to save recommendations (should fail due to permissions)
	recommendations := []TimingRecommendation{
		{Day: "Monday", Time: "16:00", Reasoning: "Test"},
	}
	err = SaveTimingRecommendations("settings.yaml", recommendations)

	// Should get an error
	assert.Error(t, err, "Expected error when writing to read-only file")
	assert.Contains(t, err.Error(), "failed to write settings.yaml", "Error message should mention write failure")
}

// TestLoadTimingRecommendationsAbsolutePath tests using an absolute path instead of CWD
func TestLoadTimingRecommendationsAbsolutePath(t *testing.T) {
	tempDir := t.TempDir()

	settingsPath := filepath.Join(tempDir, "settings.yaml")
	yamlContent := `timing:
  recommendations:
    - day: "Friday"
      time: "15:00"
      reasoning: "End of week"
`
	err := os.WriteFile(settingsPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	recs, err := LoadTimingRecommendations(settingsPath)
	require.NoError(t, err)
	assert.Len(t, recs, 1)
	assert.Equal(t, "Friday", recs[0].Day)
	assert.Equal(t, "15:00", recs[0].Time)
}

// TestSaveTimingRecommendationsAbsolutePath tests saving with an absolute path
func TestSaveTimingRecommendationsAbsolutePath(t *testing.T) {
	tempDir := t.TempDir()

	settingsPath := filepath.Join(tempDir, "settings.yaml")
	initialYAML := `email:
  from: test@example.com
  thumbnailTo: thumb@example.com
  editTo: edit@example.com
  financeTo: finance@example.com
`
	err := os.WriteFile(settingsPath, []byte(initialYAML), 0644)
	require.NoError(t, err)

	recs := []TimingRecommendation{
		{Day: "Saturday", Time: "10:00", Reasoning: "Weekend morning"},
	}
	err = SaveTimingRecommendations(settingsPath, recs)
	require.NoError(t, err)

	loaded, err := LoadTimingRecommendations(settingsPath)
	require.NoError(t, err)
	assert.Equal(t, recs, loaded)
}

// TestLoadTimingRecommendationsWithSpecialCharacters tests handling of special characters
func TestLoadTimingRecommendationsWithSpecialCharacters(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "timing-special-test-*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	// Save original working directory
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	defer func() {
		err = os.Chdir(originalWd)
		require.NoError(t, err, "Failed to restore original working directory")
	}()

	// Change to temp directory
	err = os.Chdir(tempDir)
	require.NoError(t, err, "Failed to change to temp directory")

	// Create settings with special characters in reasoning
	yamlContent := `email:
  from: test@example.com
  thumbnailTo: thumb@example.com
  editTo: edit@example.com
  financeTo: finance@example.com
ai:
  provider: azure
  azure:
    endpoint: https://ai.example.com
    key: ai-key
    deployment: gpt-4
youtube:
  apiKey: yt-key
hugo:
  path: /hugo/path
timing:
  recommendations:
    - day: "Monday"
      time: "16:00"
      reasoning: "Test with quotes: \"value\" and 'value'"
    - day: "Tuesday"
      time: "09:00"
      reasoning: "Test with newlines:\nLine 2"
    - day: "Wednesday"
      time: "14:00"
      reasoning: "Unicode: 日本語, Emoji: 🚀"
`
	err = os.WriteFile("settings.yaml", []byte(yamlContent), 0644)
	require.NoError(t, err, "Failed to write settings file")

	// Load recommendations
	recommendations, err := LoadTimingRecommendations("settings.yaml")
	require.NoError(t, err, "Failed to load recommendations with special characters")

	// Verify special characters are preserved
	assert.Len(t, recommendations, 3, "Expected 3 recommendations")
	assert.Contains(t, recommendations[0].Reasoning, "quotes:", "Quote handling failed")
	assert.Contains(t, recommendations[1].Reasoning, "\n", "Newline handling failed")
	assert.Contains(t, recommendations[2].Reasoning, "🚀", "Emoji handling failed")
}
