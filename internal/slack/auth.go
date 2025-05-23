package slack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	minTokenLength  = 45 // Approximate minimum valid length after prefix
	botTokenPrefix  = "xoxb-"
	userTokenPrefix = "xoxp-"
	requestTimeout  = 5 * time.Second // Example timeout
)

var authTestURL = "https://slack.com/api/auth.test"

// SlackAuth stores Slack authentication details.
type SlackAuth struct {
	Token string
}

// authTestResponse is used to unmarshal the JSON response from Slack's auth.test endpoint.
type authTestResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	Team  string `json:"team,omitempty"`
	User  string `json:"user,omitempty"`
}

// newSlackAuth_actual creates a new SlackAuth instance after validating the token format.
// It checks for a valid prefix ("xoxb-" or "xoxp-") and a minimum overall length.
func newSlackAuth_actual(token string) (*SlackAuth, error) {
	if len(token) < minTokenLength {
		return nil, errors.New("invalid Slack token: token too short")
	}
	if !strings.HasPrefix(token, botTokenPrefix) && !strings.HasPrefix(token, userTokenPrefix) {
		return nil, errors.New("invalid Slack token: must start with '" + botTokenPrefix + "' or '" + userTokenPrefix + "'")
	}
	return &SlackAuth{Token: token}, nil
}

// NewSlackAuth is a function variable for creating a new SlackAuth instance.
// This allows it to be replaced for mocking in tests.
var NewSlackAuth = newSlackAuth_actual

// GetToken returns the stored Slack token.
func (a *SlackAuth) GetToken() string {
	return a.Token
}

// ValidateToken checks if the token is still valid by making a test API call to Slack.
// Note: Standard Slack API tokens (xoxb-, xoxp-) do not expire and do not have a refresh mechanism.
// They are valid until revoked. This function checks if the token is currently active and not revoked.
func (a *SlackAuth) ValidateToken() error {
	if a.Token == "" {
		return errors.New("token is empty")
	}

	client := &http.Client{
		Timeout: requestTimeout,
	}

	req, err := http.NewRequest("GET", authTestURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth.test request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+a.Token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform auth.test request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Consider reading body for more detailed error if Slack provides one here
		return fmt.Errorf("auth.test request failed with status: %s", resp.Status)
	}

	var apiResp authTestResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode auth.test response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("slack token validation failed: %s (Team: %s, User: %s)", apiResp.Error, apiResp.Team, apiResp.User)
	}

	return nil // Token is valid
}

// LoadTokenFromEnv loads a Slack token from the specified environment variable.
func LoadTokenFromEnv(envVarName string) (string, error) {
	token := os.Getenv(envVarName)
	if token == "" {
		return "", fmt.Errorf("Slack token not found in environment variable %q", envVarName)
	}
	return token, nil
}

type slackConfig struct {
	SlackToken string `json:"slack_token"`
}

// LoadTokenFromConfig loads a Slack token from a JSON configuration file.
// The JSON file is expected to have a field "slack_token".
func LoadTokenFromConfig(configPath string) (string, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Slack config file %q: %w", configPath, err)
	}

	var config slackConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to unmarshal Slack config from %q: %w", configPath, err)
	}

	if config.SlackToken == "" {
		return "", fmt.Errorf("slack_token not found or empty in config file %q", configPath)
	}
	return config.SlackToken, nil
}
