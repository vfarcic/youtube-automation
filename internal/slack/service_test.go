package slack

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a valid minimal SlackConfig for testing
func newTestSlackConfig() *SlackConfig {
	return &SlackConfig{
		Token:               "xoxb-test-token-valid-for-newslackauth", // Needs to pass NewSlackAuth
		RetryAttempts:       3,
		RetryBackoffSeconds: 2,
	}
}

func TestNewSlackService(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := newTestSlackConfig()

		originalNewSlackAuth := NewSlackAuth
		NewSlackAuth = func(token string) (*SlackAuth, error) {
			if token == "" {
				return nil, fmt.Errorf("mocked NewSlackAuth: token is empty")
			}
			return &SlackAuth{Token: token}, nil
		}
		defer func() { NewSlackAuth = originalNewSlackAuth }()

		originalNewSlackClient := NewSlackClient
		NewSlackClient = func(auth *SlackAuth, opts ...ClientOption) (*SlackClient, error) {
			if auth == nil {
				return nil, fmt.Errorf("mocked NewSlackClient: auth is nil")
			}
			return &SlackClient{auth: auth}, nil // Return a minimal client
		}
		defer func() { NewSlackClient = originalNewSlackClient }()

		service, err := NewSlackService(cfg)
		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, cfg, service.config)
		require.NotNil(t, service.client)
	})

	t.Run("invalid config - nil", func(t *testing.T) {
		_, err := NewSlackService(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Slack configuration: config is nil")
	})

	t.Run("invalid config - token missing", func(t *testing.T) {
		cfg := newTestSlackConfig()
		cfg.Token = "" // Invalid token
		_, err := NewSlackService(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid Slack configuration")
		assert.Contains(t, err.Error(), "slack API token (SLACK_API_TOKEN) is required")
	})

	t.Run("NewSlackAuth fails", func(t *testing.T) {
		cfg := newTestSlackConfig()

		originalNewSlackAuth := NewSlackAuth
		NewSlackAuth = func(token string) (*SlackAuth, error) {
			return nil, fmt.Errorf("mocked NewSlackAuth failure")
		}
		defer func() { NewSlackAuth = originalNewSlackAuth }()

		_, err := NewSlackService(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create Slack auth")
		assert.Contains(t, err.Error(), "mocked NewSlackAuth failure")
	})

	t.Run("NewSlackClient fails", func(t *testing.T) {
		cfg := newTestSlackConfig()

		originalNewSlackAuth := NewSlackAuth
		NewSlackAuth = func(token string) (*SlackAuth, error) {
			return &SlackAuth{Token: token}, nil
		}
		defer func() { NewSlackAuth = originalNewSlackAuth }()

		originalNewSlackClient := NewSlackClient
		NewSlackClient = func(auth *SlackAuth, opts ...ClientOption) (*SlackClient, error) {
			return nil, fmt.Errorf("mocked NewSlackClient failure")
		}
		defer func() { NewSlackClient = originalNewSlackClient }()

		_, err := NewSlackService(cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create Slack client")
		assert.Contains(t, err.Error(), "mocked NewSlackClient failure")
	})

	t.Run("config validation fails via Validate in NewSlackService", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidSettingsPath := tempDir + "/invalid_settings.yaml"
		invalidConfigContent := `
slack:
`
		err := os.WriteFile(invalidSettingsPath, []byte(invalidConfigContent), 0600)
		require.NoError(t, err)

		cfg := &SlackConfig{} // Token is missing. DefaultChannel removed.

		_, err = NewSlackService(cfg)
		require.Error(t, err, "NewSlackService should fail if config validation fails")
		assert.Contains(t, err.Error(), "invalid Slack configuration", "Error message should indicate config validation failure")
		assert.Contains(t, err.Error(), "slack API token (SLACK_API_TOKEN) is required", "Error message should specify missing token")
	})

}

// Note: To make NewSlackAuth and NewSlackClient truly mockable without
// needing to redefine them in tests, they should be interfaces passed to
// NewSlackService, or NewSlackService should accept them as function parameters.
// The current approach of redefining package-level functions works for this test file
// but isn't ideal for larger systems or concurrent tests.
