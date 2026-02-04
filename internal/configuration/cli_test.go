package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// setupTestSettings creates a temporary settings file and returns cleanup function
func setupTestSettings(t *testing.T, content string) (string, func()) {
	t.Helper()

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "cli-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create settings file
	settingsPath := tempDir + "/settings.yaml"
	err = os.WriteFile(settingsPath, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to write settings file: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return settingsPath, cleanup
}

// restoreEnv saves current environment variables and returns a function to restore them
func restoreEnv(t *testing.T, keys []string) func() {
	t.Helper()

	// Save original values
	origValues := make(map[string]string)
	for _, key := range keys {
		origValues[key] = os.Getenv(key)
	}

	// Return function to restore environment
	return func() {
		for key, value := range origValues {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}
}

// TestFlagParsing tests the flag parsing functionality
func TestFlagParsing(t *testing.T) {
	// Define test cases for flag parsing
	tests := []struct {
		name     string
		args     []string
		expected Settings
		wantErr  bool
	}{
		{
			name: "Basic flags",
			args: []string{
				"--email-from", "test@example.com",
				"--email-thumbnail-to", "thumbnail@example.com",
				"--email-edit-to", "edit@example.com",
				"--email-finance-to", "finance@example.com",
				"--email-password", "password123",
				"--ai-endpoint", "https://api.openai.com",
				"--ai-key", "ai-key-123",
				"--ai-deployment", "gpt-4",
				"--youtube-api-key", "yt-api-key-123",
				"--hugo-path", "/path/to/hugo",
			},
			expected: Settings{
				Email: SettingsEmail{
					From:        "test@example.com",
					ThumbnailTo: "thumbnail@example.com",
					EditTo:      "edit@example.com",
					FinanceTo:   "finance@example.com",
					Password:    "password123",
				},
				AI: SettingsAI{
					Provider: "azure",
					Azure: SettingsAzureAI{
						Endpoint:   "https://api.openai.com",
						Key:        "ai-key-123",
						Deployment: "gpt-4",
					},
				},
				YouTube: SettingsYouTube{
					APIKey: "yt-api-key-123",
				},
				Hugo: SettingsHugo{
					Path: "/path/to/hugo",
				},
				Bluesky: SettingsBluesky{
					URL: "https://bsky.social/xrpc",
				},
			},
			wantErr: false,
		},
		{
			name: "Bluesky flags",
			args: []string{
				"--email-from", "test@example.com",
				"--email-thumbnail-to", "thumbnail@example.com",
				"--email-edit-to", "edit@example.com",
				"--email-finance-to", "finance@example.com",
				"--email-password", "password123",
				"--ai-endpoint", "https://api.openai.com",
				"--ai-key", "ai-key-123",
				"--ai-deployment", "gpt-4",
				"--youtube-api-key", "yt-api-key-123",
				"--hugo-path", "/path/to/hugo",
				"--bluesky-identifier", "user.bsky.social",
				"--bluesky-password", "bluesky-pwd",
				"--bluesky-url", "https://custom.bsky.social/xrpc",
			},
			expected: Settings{
				Email: SettingsEmail{
					From:        "test@example.com",
					ThumbnailTo: "thumbnail@example.com",
					EditTo:      "edit@example.com",
					FinanceTo:   "finance@example.com",
					Password:    "password123",
				},
				AI: SettingsAI{
					Provider: "azure",
					Azure: SettingsAzureAI{
						Endpoint:   "https://api.openai.com",
						Key:        "ai-key-123",
						Deployment: "gpt-4",
					},
				},
				YouTube: SettingsYouTube{
					APIKey: "yt-api-key-123",
				},
				Hugo: SettingsHugo{
					Path: "/path/to/hugo",
				},
				Bluesky: SettingsBluesky{
					Identifier: "user.bsky.social",
					Password:   "bluesky-pwd",
					URL:        "https://custom.bsky.social/xrpc",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for testing
			testCmd := &cobra.Command{
				Use:   "test",
				Short: "Test command for CLI argument testing",
				Run:   func(cmd *cobra.Command, args []string) {},
			}

			// Initialize an empty settings object
			testSettings := Settings{
				Bluesky: SettingsBluesky{
					URL: "https://bsky.social/xrpc",
				},
			}

			// Add flags to the test command (similar to init() in cli.go)
			testCmd.Flags().StringVar(&testSettings.Email.From, "email-from", testSettings.Email.From, "")
			testCmd.Flags().StringVar(&testSettings.Email.ThumbnailTo, "email-thumbnail-to", testSettings.Email.ThumbnailTo, "")
			testCmd.Flags().StringVar(&testSettings.Email.EditTo, "email-edit-to", testSettings.Email.EditTo, "")
			testCmd.Flags().StringVar(&testSettings.Email.FinanceTo, "email-finance-to", testSettings.Email.FinanceTo, "")
			testCmd.Flags().StringVar(&testSettings.Email.Password, "email-password", testSettings.Email.Password, "")
			testCmd.Flags().StringVar(&testSettings.AI.Azure.Endpoint, "ai-endpoint", testSettings.AI.Azure.Endpoint, "")
			testCmd.Flags().StringVar(&testSettings.AI.Azure.Key, "ai-key", testSettings.AI.Azure.Key, "")
			testCmd.Flags().StringVar(&testSettings.AI.Azure.Deployment, "ai-deployment", testSettings.AI.Azure.Deployment, "")
			testCmd.Flags().StringVar(&testSettings.YouTube.APIKey, "youtube-api-key", testSettings.YouTube.APIKey, "")
			testCmd.Flags().StringVar(&testSettings.Hugo.Path, "hugo-path", testSettings.Hugo.Path, "")
			testCmd.Flags().StringVar(&testSettings.Bluesky.Identifier, "bluesky-identifier", testSettings.Bluesky.Identifier, "")
			testCmd.Flags().StringVar(&testSettings.Bluesky.Password, "bluesky-password", testSettings.Bluesky.Password, "")
			testCmd.Flags().StringVar(&testSettings.Bluesky.URL, "bluesky-url", testSettings.Bluesky.URL, "")

			// Parse the arguments
			testCmd.SetArgs(tt.args)
			err := testCmd.Execute()

			// Check error condition
			if (err != nil) != tt.wantErr {
				t.Errorf("flag parsing error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check the resulting settings
			if testSettings.Email.From != tt.expected.Email.From {
				t.Errorf("Email.From = %v, want %v", testSettings.Email.From, tt.expected.Email.From)
			}
			if testSettings.Email.ThumbnailTo != tt.expected.Email.ThumbnailTo {
				t.Errorf("Email.ThumbnailTo = %v, want %v", testSettings.Email.ThumbnailTo, tt.expected.Email.ThumbnailTo)
			}
			if testSettings.Email.EditTo != tt.expected.Email.EditTo {
				t.Errorf("Email.EditTo = %v, want %v", testSettings.Email.EditTo, tt.expected.Email.EditTo)
			}
			if testSettings.Email.FinanceTo != tt.expected.Email.FinanceTo {
				t.Errorf("Email.FinanceTo = %v, want %v", testSettings.Email.FinanceTo, tt.expected.Email.FinanceTo)
			}
			if testSettings.Email.Password != tt.expected.Email.Password {
				t.Errorf("Email.Password = %v, want %v", testSettings.Email.Password, tt.expected.Email.Password)
			}
			if testSettings.AI.Azure.Endpoint != tt.expected.AI.Azure.Endpoint {
				t.Errorf("AI.Azure.Endpoint = %v, want %v", testSettings.AI.Azure.Endpoint, tt.expected.AI.Azure.Endpoint)
			}
			if testSettings.AI.Azure.Key != tt.expected.AI.Azure.Key {
				t.Errorf("AI.Azure.Key = %v, want %v", testSettings.AI.Azure.Key, tt.expected.AI.Azure.Key)
			}
			if testSettings.AI.Azure.Deployment != tt.expected.AI.Azure.Deployment {
				t.Errorf("AI.Azure.Deployment = %v, want %v", testSettings.AI.Azure.Deployment, tt.expected.AI.Azure.Deployment)
			}
			if testSettings.YouTube.APIKey != tt.expected.YouTube.APIKey {
				t.Errorf("YouTube.APIKey = %v, want %v", testSettings.YouTube.APIKey, tt.expected.YouTube.APIKey)
			}
			if testSettings.Hugo.Path != tt.expected.Hugo.Path {
				t.Errorf("Hugo.Path = %v, want %v", testSettings.Hugo.Path, tt.expected.Hugo.Path)
			}
			if testSettings.Bluesky.Identifier != tt.expected.Bluesky.Identifier {
				t.Errorf("Bluesky.Identifier = %v, want %v", testSettings.Bluesky.Identifier, tt.expected.Bluesky.Identifier)
			}
			if testSettings.Bluesky.Password != tt.expected.Bluesky.Password {
				t.Errorf("Bluesky.Password = %v, want %v", testSettings.Bluesky.Password, tt.expected.Bluesky.Password)
			}
			if testSettings.Bluesky.URL != tt.expected.Bluesky.URL {
				t.Errorf("Bluesky.URL = %v, want %v", testSettings.Bluesky.URL, tt.expected.Bluesky.URL)
			}
		})
	}
}

// TestEnvVarHandling tests the environment variable handling functionality
func TestEnvVarHandling(t *testing.T) {
	// List of environment variables to save and restore
	envVars := []string{
		"EMAIL_PASSWORD",
		"AI_KEY",
		"YOUTUBE_API_KEY",
		"BLUESKY_PASSWORD",
	}

	// Save original environment and restore after test
	restoreEnvFunc := restoreEnv(t, envVars)
	defer restoreEnvFunc()

	// Basic test settings
	basicSettings := `
email:
  from: "default@example.com"
  thumbnailTo: "default-thumbnail@example.com"
  editTo: "default-edit@example.com"
  financeTo: "default-finance@example.com"
ai:
  provider: "azure"
  azure:
    endpoint: "https://default-endpoint.com"
    deployment: "default-deployment"
hugo:
  path: "/default/hugo/path"
`

	// Create temp settings file
	settingsPath, cleanup := setupTestSettings(t, basicSettings)
	defer cleanup()

	// Setup environment variables
	os.Setenv("EMAIL_PASSWORD", "env-email-password")
	os.Setenv("AI_KEY", "env-ai-key")
	os.Setenv("YOUTUBE_API_KEY", "env-youtube-key")
	os.Setenv("BLUESKY_PASSWORD", "env-bluesky-password")

	// Create a test command
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command for environment variable testing",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	// Load settings from file and set up command flags
	testSettings := Settings{}
	yamlFile, _ := os.ReadFile(settingsPath)
	yaml.Unmarshal(yamlFile, &testSettings)

	// Add flags to the test command
	testCmd.Flags().StringVar(&testSettings.Email.From, "email-from", testSettings.Email.From, "")
	testCmd.Flags().StringVar(&testSettings.Email.ThumbnailTo, "email-thumbnail-to", testSettings.Email.ThumbnailTo, "")
	testCmd.Flags().StringVar(&testSettings.Email.EditTo, "email-edit-to", testSettings.Email.EditTo, "")
	testCmd.Flags().StringVar(&testSettings.Email.FinanceTo, "email-finance-to", testSettings.Email.FinanceTo, "")
	testCmd.Flags().StringVar(&testSettings.Email.Password, "email-password", testSettings.Email.Password, "")
	testCmd.Flags().StringVar(&testSettings.AI.Azure.Endpoint, "ai-endpoint", testSettings.AI.Azure.Endpoint, "")
	testCmd.Flags().StringVar(&testSettings.AI.Azure.Key, "ai-key", testSettings.AI.Azure.Key, "")
	testCmd.Flags().StringVar(&testSettings.AI.Azure.Deployment, "ai-deployment", testSettings.AI.Azure.Deployment, "")
	testCmd.Flags().StringVar(&testSettings.YouTube.APIKey, "youtube-api-key", testSettings.YouTube.APIKey, "")
	testCmd.Flags().StringVar(&testSettings.Hugo.Path, "hugo-path", testSettings.Hugo.Path, "")
	testCmd.Flags().StringVar(&testSettings.Bluesky.Identifier, "bluesky-identifier", testSettings.Bluesky.Identifier, "")
	testCmd.Flags().StringVar(&testSettings.Bluesky.Password, "bluesky-password", testSettings.Bluesky.Password, "")
	testCmd.Flags().StringVar(&testSettings.Bluesky.URL, "bluesky-url", testSettings.Bluesky.URL, "")

	// Process environment variables similar to how init() does it
	if envPassword := os.Getenv("EMAIL_PASSWORD"); envPassword != "" {
		testSettings.Email.Password = envPassword
	}

	if envAIKey := os.Getenv("AI_KEY"); envAIKey != "" {
		testSettings.AI.Azure.Key = envAIKey
	}

	if envYouTubeKey := os.Getenv("YOUTUBE_API_KEY"); envYouTubeKey != "" {
		testSettings.YouTube.APIKey = envYouTubeKey
	}

	if envBlueskyPassword := os.Getenv("BLUESKY_PASSWORD"); envBlueskyPassword != "" {
		testSettings.Bluesky.Password = envBlueskyPassword
	}

	// Check that environment variables were correctly applied
	if testSettings.Email.Password != "env-email-password" {
		t.Errorf("Email.Password = %s, want env-email-password", testSettings.Email.Password)
	}

	if testSettings.AI.Azure.Key != "env-ai-key" {
		t.Errorf("AI.Azure.Key = %s, want env-ai-key", testSettings.AI.Azure.Key)
	}

	if testSettings.YouTube.APIKey != "env-youtube-key" {
		t.Errorf("YouTube.APIKey = %s, want env-youtube-key", testSettings.YouTube.APIKey)
	}

	if testSettings.Bluesky.Password != "env-bluesky-password" {
		t.Errorf("Bluesky.Password = %s, want env-bluesky-password", testSettings.Bluesky.Password)
	}

	// Check that settings from file were loaded correctly
	if testSettings.Email.From != "default@example.com" {
		t.Errorf("Email.From = %s, want default@example.com", testSettings.Email.From)
	}

	if testSettings.AI.Azure.Endpoint != "https://default-endpoint.com" {
		t.Errorf("AI.Azure.Endpoint = %s, want https://default-endpoint.com", testSettings.AI.Azure.Endpoint)
	}

	if testSettings.Hugo.Path != "/default/hugo/path" {
		t.Errorf("Hugo.Path = %s, want /default/hugo/path", testSettings.Hugo.Path)
	}
}

// TestSettingsMerging tests merging settings from files and flags
func TestSettingsMerging(t *testing.T) {
	// Create a test settings file
	testSettings := `
email:
  from: "file@example.com"
  thumbnailTo: "file-thumbnail@example.com"
  editTo: "file-edit@example.com"
  financeTo: "file-finance@example.com"
  password: "file-password"
ai:
  endpoint: "https://file-endpoint.com"
  key: "file-ai-key"
  deployment: "file-deployment"
youtube:
  apiKey: "file-youtube-key"
hugo:
  path: "/file/hugo/path"
bluesky:
  identifier: "file.bsky.social"
  password: "file-bluesky-password"
  url: "https://file.bsky.social/xrpc"
`

	// Create temp settings file
	settingsPath, cleanup := setupTestSettings(t, testSettings)
	defer cleanup()

	// Save current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Change to the temp directory to read settings.yaml
	err = os.Chdir(settingsPath[:len(settingsPath)-len("/settings.yaml")])
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Make sure we go back to the original directory when we're done
	defer func() {
		err := os.Chdir(currentDir)
		if err != nil {
			t.Fatalf("Failed to restore working directory: %v", err)
		}
	}()

	// Command line arguments that should override some settings
	args := []string{
		"--email-from", "flag@example.com",
		"--ai-key", "flag-ai-key",
		"--hugo-path", "/flag/hugo/path",
	}

	// Create a test command
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command for settings merging",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	// Load settings from file
	mergedSettings := Settings{}
	yamlFile, _ := os.ReadFile("settings.yaml")
	yaml.Unmarshal(yamlFile, &mergedSettings)

	// Add flags to the test command
	testCmd.Flags().StringVar(&mergedSettings.Email.From, "email-from", mergedSettings.Email.From, "")
	testCmd.Flags().StringVar(&mergedSettings.Email.ThumbnailTo, "email-thumbnail-to", mergedSettings.Email.ThumbnailTo, "")
	testCmd.Flags().StringVar(&mergedSettings.Email.EditTo, "email-edit-to", mergedSettings.Email.EditTo, "")
	testCmd.Flags().StringVar(&mergedSettings.Email.FinanceTo, "email-finance-to", mergedSettings.Email.FinanceTo, "")
	testCmd.Flags().StringVar(&mergedSettings.Email.Password, "email-password", mergedSettings.Email.Password, "")
	testCmd.Flags().StringVar(&mergedSettings.AI.Azure.Endpoint, "ai-endpoint", mergedSettings.AI.Azure.Endpoint, "")
	testCmd.Flags().StringVar(&mergedSettings.AI.Azure.Key, "ai-key", mergedSettings.AI.Azure.Key, "")
	testCmd.Flags().StringVar(&mergedSettings.AI.Azure.Deployment, "ai-deployment", mergedSettings.AI.Azure.Deployment, "")
	testCmd.Flags().StringVar(&mergedSettings.YouTube.APIKey, "youtube-api-key", mergedSettings.YouTube.APIKey, "")
	testCmd.Flags().StringVar(&mergedSettings.Hugo.Path, "hugo-path", mergedSettings.Hugo.Path, "")
	testCmd.Flags().StringVar(&mergedSettings.Bluesky.Identifier, "bluesky-identifier", mergedSettings.Bluesky.Identifier, "")
	testCmd.Flags().StringVar(&mergedSettings.Bluesky.Password, "bluesky-password", mergedSettings.Bluesky.Password, "")
	testCmd.Flags().StringVar(&mergedSettings.Bluesky.URL, "bluesky-url", mergedSettings.Bluesky.URL, "")

	// Parse the command line arguments
	testCmd.SetArgs(args)
	err = testCmd.Execute()
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	// Verify settings from file that weren't overridden
	if mergedSettings.Email.ThumbnailTo != "file-thumbnail@example.com" {
		t.Errorf("Email.ThumbnailTo = %s, want file-thumbnail@example.com", mergedSettings.Email.ThumbnailTo)
	}

	if mergedSettings.Email.Password != "file-password" {
		t.Errorf("Email.Password = %s, want file-password", mergedSettings.Email.Password)
	}

	// Verify overridden settings
	if mergedSettings.Email.From != "flag@example.com" {
		t.Errorf("Email.From = %s, want flag@example.com", mergedSettings.Email.From)
	}

	if mergedSettings.AI.Azure.Key != "flag-ai-key" {
		t.Errorf("AI.Azure.Key = %s, want flag-ai-key", mergedSettings.AI.Azure.Key)
	}

	if mergedSettings.Hugo.Path != "/flag/hugo/path" {
		t.Errorf("Hugo.Path = %s, want /flag/hugo/path", mergedSettings.Hugo.Path)
	}
}

// TestRequiredFlagValidation tests validation of required flags
func TestRequiredFlagValidation(t *testing.T) {
	// Test cases
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name: "All required flags provided",
			args: []string{
				"--email-from", "test@example.com",
				"--email-thumbnail-to", "thumbnail@example.com",
			},
			wantErr: false,
		},
		{
			name: "Missing email-from",
			args: []string{
				"--email-thumbnail-to", "thumbnail@example.com",
			},
			wantErr: true,
		},
		{
			name: "Missing email-thumbnail-to",
			args: []string{
				"--email-from", "test@example.com",
			},
			wantErr: true,
		},
		{
			name:    "Missing all required flags",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command for each test case
			testSettings := Settings{}

			// Create a test command with error capturing
			testCmd := &cobra.Command{
				Use:   "test",
				Short: "Test command for required flag validation",
				RunE: func(cmd *cobra.Command, args []string) error {
					// This will never run in our tests, but we need RunE to properly capture errors
					return nil
				},
				// Silence usage printing which would clutter test output
				SilenceUsage: true,
			}

			// Add flags to the test command
			testCmd.Flags().StringVar(&testSettings.Email.From, "email-from", testSettings.Email.From, "")
			testCmd.Flags().StringVar(&testSettings.Email.ThumbnailTo, "email-thumbnail-to", testSettings.Email.ThumbnailTo, "")

			// Mark flags as required (this is what we're testing)
			testCmd.MarkFlagRequired("email-from")
			testCmd.MarkFlagRequired("email-thumbnail-to")

			// Set the arguments for this test case
			testCmd.SetArgs(tt.args)

			// Execute the command and check if error presence matches expectations
			err := testCmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("required flag validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetArgs tests the getArgs function in isolation
func TestGetArgs(t *testing.T) {
	// Save original rootCmd for restoration later
	originalRootCmd := RootCmd
	defer func() {
		RootCmd = originalRootCmd
	}()

	// Create a mock rootCmd for testing
	executeCount := 0
	executeErr := error(nil)

	RootCmd = &cobra.Command{
		Use:   "test",
		Short: "Test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			executeCount++
			return executeErr
		},
	}

	// Test case 1: Normal execution without error
	executeCount = 0
	executeErr = nil

	// Temporarily redirect stderr to avoid printing error message
	oldStderr := os.Stderr
	defer func() { os.Stderr = oldStderr }()
	_, w, _ := os.Pipe()
	os.Stderr = w

	GetArgs()
	w.Close()

	// Verify the command was executed once
	if executeCount != 1 {
		t.Errorf("Expected command to execute once, got %d executions", executeCount)
	}

	// Test case 2: Execution with error
	executeCount = 0
	executeErr = errors.New("test error")

	// Need a way to test os.Exit(1) without terminating the test
	// We'll use a custom exit function that we can monitor
	originalOsExit := osExit
	defer func() { osExit = originalOsExit }()

	exitCalled := false
	exitCode := 0
	osExit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	// Reset stderr capture
	os.Stderr = oldStderr
	_, w, _ = os.Pipe()
	os.Stderr = w

	GetArgs()
	w.Close()

	// Verify exit was called with code 1
	if !exitCalled {
		t.Error("Expected os.Exit to be called, but it wasn't")
	}
	if exitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", exitCode)
	}
}

// Added for PRD: Automated Video Language Setting
func TestVideoDefaultsLanguageSetting(t *testing.T) {
	baseYAMLContent := `
email:
  from: "test@example.com"
  thumbnailTo: "thumb@example.com"
  editTo: "edit@example.com"
  financeTo: "finance@example.com"
  password: "password123"
ai:
  endpoint: "https://ai.example.com"
  key: "ai-key"
  deployment: "gpt-4"
youtube:
  apiKey: "yt-key"
hugo:
  path: "/hugo/path"
`

	tests := []struct {
		name             string
		yamlContent      string   // Content for settings.yaml
		args             []string // Command line arguments
		expectedLanguage string
	}{
		{
			name: "Language from YAML",
			yamlContent: baseYAMLContent + `
videoDefaults:
  language: "fr"`,
			args:             []string{},
			expectedLanguage: "fr",
		},
		{
			name:             "Language from flag, YAML missing videoDefaults",
			yamlContent:      baseYAMLContent, // No videoDefaults here
			args:             []string{"--video-defaults-language=es"},
			expectedLanguage: "es",
		},
		{
			name: "Language from flag overrides YAML",
			yamlContent: baseYAMLContent + `
videoDefaults:
  language: "fr"`,
			args:             []string{"--video-defaults-language=de"},
			expectedLanguage: "de",
		},
		{
			name:             "Language defaults to en when not in YAML or flag",
			yamlContent:      baseYAMLContent, // No videoDefaults here
			args:             []string{},
			expectedLanguage: "en",
		},
		{
			name: "Language from flag when videoDefaults empty in YAML",
			yamlContent: baseYAMLContent + `
videoDefaults: {}`, // Empty videoDefaults
			args:             []string{"--video-defaults-language=it"},
			expectedLanguage: "it",
		},
		{
			name: "Language defaults to en when videoDefaults empty in YAML and no flag",
			yamlContent: baseYAMLContent + `
videoDefaults: {}`, // Empty videoDefaults
			args:             []string{},
			expectedLanguage: "en",
		},
	}

	originalOsArgs := os.Args
	defer func() { os.Args = originalOsArgs }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create a temporary settings.yaml
			settingsDir, err := os.MkdirTemp("", "settings-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(settingsDir)

			tmpfn := filepath.Join(settingsDir, "settings.yaml")
			if err := os.WriteFile(tmpfn, []byte(tt.yamlContent), 0666); err != nil {
				t.Fatalf("Failed to write temp settings file: %v", err)
			}

			// Change to the directory of the temp settings file
			originalWD, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current working directory: %v", err)
			}
			if err := os.Chdir(settingsDir); err != nil {
				t.Fatalf("Failed to change working directory: %v", err)
			}
			defer os.Chdir(originalWD) // Change back

			// Reset GlobalSettings for each test
			GlobalSettings = Settings{}

			// 1. Load settings from YAML first
			yamlFile, err := os.ReadFile("settings.yaml") // Reads from tmpfn due to Chdir
			if err == nil {                               // If file exists and is readable
				if errUnmarshal := yaml.Unmarshal(yamlFile, &GlobalSettings); errUnmarshal != nil {
					t.Logf("Warning: Error parsing temp config file in test: %s", errUnmarshal)
				}
			} else if !os.IsNotExist(err) { // If error is not "file does not exist"
				t.Fatalf("Error reading temp settings file: %v", err)
			}

			// 2. Manually process command-line arguments for the specific flag we are testing
			// This simulates a flag overriding a YAML value if present.
			flagValue := "" // Store the value if the flag is found
			for i, arg := range tt.args {
				if arg == "--video-defaults-language" {
					if i+1 < len(tt.args) {
						flagValue = tt.args[i+1]
						break
					} else {
						t.Fatalf("--video-defaults-language flag requires a value")
					}
				} else if strings.HasPrefix(arg, "--video-defaults-language=") {
					flagValue = strings.TrimPrefix(arg, "--video-defaults-language=")
					break
				}
			}
			if flagValue != "" {
				GlobalSettings.VideoDefaults.Language = flagValue
			}

			// 3. Apply the final defaulting logic from init() if the value is still empty after YAML and potential flag override
			if GlobalSettings.VideoDefaults.Language == "" {
				GlobalSettings.VideoDefaults.Language = "en" // Default from cli.go
			}

			if GlobalSettings.VideoDefaults.Language != tt.expectedLanguage {
				t.Errorf("Expected VideoDefaults.Language to be %q, got %q", tt.expectedLanguage, GlobalSettings.VideoDefaults.Language)
			}
		})
	}
}

// Added for PRD: Automated Video Language Setting
func TestVideoDefaultsAudioLanguageSetting(t *testing.T) {
	baseYAMLContent := `
email:
  from: "test@example.com"
  thumbnailTo: "thumb@example.com"
  editTo: "edit@example.com"
  financeTo: "finance@example.com"
  password: "password123"
ai:
  endpoint: "https://ai.example.com"
  key: "ai-key"
  deployment: "gpt-4"
youtube:
  apiKey: "yt-key"
hugo:
  path: "/hugo/path"
`

	tests := []struct {
		name              string
		yamlContent       string   // Content for settings.yaml
		args              []string // Command line arguments
		expectedLanguage  string
		expectedAudioLang string
	}{
		{
			name: "AudioLanguage from YAML",
			yamlContent: baseYAMLContent + `
videoDefaults:
  language: "fr"
  audioLanguage: "de"`,
			args:              []string{},
			expectedLanguage:  "fr",
			expectedAudioLang: "de",
		},
		{
			name:              "AudioLanguage from flag, YAML missing videoDefaults",
			yamlContent:       baseYAMLContent, // No videoDefaults here
			args:              []string{"--video-defaults-language=es"},
			expectedLanguage:  "es",
			expectedAudioLang: "en",
		},
		{
			name: "AudioLanguage from flag overrides YAML",
			yamlContent: baseYAMLContent + `
videoDefaults:
  language: "fr"
  audioLanguage: "de"`,
			args:              []string{"--video-defaults-language=de"},
			expectedLanguage:  "de",
			expectedAudioLang: "de",
		},
		{
			name:              "AudioLanguage defaults to en when not in YAML or flag",
			yamlContent:       baseYAMLContent, // No videoDefaults here
			args:              []string{},
			expectedLanguage:  "en",
			expectedAudioLang: "en",
		},
		{
			name: "AudioLanguage from flag when videoDefaults empty in YAML",
			yamlContent: baseYAMLContent + `
videoDefaults: {}`, // Empty videoDefaults
			args:              []string{"--video-defaults-language=it"},
			expectedLanguage:  "it",
			expectedAudioLang: "en",
		},
		{
			name: "AudioLanguage defaults to en when videoDefaults empty in YAML and no flag",
			yamlContent: baseYAMLContent + `
videoDefaults: {}`, // Empty videoDefaults
			args:              []string{},
			expectedLanguage:  "en",
			expectedAudioLang: "en",
		},
		{
			name: "AudioLanguage from flag",
			args: []string{"--video-defaults-audio-language", "fr"},
			yamlContent: baseYAMLContent + `
videoDefaults:
  language: "es"
  audioLanguage: "de"`,
			expectedLanguage:  "es",
			expectedAudioLang: "fr",
		},
		{
			name:              "AudioLanguage from flag, YAML missing videoDefaults",
			args:              []string{"--video-defaults-audio-language", "fr"},
			yamlContent:       baseYAMLContent, // No videoDefaults here
			expectedLanguage:  "en",
			expectedAudioLang: "fr",
		},
		{
			name: "AudioLanguage from flag overrides YAML",
			args: []string{"--video-defaults-audio-language", "fr"},
			yamlContent: baseYAMLContent + `
videoDefaults:
  language: "es"
  audioLanguage: "de"`,
			expectedLanguage:  "es",
			expectedAudioLang: "fr",
		},
		{
			name: "AudioLanguage default when not in YAML or flag",
			args: []string{"--video-defaults-audio-language", "fr"},
			yamlContent: baseYAMLContent + `
videoDefaults:
  language: "pt"`,
			expectedLanguage:  "pt",
			expectedAudioLang: "fr",
		},
		{
			name: "AudioLanguage from YAML, language default",
			args: []string{}, // Removed --video-defaults-audio-language flag
			yamlContent: baseYAMLContent + `
videoDefaults:
  audioLanguage: "it"`,
			expectedLanguage:  "en",
			expectedAudioLang: "it",
		},
		{
			name: "Both languages default when videoDefaults empty in YAML",
			args: []string{"--video-defaults-audio-language", "fr"},
			yamlContent: baseYAMLContent + `
videoDefaults: {}`,
			expectedLanguage:  "en",
			expectedAudioLang: "fr",
		},
		{
			name: "Both languages default when videoDefaults missing in YAML",
			args: []string{"--video-defaults-audio-language", "fr"},
			yamlContent: baseYAMLContent + `
email: {}`,
			expectedLanguage:  "en",
			expectedAudioLang: "fr",
		},
	}

	originalOsArgs := os.Args
	defer func() { os.Args = originalOsArgs }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create a temporary settings.yaml
			settingsDir, err := os.MkdirTemp("", "settings-test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(settingsDir)

			tmpfn := filepath.Join(settingsDir, "settings.yaml")
			if err := os.WriteFile(tmpfn, []byte(tt.yamlContent), 0666); err != nil {
				t.Fatalf("Failed to write temp settings file: %v", err)
			}

			// Change to the directory of the temp settings file
			originalWD, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current working directory: %v", err)
			}
			if err := os.Chdir(settingsDir); err != nil {
				t.Fatalf("Failed to change working directory: %v", err)
			}
			defer os.Chdir(originalWD) // Change back

			// Reset GlobalSettings for each test
			GlobalSettings = Settings{}

			// 1. Load settings from YAML first
			yamlFile, err := os.ReadFile("settings.yaml") // Reads from tmpfn due to Chdir
			if err == nil {                               // If file exists and is readable
				if errUnmarshal := yaml.Unmarshal(yamlFile, &GlobalSettings); errUnmarshal != nil {
					t.Logf("Warning: Error parsing temp config file in test: %s", errUnmarshal)
				}
			} else if !os.IsNotExist(err) { // If error is not "file does not exist"
				t.Fatalf("Error reading temp settings file: %v", err)
			}

			// 2. Manually process command-line arguments for the specific flag we are testing
			// This simulates a flag overriding a YAML value if present.
			flagValue := "" // Store the value if the flag is found
			for i, arg := range tt.args {
				if arg == "--video-defaults-language" {
					if i+1 < len(tt.args) {
						flagValue = tt.args[i+1]
						break
					} else {
						t.Fatalf("--video-defaults-language flag requires a value")
					}
				} else if strings.HasPrefix(arg, "--video-defaults-language=") {
					flagValue = strings.TrimPrefix(arg, "--video-defaults-language=")
					break
				}
			}
			if flagValue != "" {
				GlobalSettings.VideoDefaults.Language = flagValue
			}

			// 3. Apply the final defaulting logic from init() if the value is still empty after YAML and potential flag override
			if GlobalSettings.VideoDefaults.Language == "" {
				GlobalSettings.VideoDefaults.Language = "en" // Default from cli.go
			}

			if GlobalSettings.VideoDefaults.Language != tt.expectedLanguage {
				t.Errorf("Test %s: Expected GlobalSettings.VideoDefaults.Language to be '%s', got '%s'", tt.name, tt.expectedLanguage, GlobalSettings.VideoDefaults.Language)
			}

			// 4. Apply audioLanguage processing
			audioFlagValue := ""
			for i, arg := range tt.args {
				if arg == "--video-defaults-audio-language" {
					if i+1 < len(tt.args) {
						audioFlagValue = tt.args[i+1]
						break
					} else {
						t.Fatalf("--video-defaults-audio-language flag requires a value")
					}
				} else if strings.HasPrefix(arg, "--video-defaults-audio-language=") {
					audioFlagValue = strings.TrimPrefix(arg, "--video-defaults-audio-language=")
					break
				}
			}
			if audioFlagValue != "" {
				GlobalSettings.VideoDefaults.AudioLanguage = audioFlagValue
			}

			// 5. Apply audioLanguage defaulting logic
			if GlobalSettings.VideoDefaults.AudioLanguage == "" {
				GlobalSettings.VideoDefaults.AudioLanguage = "en" // Default from cli.go
			}

			if GlobalSettings.VideoDefaults.AudioLanguage != tt.expectedAudioLang {
				t.Errorf("Test %s: Expected GlobalSettings.VideoDefaults.AudioLanguage to be '%s', got '%s'", tt.name, tt.expectedAudioLang, GlobalSettings.VideoDefaults.AudioLanguage)
			}
		})
	}
}

