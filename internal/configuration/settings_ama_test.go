package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSettingsAMAYAMLParsing(t *testing.T) {
	yamlContent := `ama:
  enabled: true
  playlistId: "PLabc123"
  schedule: "0 10 * * *"
  emailTo: "ops@example.com"
`
	var s Settings
	require.NoError(t, yaml.Unmarshal([]byte(yamlContent), &s))

	assert.True(t, s.AMA.Enabled)
	assert.Equal(t, "PLabc123", s.AMA.PlaylistID)
	assert.Equal(t, "0 10 * * *", s.AMA.Schedule)
	assert.Equal(t, "ops@example.com", s.AMA.EmailTo)
}

func TestSettingsAMAValidate(t *testing.T) {
	tests := []struct {
		name    string
		ama     SettingsAMA
		wantErr bool
		errSubstr string
	}{
		{
			name: "disabled with empty fields - no error",
			ama:  SettingsAMA{Enabled: false},
		},
		{
			name: "disabled with all fields populated - no error",
			ama: SettingsAMA{
				Enabled:    false,
				PlaylistID: "",
				Schedule:   "",
				EmailTo:    "",
			},
		},
		{
			name: "enabled with all fields - no error",
			ama: SettingsAMA{
				Enabled:    true,
				PlaylistID: "PLxyz",
				Schedule:   "0 10 * * *",
				EmailTo:    "ops@example.com",
			},
		},
		{
			name: "enabled missing playlistId",
			ama: SettingsAMA{
				Enabled:  true,
				Schedule: "0 10 * * *",
				EmailTo:  "ops@example.com",
			},
			wantErr:   true,
			errSubstr: "playlistId",
		},
		{
			name: "enabled missing schedule",
			ama: SettingsAMA{
				Enabled:    true,
				PlaylistID: "PLxyz",
				EmailTo:    "ops@example.com",
			},
			wantErr:   true,
			errSubstr: "schedule",
		},
		{
			name: "enabled invalid cron schedule",
			ama: SettingsAMA{
				Enabled:    true,
				PlaylistID: "PLxyz",
				Schedule:   "not a cron expression",
				EmailTo:    "ops@example.com",
			},
			wantErr:   true,
			errSubstr: "valid cron",
		},
		{
			name: "enabled missing emailTo",
			ama: SettingsAMA{
				Enabled:    true,
				PlaylistID: "PLxyz",
				Schedule:   "0 10 * * *",
			},
			wantErr:   true,
			errSubstr: "emailTo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ama.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errSubstr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// TestInitGlobalSettingsAMAFromYAML verifies that AMA settings load from the
// settings.yaml file when no environment overrides are present.
func TestInitGlobalSettingsAMAFromYAML(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.yaml")

	yamlContent := `ai:
  provider: anthropic
  anthropic:
    key: test-key
    model: claude-sonnet-4-20250514
ama:
  enabled: true
  playlistId: "PLfromyaml"
  schedule: "30 9 * * *"
  emailTo: "yaml@example.com"
`
	require.NoError(t, os.WriteFile(settingsPath, []byte(yamlContent), 0644))

	originalSettings := GlobalSettings
	t.Cleanup(func() { GlobalSettings = originalSettings })
	GlobalSettings = Settings{}

	t.Setenv("SETTINGS_FILE", settingsPath)
	// Make sure no env overrides leak in.
	t.Setenv("AMA_ENABLED", "")
	t.Setenv("AMA_PLAYLIST_ID", "")
	t.Setenv("AMA_SCHEDULE", "")
	t.Setenv("AMA_EMAIL_TO", "")

	require.NoError(t, InitGlobalSettings())

	assert.True(t, GlobalSettings.AMA.Enabled)
	assert.Equal(t, "PLfromyaml", GlobalSettings.AMA.PlaylistID)
	assert.Equal(t, "30 9 * * *", GlobalSettings.AMA.Schedule)
	assert.Equal(t, "yaml@example.com", GlobalSettings.AMA.EmailTo)
}

// TestInitGlobalSettingsAMAEnvOverride verifies env vars override YAML values.
func TestInitGlobalSettingsAMAEnvOverride(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.yaml")

	yamlContent := `ai:
  provider: anthropic
  anthropic:
    key: test-key
    model: claude-sonnet-4-20250514
ama:
  enabled: false
  playlistId: "PLfromyaml"
  schedule: "0 8 * * *"
  emailTo: "yaml@example.com"
`
	require.NoError(t, os.WriteFile(settingsPath, []byte(yamlContent), 0644))

	originalSettings := GlobalSettings
	t.Cleanup(func() { GlobalSettings = originalSettings })
	GlobalSettings = Settings{}

	t.Setenv("SETTINGS_FILE", settingsPath)
	t.Setenv("AMA_ENABLED", "true")
	t.Setenv("AMA_PLAYLIST_ID", "PLfromenv")
	t.Setenv("AMA_SCHEDULE", "15 11 * * *")
	t.Setenv("AMA_EMAIL_TO", "env@example.com")

	require.NoError(t, InitGlobalSettings())

	assert.True(t, GlobalSettings.AMA.Enabled)
	assert.Equal(t, "PLfromenv", GlobalSettings.AMA.PlaylistID)
	assert.Equal(t, "15 11 * * *", GlobalSettings.AMA.Schedule)
	assert.Equal(t, "env@example.com", GlobalSettings.AMA.EmailTo)
}

// TestInitGlobalSettingsAMADefaultSchedule verifies that the schedule defaults
// to "0 10 * * *" when neither YAML nor env supplies one. AMA stays disabled
// by default, so this only matters when an operator later flips Enabled.
func TestInitGlobalSettingsAMADefaultSchedule(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.yaml")

	yamlContent := `ai:
  provider: anthropic
  anthropic:
    key: test-key
    model: claude-sonnet-4-20250514
`
	require.NoError(t, os.WriteFile(settingsPath, []byte(yamlContent), 0644))

	originalSettings := GlobalSettings
	t.Cleanup(func() { GlobalSettings = originalSettings })
	GlobalSettings = Settings{}

	t.Setenv("SETTINGS_FILE", settingsPath)
	t.Setenv("AMA_ENABLED", "")
	t.Setenv("AMA_PLAYLIST_ID", "")
	t.Setenv("AMA_SCHEDULE", "")
	t.Setenv("AMA_EMAIL_TO", "")

	require.NoError(t, InitGlobalSettings())

	assert.False(t, GlobalSettings.AMA.Enabled)
	assert.Equal(t, "0 10 * * *", GlobalSettings.AMA.Schedule)
}

// TestInitGlobalSettingsAMAValidationFailure verifies that misconfigured AMA
// (Enabled=true with a missing required field) makes startup fail loud.
func TestInitGlobalSettingsAMAValidationFailure(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.yaml")

	yamlContent := `ai:
  provider: anthropic
  anthropic:
    key: test-key
    model: claude-sonnet-4-20250514
ama:
  enabled: true
  playlistId: ""
  schedule: "0 10 * * *"
  emailTo: "ops@example.com"
`
	require.NoError(t, os.WriteFile(settingsPath, []byte(yamlContent), 0644))

	originalSettings := GlobalSettings
	t.Cleanup(func() { GlobalSettings = originalSettings })
	GlobalSettings = Settings{}

	t.Setenv("SETTINGS_FILE", settingsPath)
	t.Setenv("AMA_ENABLED", "")
	t.Setenv("AMA_PLAYLIST_ID", "")
	t.Setenv("AMA_SCHEDULE", "")
	t.Setenv("AMA_EMAIL_TO", "")

	err := InitGlobalSettings()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "playlistId")
}
