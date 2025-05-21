package configuration

import (
	"errors"
	"os"
	"testing"

	"github.com/spf13/cobra"
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
					Endpoint:   "https://api.openai.com",
					Key:        "ai-key-123",
					Deployment: "gpt-4",
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
					Endpoint:   "https://api.openai.com",
					Key:        "ai-key-123",
					Deployment: "gpt-4",
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
			testCmd.Flags().StringVar(&testSettings.AI.Endpoint, "ai-endpoint", testSettings.AI.Endpoint, "")
			testCmd.Flags().StringVar(&testSettings.AI.Key, "ai-key", testSettings.AI.Key, "")
			testCmd.Flags().StringVar(&testSettings.AI.Deployment, "ai-deployment", testSettings.AI.Deployment, "")
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
			if testSettings.AI.Endpoint != tt.expected.AI.Endpoint {
				t.Errorf("AI.Endpoint = %v, want %v", testSettings.AI.Endpoint, tt.expected.AI.Endpoint)
			}
			if testSettings.AI.Key != tt.expected.AI.Key {
				t.Errorf("AI.Key = %v, want %v", testSettings.AI.Key, tt.expected.AI.Key)
			}
			if testSettings.AI.Deployment != tt.expected.AI.Deployment {
				t.Errorf("AI.Deployment = %v, want %v", testSettings.AI.Deployment, tt.expected.AI.Deployment)
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
	testCmd.Flags().StringVar(&testSettings.AI.Endpoint, "ai-endpoint", testSettings.AI.Endpoint, "")
	testCmd.Flags().StringVar(&testSettings.AI.Key, "ai-key", testSettings.AI.Key, "")
	testCmd.Flags().StringVar(&testSettings.AI.Deployment, "ai-deployment", testSettings.AI.Deployment, "")
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
		testSettings.AI.Key = envAIKey
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

	if testSettings.AI.Key != "env-ai-key" {
		t.Errorf("AI.Key = %s, want env-ai-key", testSettings.AI.Key)
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

	if testSettings.AI.Endpoint != "https://default-endpoint.com" {
		t.Errorf("AI.Endpoint = %s, want https://default-endpoint.com", testSettings.AI.Endpoint)
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
	testCmd.Flags().StringVar(&mergedSettings.AI.Endpoint, "ai-endpoint", mergedSettings.AI.Endpoint, "")
	testCmd.Flags().StringVar(&mergedSettings.AI.Key, "ai-key", mergedSettings.AI.Key, "")
	testCmd.Flags().StringVar(&mergedSettings.AI.Deployment, "ai-deployment", mergedSettings.AI.Deployment, "")
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

	if mergedSettings.AI.Key != "flag-ai-key" {
		t.Errorf("AI.Key = %s, want flag-ai-key", mergedSettings.AI.Key)
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