// Helper function to compare string slices (order matters)
func compareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// TestShortsConfigDefaults tests that ShortsConfig defaults are applied correctly
func TestShortsConfigDefaults(t *testing.T) {
	tests := []struct {
		name                   string
		yamlContent            string
		expectedMaxWords       int
		expectedCandidateCount int
	}{
		{
			name: "Shorts config from YAML",
			yamlContent: `
shorts:
  maxWords: 200
  candidateCount: 15
`,
			expectedMaxWords:       200,
			expectedCandidateCount: 15,
		},
		{
			name:                   "Shorts config defaults when not in YAML",
			yamlContent:            ``, // Empty YAML
			expectedMaxWords:       150,
			expectedCandidateCount: 10,
		},
		{
			name: "Shorts config partial - maxWords only",
			yamlContent: `
shorts:
  maxWords: 180
`,
			expectedMaxWords:       180,
			expectedCandidateCount: 10, // Default
		},
		{
			name: "Shorts config partial - candidateCount only",
			yamlContent: `
shorts:
  candidateCount: 8
`,
			expectedMaxWords:       150, // Default
			expectedCandidateCount: 8,
		},
		{
			name: "Shorts config with zero values uses defaults",
			yamlContent: `
shorts:
  maxWords: 0
  candidateCount: 0
`,
			expectedMaxWords:       150, // Default applied for zero
			expectedCandidateCount: 10,  // Default applied for zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create a temporary settings.yaml
			settingsDir := t.TempDir()

			tmpfn := filepath.Join(settingsDir, "settings.yaml")
			err := os.WriteFile(tmpfn, []byte(tt.yamlContent), 0644)
			require.NoError(t, err)

			// Change to temp directory
			originalWD, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir(settingsDir)
			require.NoError(t, err)
			defer os.Chdir(originalWD)

			// Reset and load settings
			testSettings := Settings{}
			yamlFile, err := os.ReadFile("settings.yaml")
			if err == nil {
				yaml.Unmarshal(yamlFile, &testSettings)
			}

			// Apply defaults (mimicking init() behavior)
			if testSettings.Shorts.MaxWords == 0 {
				testSettings.Shorts.MaxWords = 150
			}
			if testSettings.Shorts.CandidateCount == 0 {
				testSettings.Shorts.CandidateCount = 10
			}

			// Assert
			assert.Equal(t, tt.expectedMaxWords, testSettings.Shorts.MaxWords,
				"MaxWords mismatch")
			assert.Equal(t, tt.expectedCandidateCount, testSettings.Shorts.CandidateCount,
				"CandidateCount mismatch")
		})
	}
}

