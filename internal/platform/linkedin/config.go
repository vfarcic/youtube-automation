package linkedin

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config holds the configuration for LinkedIn API
type Config struct {
	AccessToken string
	APIUrl      string
}

// DefaultAPIURL is the default LinkedIn API endpoint
const DefaultAPIURL = "https://api.linkedin.com/v2"

// NewConfig creates a new LinkedIn configuration with the provided access token
func NewConfig(accessToken string) *Config {
	return &Config{
		AccessToken: accessToken,
		APIUrl:      DefaultAPIURL,
	}
}

// LoadConfigFromEnv loads LinkedIn configuration from environment variables
func LoadConfigFromEnv() (*Config, error) {
	accessToken := os.Getenv("LINKEDIN_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("LINKEDIN_ACCESS_TOKEN environment variable is not set")
	}

	config := NewConfig(accessToken)

	// Optional: Override API URL if specified
	if apiURL := os.Getenv("LINKEDIN_API_URL"); apiURL != "" {
		config.APIUrl = apiURL
	}

	return config, nil
}

// LoadConfigFromYAML loads LinkedIn configuration from settings YAML
func LoadConfigFromYAML(settings map[string]interface{}) (*Config, error) {
	linkedinSettings, ok := settings["linkedin"].(map[string]interface{})
	if !ok {
		return nil, errors.New("linkedin section not found in settings")
	}

	accessToken, ok := linkedinSettings["accessToken"].(string)
	if !ok || accessToken == "" {
		return nil, errors.New("linkedin.accessToken not found in settings")
	}

	config := NewConfig(accessToken)

	// Optional: Override API URL if specified
	if apiURL, ok := linkedinSettings["apiUrl"].(string); ok && apiURL != "" {
		config.APIUrl = apiURL
	}

	return config, nil
}

// ValidateConfig validates the LinkedIn configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return errors.New("LinkedIn configuration is nil")
	}

	if config.AccessToken == "" {
		return errors.New("LinkedIn access token cannot be empty")
	}

	// Basic token format validation
	if !strings.HasPrefix(config.AccessToken, "AQ") {
		return fmt.Errorf("LinkedIn access token has invalid format (should start with 'AQ')")
	}

	if len(config.AccessToken) < 20 {
		return errors.New("LinkedIn access token is too short")
	}

	if config.APIUrl == "" {
		return errors.New("LinkedIn API URL cannot be empty")
	}

	return nil
}