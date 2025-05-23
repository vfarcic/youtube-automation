package slack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	// Ensure yaml.v3 is in go.mod if not already, or tests might fail on CI
)

const testDir = "test_config_data"

// Helper to create a temporary YAML config file for testing
func createTestYAML(t *testing.T, content string) string {
	t.Helper()
	err := os.MkdirAll(testDir, 0755)
	assert.NoError(t, err)

	tempFile, err := os.CreateTemp(testDir, "settings-*.yaml")
	assert.NoError(t, err)

	_, err = tempFile.Write([]byte(content))
	assert.NoError(t, err)
	err = tempFile.Close()
	assert.NoError(t, err)

	return tempFile.Name()
}

// Helper to clean up test files and directory
func cleanupTestFiles(t *testing.T) {
	t.Helper()
	err := os.RemoveAll(testDir)
	if err != nil {
		// Log the error but don't fail the test, as it's a cleanup step
		t.Logf("Warning: failed to remove test config directory %s: %v", testDir, err)
	}
}

// Helper to set environment variables for a test and restore them afterwards
func withEnv(t *testing.T, keyvals map[string]string, testFunc func()) {
	t.Helper()
	originalEnv := make(map[string]string)

	for key, value := range keyvals {
		originalEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	testFunc()

	for key, value := range originalEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.Equal(t, 1, cfg.RetryBackoffSeconds)
	assert.Empty(t, cfg.Token, "Token should be empty in default config")
}

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedCfg SlackConfig
		expectError bool
	}{
		{
			name: "All env vars set",
			envVars: map[string]string{
				"SLACK_API_TOKEN":             "env-token",
				"SLACK_RETRY_ATTEMPTS":        "5",
				"SLACK_RETRY_BACKOFF_SECONDS": "10",
			},
			expectedCfg: SlackConfig{
				Token:               "env-token",
				RetryAttempts:       5,
				RetryBackoffSeconds: 10,
			},
			expectError: false,
		},
		{
			name: "Partial env vars",
			envVars: map[string]string{
				"SLACK_API_TOKEN": "partial-token",
			},
			expectedCfg: SlackConfig{
				Token:               "partial-token",
				RetryAttempts:       3,
				RetryBackoffSeconds: 1,
			},
			expectError: false,
		},
		{
			name: "Malformed retry attempts",
			envVars: map[string]string{
				"SLACK_RETRY_ATTEMPTS": "not-a-number",
			},
			expectError: true,
		},
		{
			name: "Malformed retry backoff seconds",
			envVars: map[string]string{
				"SLACK_RETRY_BACKOFF_SECONDS": "invalid",
			},
			expectError: true,
		},
		{
			name:    "Empty category channel value",
			envVars: map[string]string{
				// "SLACK_CATEGORY_CHANNELS_EMPTY": " ,,,",
			},
			expectedCfg: SlackConfig{
				Token:               "",
				RetryAttempts:       3,
				RetryBackoffSeconds: 1,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultConfig()
			withEnv(t, tt.envVars, func() {
				err := cfg.LoadFromEnv()
				if tt.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedCfg.Token, cfg.Token)
					assert.Equal(t, tt.expectedCfg.RetryAttempts, cfg.RetryAttempts)
					assert.Equal(t, tt.expectedCfg.RetryBackoffSeconds, cfg.RetryBackoffSeconds)
				}
			})
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	defer cleanupTestFiles(t)

	tests := []struct {
		name        string
		yamlContent string
		expectedCfg SlackConfig
		expectError bool
	}{
		{
			name: "Full config file",
			yamlContent: `
slack:
  token: "file-token-should-be-ignored"
  retryAttempts: 7
  retryBackoffSeconds: 12
`,
			expectedCfg: SlackConfig{
				Token:               "",
				RetryAttempts:       7,
				RetryBackoffSeconds: 12,
			},
			expectError: false,
		},
		{
			name: "Partial config file",
			yamlContent: `
slack:
  retryAttempts: 5
`,
			expectedCfg: SlackConfig{
				Token:               "",
				RetryAttempts:       5,
				RetryBackoffSeconds: 1,
			},
			expectError: false,
		},
		{
			name: "File with only token (should be ignored)",
			yamlContent: `
slack:
  token: "file-token-secret"
`,
			expectedCfg: SlackConfig{
				Token:               "",
				RetryAttempts:       3,
				RetryBackoffSeconds: 1,
			},
			expectError: false,
		},
		{
			name:        "File not found",
			yamlContent: "",
			expectedCfg: *NewDefaultConfig(),
			expectError: false,
		},
		{
			name:        "Malformed YAML",
			yamlContent: `slack: { token: "bad yaml`,
			expectError: true,
		},
		{
			name:        "Empty file",
			yamlContent: ``,
			expectedCfg: *NewDefaultConfig(),
			expectError: false,
		},
		{
			name:        "YAML without slack key",
			yamlContent: `otherkey: value`,
			expectedCfg: *NewDefaultConfig(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultConfig()
			var filePath string
			if tt.name == "File not found" {
				filePath = filepath.Join(testDir, "non_existent_settings.yaml")
			} else {
				filePath = createTestYAML(t, tt.yamlContent)
			}

			err := cfg.LoadFromFile(filePath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCfg.Token, cfg.Token)
				assert.Equal(t, tt.expectedCfg.RetryAttempts, cfg.RetryAttempts)
				assert.Equal(t, tt.expectedCfg.RetryBackoffSeconds, cfg.RetryBackoffSeconds)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         SlackConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid config",
			cfg: SlackConfig{
				Token:               "valid-token",
				RetryAttempts:       3,
				RetryBackoffSeconds: 1,
			},
			expectError: false,
		},
		{
			name: "Missing token",
			cfg: SlackConfig{
				RetryAttempts:       3,
				RetryBackoffSeconds: 1,
			},
			expectError: true,
			errorMsg:    "slack API token (SLACK_API_TOKEN) is required",
		},
		{
			name: "Invalid retry attempts",
			cfg: SlackConfig{
				Token:               "valid-token",
				RetryAttempts:       -1,
				RetryBackoffSeconds: 1,
			},
			expectError: true,
			errorMsg:    "retry attempts must be non-negative",
		},
		{
			name: "Invalid retry backoff seconds",
			cfg: SlackConfig{
				Token:               "valid-token",
				RetryAttempts:       3,
				RetryBackoffSeconds: 0,
			},
			expectError: true,
			errorMsg:    "retry backoff seconds must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadAndValidateSlackConfig(t *testing.T) {
	defer cleanupTestFiles(t)

	tests := []struct {
		name           string
		yamlContent    string
		envVars        map[string]string
		configFilePath string
		expectError    bool
		errorContains  string
		expected       *SlackConfig
	}{
		{
			name: "Valid: File and Env",
			yamlContent: `
slack:
  retryAttempts: 5
`,
			envVars: map[string]string{
				"SLACK_API_TOKEN":      "env-token-is-king",
				"SLACK_RETRY_ATTEMPTS": "7",
			},
			expected: &SlackConfig{
				Token:               "env-token-is-king",
				RetryAttempts:       7,
				RetryBackoffSeconds: 1,
			},
		},
		{
			name: "Valid: Only Env, no file",
			envVars: map[string]string{
				"SLACK_API_TOKEN":             "env-only-token",
				"SLACK_RETRY_ATTEMPTS":        "10",
				"SLACK_RETRY_BACKOFF_SECONDS": "20",
			},
			configFilePath: filepath.Join(testDir, "no-such-file.yaml"),
			expected: &SlackConfig{
				Token:               "env-only-token",
				RetryAttempts:       10,
				RetryBackoffSeconds: 20,
			},
		},
		{
			name: "Valid: Only File, no env override for file values (except token)",
			yamlContent: `
slack:
  token: "file-token-ignored"
  retryAttempts: 2
`,
			envVars: map[string]string{
				"SLACK_API_TOKEN": "env-token-must-be-present",
			},
			expected: &SlackConfig{
				Token:               "env-token-must-be-present",
				RetryAttempts:       2,
				RetryBackoffSeconds: 1,
			},
		},
		{
			name:           "Invalid: Missing Token (no file, no env)",
			envVars:        map[string]string{},
			configFilePath: filepath.Join(testDir, "no-such-file-either.yaml"),
			expectError:    true,
			errorContains:  "slack API token (SLACK_API_TOKEN) is required",
		},
		{
			name:          "Invalid: Malformed YAML file",
			yamlContent:   `slack: { token: "bad`,
			envVars:       map[string]string{"SLACK_API_TOKEN": "token"},
			expectError:   true,
			errorContains: "failed to parse YAML config file",
		},
		{
			name: "Invalid: Malformed Env Var for number",
			envVars: map[string]string{
				"SLACK_API_TOKEN":      "token",
				"SLACK_RETRY_ATTEMPTS": "not-an-int",
			},
			expectError:   true,
			errorContains: "invalid SLACK_RETRY_ATTEMPTS value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.configFilePath
			if filePath == "" && tt.yamlContent != "" {
				filePath = createTestYAML(t, tt.yamlContent)
			} else if filePath == "" {
				filePath = createTestYAML(t, "")
			}

			originalGlobal := GlobalSlackConfig
			defer func() { GlobalSlackConfig = originalGlobal }()

			withEnv(t, tt.envVars, func() {
				err := LoadAndValidateSlackConfig(filePath)
				if tt.expectError {
					assert.Error(t, err)
					if tt.errorContains != "" {
						assert.True(t, strings.Contains(err.Error(), tt.errorContains), "Error message '%s' did not contain '%s'", err.Error(), tt.errorContains)
					}
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, GlobalSlackConfig, "GlobalSlackConfig should be populated")
					assert.Equal(t, tt.expected.Token, GlobalSlackConfig.Token)
					assert.Equal(t, tt.expected.RetryAttempts, GlobalSlackConfig.RetryAttempts)
					assert.Equal(t, tt.expected.RetryBackoffSeconds, GlobalSlackConfig.RetryBackoffSeconds)
				}
			})
		})
	}
}