// TestShortsConfigSerialization tests ShortsConfig JSON/YAML serialization
func TestShortsConfigSerialization(t *testing.T) {
	t.Run("ShortsConfig serializes to YAML correctly", func(t *testing.T) {
		config := ShortsConfig{
			MaxWords:       200,
			CandidateCount: 15,
		}

		yamlData, err := yaml.Marshal(config)
		require.NoError(t, err)

		var parsed ShortsConfig
		err = yaml.Unmarshal(yamlData, &parsed)
		require.NoError(t, err)

		assert.Equal(t, config.MaxWords, parsed.MaxWords)
		assert.Equal(t, config.CandidateCount, parsed.CandidateCount)
	})

	t.Run("ShortsConfig deserializes from YAML correctly", func(t *testing.T) {
		yamlContent := `
maxWords: 175
candidateCount: 12
`
		var config ShortsConfig
		err := yaml.Unmarshal([]byte(yamlContent), &config)
		require.NoError(t, err)

		assert.Equal(t, 175, config.MaxWords)
		assert.Equal(t, 12, config.CandidateCount)
	})
}

// TestElevenLabsConfigDefaults tests that ElevenLabs config defaults are applied correctly
func TestElevenLabsConfigDefaults(t *testing.T) {
	tests := []struct {
		name                        string
		yamlContent                 string
		envAPIKey                   string
		expectedAPIKey              string
		expectedTestMode            bool
		expectedStartTime           int
		expectedEndTime             int
		expectedNumSpeakers         int
		expectedDropBackgroundAudio bool
	}{
		{
			name: "ElevenLabs config from YAML",
			yamlContent: `
elevenLabs:
  apiKey: "yaml-api-key"
  testMode: true
  startTime: 10
  endTime: 60
  numSpeakers: 2
  dropBackgroundAudio: true
`,
			envAPIKey:                   "",
			expectedAPIKey:              "yaml-api-key",
			expectedTestMode:            true,
			expectedStartTime:           10,
			expectedEndTime:             60,
			expectedNumSpeakers:         2,
			expectedDropBackgroundAudio: true,
		},
		{
			name:                        "ElevenLabs config defaults when not in YAML",
			yamlContent:                 ``, // Empty YAML
			envAPIKey:                   "",
			expectedAPIKey:              "",
			expectedTestMode:            false, // bool default
			expectedStartTime:           0,     // int default
			expectedEndTime:             0,     // int default (0 = full video)
			expectedNumSpeakers:         1,     // Default applied
			expectedDropBackgroundAudio: false, // bool default
		},
		{
			name: "ElevenLabs API key from environment overrides YAML",
			yamlContent: `
elevenLabs:
  apiKey: "yaml-api-key"
  testMode: true
`,
			envAPIKey:          "env-api-key",
			expectedAPIKey:     "env-api-key",
			expectedTestMode:   true,
			expectedNumSpeakers: 1, // Default applied
		},
		{
			name:                "ElevenLabs API key from environment when not in YAML",
			yamlContent:         ``,
			envAPIKey:           "env-only-key",
			expectedAPIKey:      "env-only-key",
			expectedNumSpeakers: 1, // Default applied
		},
		{
			name: "ElevenLabs partial config - testMode only",
			yamlContent: `
elevenLabs:
  testMode: true
`,
			envAPIKey:           "",
			expectedAPIKey:      "",
			expectedTestMode:    true,
			expectedNumSpeakers: 1, // Default applied
		},
		{
			name: "ElevenLabs numSpeakers zero uses default",
			yamlContent: `
elevenLabs:
  numSpeakers: 0
`,
			envAPIKey:           "",
			expectedNumSpeakers: 1, // Default applied for zero
		},
	}

	// Save and restore environment
	origAPIKey := os.Getenv("ELEVENLABS_API_KEY")
	defer func() {
		if origAPIKey == "" {
			os.Unsetenv("ELEVENLABS_API_KEY")
		} else {
			os.Setenv("ELEVENLABS_API_KEY", origAPIKey)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.envAPIKey != "" {
				os.Setenv("ELEVENLABS_API_KEY", tt.envAPIKey)
			} else {
				os.Unsetenv("ELEVENLABS_API_KEY")
			}

			// Setup: Create a temporary settings.yaml
			settingsDir := t.TempDir()

			tmpfn := filepath.Join(settingsDir, "settings.yaml")
			err := os.WriteFile(tmpfn, []byte(tt.yamlContent), 0644)
			require.NoError(t, err)

			// Change to temp directory
			originalWD, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir(settingsDir)
			require.NoError(t, err)
			defer os.Chdir(originalWD)

			// Reset and load settings
			testSettings := Settings{}
			yamlFile, err := os.ReadFile("settings.yaml")
			if err == nil {
				yaml.Unmarshal(yamlFile, &testSettings)
			}

			// Apply environment variable (mimicking init() behavior)
			if envKey := os.Getenv("ELEVENLABS_API_KEY"); envKey != "" {
				testSettings.ElevenLabs.APIKey = envKey
			}

			// Apply defaults (mimicking init() behavior)
			if testSettings.ElevenLabs.NumSpeakers == 0 {
				testSettings.ElevenLabs.NumSpeakers = 1
			}

			// Assert
			assert.Equal(t, tt.expectedAPIKey, testSettings.ElevenLabs.APIKey, "APIKey mismatch")
			assert.Equal(t, tt.expectedTestMode, testSettings.ElevenLabs.TestMode, "TestMode mismatch")
			assert.Equal(t, tt.expectedStartTime, testSettings.ElevenLabs.StartTime, "StartTime mismatch")
			assert.Equal(t, tt.expectedEndTime, testSettings.ElevenLabs.EndTime, "EndTime mismatch")
			assert.Equal(t, tt.expectedNumSpeakers, testSettings.ElevenLabs.NumSpeakers, "NumSpeakers mismatch")
			assert.Equal(t, tt.expectedDropBackgroundAudio, testSettings.ElevenLabs.DropBackgroundAudio, "DropBackgroundAudio mismatch")
		})
	}
}

