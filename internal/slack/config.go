package slack

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3" // Using v3 for better error handling and features
)

// SlackConfig holds configuration parameters for Slack integration, primarily token and retry settings.
// Channel management is handled globally.
type SlackConfig struct {
	Token string `yaml:"token"`
	// DefaultChannel      string              `yaml:"defaultChannel"` // Removed
	// CategoryChannels    map[string][]string `yaml:"categoryChannels"` // Removed
	RetryAttempts       int `yaml:"retryAttempts"`
	RetryBackoffSeconds int `yaml:"retryBackoffSeconds"`
}

// GlobalSlackConfig stores the loaded and validated Slack configuration.
var GlobalSlackConfig = NewDefaultConfig()

// NewDefaultConfig creates a config with sensible defaults for retry settings.
// Token is expected to be loaded from environment or file.
func NewDefaultConfig() *SlackConfig {
	return &SlackConfig{
		// DefaultChannel:      "general", // Removed
		// CategoryChannels:    make(map[string][]string), // Removed
		RetryAttempts:       3,
		RetryBackoffSeconds: 1,
	}
}

// LoadFromEnv loads configuration from environment variables, overriding existing values.
func (c *SlackConfig) LoadFromEnv() error {
	if token := os.Getenv("SLACK_API_TOKEN"); token != "" {
		c.Token = token
	}

	// Removed loading for SLACK_DEFAULT_CHANNEL
	// Removed loading for SLACK_CATEGORY_CHANNELS_*

	if retryAttemptsStr := os.Getenv("SLACK_RETRY_ATTEMPTS"); retryAttemptsStr != "" {
		if retryAttempts, err := strconv.Atoi(retryAttemptsStr); err == nil {
			c.RetryAttempts = retryAttempts
		} else {
			return fmt.Errorf("invalid SLACK_RETRY_ATTEMPTS value '%s': %w", retryAttemptsStr, err)
		}
	}

	if retryBackoffSecondsStr := os.Getenv("SLACK_RETRY_BACKOFF_SECONDS"); retryBackoffSecondsStr != "" {
		if retryBackoffSeconds, err := strconv.Atoi(retryBackoffSecondsStr); err == nil {
			c.RetryBackoffSeconds = retryBackoffSeconds
		} else {
			return fmt.Errorf("invalid SLACK_RETRY_BACKOFF_SECONDS value '%s': %w", retryBackoffSecondsStr, err)
		}
	}
	return nil
}

// LoadFromFile loads configuration from a YAML file.
// It expects the YAML to have a top-level 'slack:' key.
func (c *SlackConfig) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No config file found, not an error itself.
		}
		return fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	var configFile struct {
		Slack SlackConfig `yaml:"slack"`
	}

	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return fmt.Errorf("failed to parse YAML config file '%s': %w", path, err)
	}

	// Apply non-confidential settings from the file. Token is typically from env.
	// DefaultChannel and CategoryChannels loading removed.

	if configFile.Slack.RetryAttempts != 0 { // Consider if 0 is a valid override
		c.RetryAttempts = configFile.Slack.RetryAttempts
	}
	if configFile.Slack.RetryBackoffSeconds != 0 { // Consider if 0 is a valid override
		c.RetryBackoffSeconds = configFile.Slack.RetryBackoffSeconds
	}
	// If token can also come from file (though env is usually preferred for secrets):
	// if configFile.Slack.Token != "" {
	// 	c.Token = configFile.Slack.Token
	// }

	return nil
}

// Validate checks if the configuration is valid.
func (c *SlackConfig) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("slack API token (SLACK_API_TOKEN) is required")
	}
	// Removed validation for DefaultChannel
	if c.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts must be non-negative")
	}
	if c.RetryBackoffSeconds <= 0 {
		return fmt.Errorf("retry backoff seconds must be positive")
	}
	return nil
}

// LoadAndValidateSlackConfig loads configuration from a specified file path and environment variables,
// then validates it. It updates the GlobalSlackConfig.
// The order is: Defaults -> File -> Environment Variables -> Validation.
func LoadAndValidateSlackConfig(configFilePath string) error {
	GlobalSlackConfig = NewDefaultConfig() // Start with defaults

	if configFilePath != "" {
		if err := GlobalSlackConfig.LoadFromFile(configFilePath); err != nil {
			return fmt.Errorf("error loading slack config from file '%s': %w", configFilePath, err)
		}
	}

	if err := GlobalSlackConfig.LoadFromEnv(); err != nil {
		return fmt.Errorf("error loading slack config from environment: %w", err)
	}

	if err := GlobalSlackConfig.Validate(); err != nil {
		return fmt.Errorf("slack configuration validation failed: %w", err)
	}
	return nil
}
