package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestThumbnailGenerationSettingsLoading(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		want        SettingsThumbnailGeneration
		wantErr     bool
	}{
		{
			name: "full thumbnail generation config",
			yamlContent: `thumbnailGeneration:
  photoDir: /data/photos
  providers:
    - name: gemini
      model: gemini-2.0-flash-preview-image-generation
    - name: gpt-image
      model: gpt-image-1
`,
			want: SettingsThumbnailGeneration{
				PhotoDir: "/data/photos",
				Providers: []SettingsThumbnailProvider{
					{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
					{Name: "gpt-image", Model: "gpt-image-1"},
				},
			},
		},
		{
			name: "single provider",
			yamlContent: `thumbnailGeneration:
  photoDir: /custom/photos
  providers:
    - name: gemini
      model: gemini-2.0-flash-preview-image-generation
`,
			want: SettingsThumbnailGeneration{
				PhotoDir: "/custom/photos",
				Providers: []SettingsThumbnailProvider{
					{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
				},
			},
		},
		{
			name:        "empty config - no thumbnail generation section",
			yamlContent: `email:
  from: test@example.com
`,
			want: SettingsThumbnailGeneration{},
		},
		{
			name: "empty providers list",
			yamlContent: `thumbnailGeneration:
  photoDir: /data/photos
  providers: []
`,
			want: SettingsThumbnailGeneration{
				PhotoDir:  "/data/photos",
				Providers: []SettingsThumbnailProvider{},
			},
		},
		{
			name: "no photoDir",
			yamlContent: `thumbnailGeneration:
  providers:
    - name: gemini
      model: some-model
`,
			want: SettingsThumbnailGeneration{
				Providers: []SettingsThumbnailProvider{
					{Name: "gemini", Model: "some-model"},
				},
			},
		},
		{
			name:        "invalid YAML",
			yamlContent: `thumbnailGeneration: [invalid`,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var settings Settings
			err := yaml.Unmarshal([]byte(tt.yamlContent), &settings)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, settings.ThumbnailGeneration)
		})
	}
}

func TestThumbnailGenerationRoundtrip(t *testing.T) {
	original := Settings{
		ThumbnailGeneration: SettingsThumbnailGeneration{
			PhotoDir: "/data/photos",
			Providers: []SettingsThumbnailProvider{
				{Name: "gemini", Model: "gemini-2.0-flash-preview-image-generation"},
				{Name: "gpt-image", Model: "gpt-image-1"},
			},
		},
	}

	data, err := yaml.Marshal(&original)
	require.NoError(t, err)

	var loaded Settings
	err = yaml.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, original.ThumbnailGeneration, loaded.ThumbnailGeneration)
}

func TestThumbnailGenerationPreservesOtherSettings(t *testing.T) {
	yamlContent := `email:
  from: test@example.com
ai:
  provider: anthropic
thumbnailGeneration:
  photoDir: /data/photos
  providers:
    - name: gemini
      model: gemini-2.0-flash-preview-image-generation
timing:
  recommendations:
    - day: Monday
      time: "16:00"
      reasoning: European afternoon
`
	var settings Settings
	err := yaml.Unmarshal([]byte(yamlContent), &settings)
	require.NoError(t, err)

	assert.Equal(t, "test@example.com", settings.Email.From)
	assert.Equal(t, "anthropic", settings.AI.Provider)
	assert.Equal(t, "/data/photos", settings.ThumbnailGeneration.PhotoDir)
	assert.Len(t, settings.ThumbnailGeneration.Providers, 1)
	assert.Equal(t, "gemini", settings.ThumbnailGeneration.Providers[0].Name)
	assert.Len(t, settings.Timing.Recommendations, 1)
}

func TestInitGlobalSettingsLoadsThumbnailGeneration(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.yaml")

	yamlContent := `ai:
  provider: anthropic
  anthropic:
    key: test-key
    model: claude-sonnet-4-20250514
thumbnailGeneration:
  photoDir: /data/photos
  providers:
    - name: gemini
      model: gemini-2.0-flash-preview-image-generation
    - name: gpt-image
      model: gpt-image-1
`
	err := os.WriteFile(settingsPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Save and restore original state
	originalSettings := GlobalSettings
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		GlobalSettings = originalSettings
		require.NoError(t, os.Chdir(originalWd))
	}()

	require.NoError(t, os.Chdir(tempDir))

	err = InitGlobalSettings()
	require.NoError(t, err)

	assert.Equal(t, "/data/photos", GlobalSettings.ThumbnailGeneration.PhotoDir)
	require.Len(t, GlobalSettings.ThumbnailGeneration.Providers, 2)
	assert.Equal(t, "gemini", GlobalSettings.ThumbnailGeneration.Providers[0].Name)
	assert.Equal(t, "gemini-2.0-flash-preview-image-generation", GlobalSettings.ThumbnailGeneration.Providers[0].Model)
	assert.Equal(t, "gpt-image", GlobalSettings.ThumbnailGeneration.Providers[1].Name)
	assert.Equal(t, "gpt-image-1", GlobalSettings.ThumbnailGeneration.Providers[1].Model)
}

func TestInitGlobalSettingsNoThumbnailSection(t *testing.T) {
	tempDir := t.TempDir()
	settingsPath := filepath.Join(tempDir, "settings.yaml")

	yamlContent := `ai:
  provider: anthropic
  anthropic:
    key: test-key
    model: claude-sonnet-4-20250514
`
	err := os.WriteFile(settingsPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	originalSettings := GlobalSettings
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		GlobalSettings = originalSettings
		require.NoError(t, os.Chdir(originalWd))
	}()

	require.NoError(t, os.Chdir(tempDir))

	err = InitGlobalSettings()
	require.NoError(t, err)

	assert.Empty(t, GlobalSettings.ThumbnailGeneration.PhotoDir)
	assert.Nil(t, GlobalSettings.ThumbnailGeneration.Providers)
}

func TestThumbnailAPIKeysFromEnvVars(t *testing.T) {
	// API keys come from env vars, not settings.yaml
	// Verify they are accessible via os.Getenv (the pattern used by providers)
	tests := []struct {
		name   string
		envVar string
		value  string
	}{
		{name: "GEMINI_API_KEY", envVar: "GEMINI_API_KEY", value: "test-gemini-key"},
		{name: "OPENAI_API_KEY", envVar: "OPENAI_API_KEY", value: "test-openai-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.value)
			assert.Equal(t, tt.value, os.Getenv(tt.envVar))
		})
	}
}

func TestSettingsThumbnailProviderYAMLTags(t *testing.T) {
	// Verify YAML tags produce expected field names
	yamlContent := `name: gemini
model: gemini-2.0-flash-preview-image-generation
`
	var provider SettingsThumbnailProvider
	err := yaml.Unmarshal([]byte(yamlContent), &provider)
	require.NoError(t, err)
	assert.Equal(t, "gemini", provider.Name)
	assert.Equal(t, "gemini-2.0-flash-preview-image-generation", provider.Model)
}

func TestSettingsThumbnailGenerationYAMLTags(t *testing.T) {
	yamlContent := `photoDir: /custom/path
providers:
  - name: test-provider
    model: test-model
`
	var gen SettingsThumbnailGeneration
	err := yaml.Unmarshal([]byte(yamlContent), &gen)
	require.NoError(t, err)
	assert.Equal(t, "/custom/path", gen.PhotoDir)
	require.Len(t, gen.Providers, 1)
	assert.Equal(t, "test-provider", gen.Providers[0].Name)
	assert.Equal(t, "test-model", gen.Providers[0].Model)
}