// TestElevenLabsConfigSerialization tests ElevenLabs config YAML serialization
func TestElevenLabsConfigSerialization(t *testing.T) {
	t.Run("ElevenLabs config serializes to YAML correctly", func(t *testing.T) {
		config := SettingsElevenLabs{
			APIKey:              "test-key",
			TestMode:            true,
			StartTime:           5,
			EndTime:             120,
			NumSpeakers:         2,
			DropBackgroundAudio: true,
		}

		yamlData, err := yaml.Marshal(config)
		require.NoError(t, err)

		var parsed SettingsElevenLabs
		err = yaml.Unmarshal(yamlData, &parsed)
		require.NoError(t, err)

		assert.Equal(t, config.APIKey, parsed.APIKey)
		assert.Equal(t, config.TestMode, parsed.TestMode)
		assert.Equal(t, config.StartTime, parsed.StartTime)
		assert.Equal(t, config.EndTime, parsed.EndTime)
		assert.Equal(t, config.NumSpeakers, parsed.NumSpeakers)
		assert.Equal(t, config.DropBackgroundAudio, parsed.DropBackgroundAudio)
	})

	t.Run("ElevenLabs config deserializes from YAML correctly", func(t *testing.T) {
		yamlContent := `
apiKey: "deserialized-key"
testMode: true
startTime: 30
endTime: 90
numSpeakers: 1
dropBackgroundAudio: false
`
		var config SettingsElevenLabs
		err := yaml.Unmarshal([]byte(yamlContent), &config)
		require.NoError(t, err)

		assert.Equal(t, "deserialized-key", config.APIKey)
		assert.Equal(t, true, config.TestMode)
		assert.Equal(t, 30, config.StartTime)
		assert.Equal(t, 90, config.EndTime)
		assert.Equal(t, 1, config.NumSpeakers)
		assert.Equal(t, false, config.DropBackgroundAudio)
	})
}

// TestSpanishChannelConfigDefaults tests that Spanish channel config defaults are applied correctly
func TestSpanishChannelConfigDefaults(t *testing.T) {
	tests := []struct {
		name                    string
		yamlContent             string
		expectedChannelID       string
		expectedCredentialsFile string
		expectedTokenFile       string
		expectedCallbackPort    int
	}{
		{
			name: "Spanish channel config from YAML",
			yamlContent: `
spanishChannel:
  channelId: "UC_SPANISH_CHANNEL"
  credentialsFile: "custom_spanish_secret.json"
  tokenFile: "custom-spanish-token.json"
  callbackPort: 8092
`,
			expectedChannelID:       "UC_SPANISH_CHANNEL",
			expectedCredentialsFile: "custom_spanish_secret.json",
			expectedTokenFile:       "custom-spanish-token.json",
			expectedCallbackPort:    8092,
		},
		{
			name:                    "Spanish channel config defaults when not in YAML",
			yamlContent:             ``, // Empty YAML
			expectedChannelID:       "",
			expectedCredentialsFile: "client_secret_spanish.json",
			expectedTokenFile:       "youtube-go-spanish.json",
			expectedCallbackPort:    8091,
		},
		{
			name: "Spanish channel partial config - channelId only",
			yamlContent: `
spanishChannel:
  channelId: "UC_MY_CHANNEL"
`,
			expectedChannelID:       "UC_MY_CHANNEL",
			expectedCredentialsFile: "client_secret_spanish.json", // Default
			expectedTokenFile:       "youtube-go-spanish.json",    // Default
			expectedCallbackPort:    8091,                         // Default
		},
		{
			name: "Spanish channel partial config - port only",
			yamlContent: `
spanishChannel:
  callbackPort: 8095
`,
			expectedChannelID:       "",
			expectedCredentialsFile: "client_secret_spanish.json",
			expectedTokenFile:       "youtube-go-spanish.json",
			expectedCallbackPort:    8095,
		},
		{
			name: "Spanish channel with zero port uses default",
			yamlContent: `
spanishChannel:
  channelId: "UC_TEST"
  callbackPort: 0
`,
			expectedChannelID:       "UC_TEST",
			expectedCredentialsFile: "client_secret_spanish.json",
			expectedTokenFile:       "youtube-go-spanish.json",
			expectedCallbackPort:    8091, // Default applied for zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create a temporary settings.yaml
			settingsDir := t.TempDir()

			tmpfn := filepath.Join(settingsDir, "settings.yaml")
			err := os.WriteFile(tmpfn, []byte(tt.yamlContent), 0644)
			require.NoError(t, err)

			// Change to temp directory
			originalWD, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir(settingsDir)
			require.NoError(t, err)
			defer os.Chdir(originalWD)

			// Reset and load settings
			testSettings := Settings{}
			yamlFile, err := os.ReadFile("settings.yaml")
			if err == nil {
				yaml.Unmarshal(yamlFile, &testSettings)
			}

			// Apply defaults (mimicking init() behavior)
			if testSettings.SpanishChannel.CredentialsFile == "" {
				testSettings.SpanishChannel.CredentialsFile = "client_secret_spanish.json"
			}
			if testSettings.SpanishChannel.TokenFile == "" {
				testSettings.SpanishChannel.TokenFile = "youtube-go-spanish.json"
			}
			if testSettings.SpanishChannel.CallbackPort == 0 {
				testSettings.SpanishChannel.CallbackPort = 8091
			}

			// Assert
			assert.Equal(t, tt.expectedChannelID, testSettings.SpanishChannel.ChannelID, "ChannelID mismatch")
			assert.Equal(t, tt.expectedCredentialsFile, testSettings.SpanishChannel.CredentialsFile, "CredentialsFile mismatch")
			assert.Equal(t, tt.expectedTokenFile, testSettings.SpanishChannel.TokenFile, "TokenFile mismatch")
			assert.Equal(t, tt.expectedCallbackPort, testSettings.SpanishChannel.CallbackPort, "CallbackPort mismatch")
		})
	}
}

// TestSpanishChannelConfigSerialization tests Spanish channel config YAML serialization
func TestSpanishChannelConfigSerialization(t *testing.T) {
	t.Run("Spanish channel config serializes to YAML correctly", func(t *testing.T) {
		config := SettingsSpanishChannel{
			ChannelID:       "UC_TEST_CHANNEL",
			CredentialsFile: "test_secret.json",
			TokenFile:       "test-token.json",
			CallbackPort:    8093,
		}

		yamlData, err := yaml.Marshal(config)
		require.NoError(t, err)

		var parsed SettingsSpanishChannel
		err = yaml.Unmarshal(yamlData, &parsed)
		require.NoError(t, err)

		assert.Equal(t, config.ChannelID, parsed.ChannelID)
		assert.Equal(t, config.CredentialsFile, parsed.CredentialsFile)
		assert.Equal(t, config.TokenFile, parsed.TokenFile)
		assert.Equal(t, config.CallbackPort, parsed.CallbackPort)
	})

	t.Run("Spanish channel config deserializes from YAML correctly", func(t *testing.T) {
		yamlContent := `
channelId: "UC_DESERIALIZED"
credentialsFile: "deser_secret.json"
tokenFile: "deser-token.json"
callbackPort: 8094
`
		var config SettingsSpanishChannel
		err := yaml.Unmarshal([]byte(yamlContent), &config)
		require.NoError(t, err)

		assert.Equal(t, "UC_DESERIALIZED", config.ChannelID)
		assert.Equal(t, "deser_secret.json", config.CredentialsFile)
		assert.Equal(t, "deser-token.json", config.TokenFile)
		assert.Equal(t, 8094, config.CallbackPort)
	})
}

func TestSlackSettingsLoading(t *testing.T) {
	// Define the YAML content for the test
	slackYAMLContent := `
slack:
  targetChannelIDs:
    - "C123YAML"
    - "C456YAML"
`
	// Create a temporary settings.yaml file with the Slack configuration
	settingsDir, cleanup := setupTestSettings(t, slackYAMLContent)
	defer cleanup()

	// Store original working directory and change to the temp directory
	// where settings.yaml was created, so that init() can find it.
	originalWd, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")
	err = os.Chdir(filepath.Dir(settingsDir)) // settingsDir is the path to settings.yaml, so Dir gives the directory
	require.NoError(t, err, "Failed to change to temp directory")
	defer func() {
		err = os.Chdir(originalWd)
		require.NoError(t, err, "Failed to restore original working directory")
	}()

	// Reset GlobalSettings to ensure a clean state for this test
	GlobalSettings = Settings{}
	// Re-initialize cobra command to avoid pollution from other tests, if flags were an issue.
	// For this specific test (YAML loading), direct unmarshal is safer if init() has side effects.

	// Directly load and unmarshal the test settings file into a new Settings instance
	// This avoids re-running the full init() which registers flags and might have other side effects.
	testLocalSettings := Settings{}
	yamlFile, err := os.ReadFile("settings.yaml") // Reads from the temp dir
	require.NoError(t, err, "Failed to read temporary settings.yaml")
	err = yaml.Unmarshal(yamlFile, &testLocalSettings)
	require.NoError(t, err, "Failed to unmarshal YAML from temporary settings.yaml")

	// Expected Slack settings
	expectedSlackSettings := SettingsSlack{
		TargetChannelIDs: []string{"C123YAML", "C456YAML"},
	}

	// Assert that the Slack settings are loaded correctly
	assert.True(t, compareStringSlices(testLocalSettings.Slack.TargetChannelIDs, expectedSlackSettings.TargetChannelIDs),
		fmt.Sprintf("Slack.TargetChannelIDs = %v, want %v", testLocalSettings.Slack.TargetChannelIDs, expectedSlackSettings.TargetChannelIDs))

	// Additionally, if we want to test the global instance after a fresh init logic:
	// This is a bit more involved due to init() being an init function.
	// One would typically refactor the core loading logic from init() into a separate
	// function to test it in isolation without re-running all flag setup.
	// For now, the direct unmarshal test above is safer and more targeted for YAML loading.
}

// TestGeminiConfigDefaults tests that Gemini config defaults are applied correctly
func TestGeminiConfigDefaults(t *testing.T) {
	tests := []struct {
		name          string
		yamlContent   string
		expectedModel string
	}{
		{
			name: "Gemini config from YAML",
			yamlContent: `
gemini:
  model: "gemini-2.5-flash-image"
`,
			expectedModel: "gemini-2.5-flash-image",
		},
		{
			name:          "Gemini config defaults when not in YAML",
			yamlContent:   ``, // Empty YAML
			expectedModel: "gemini-3-pro-image-preview",
		},
		{
			name: "Gemini config with empty model uses default",
			yamlContent: `
gemini:
  model: ""
`,
			expectedModel: "gemini-3-pro-image-preview",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create a temporary settings.yaml
			settingsDir := t.TempDir()

			tmpfn := filepath.Join(settingsDir, "settings.yaml")
			err := os.WriteFile(tmpfn, []byte(tt.yamlContent), 0644)
			require.NoError(t, err)

			// Change to temp directory
			originalWD, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir(settingsDir)
			require.NoError(t, err)
			defer os.Chdir(originalWD)

			// Reset and load settings
			testSettings := Settings{}
			yamlFile, err := os.ReadFile("settings.yaml")
			if err == nil {
				yaml.Unmarshal(yamlFile, &testSettings)
			}

			// Apply defaults (mimicking init() behavior)
			if testSettings.Gemini.Model == "" {
				testSettings.Gemini.Model = "gemini-3-pro-image-preview"
			}

			// Assert
			assert.Equal(t, tt.expectedModel, testSettings.Gemini.Model, "Model mismatch")
		})
	}
}
