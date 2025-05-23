package slack

import (
	"errors"
	"net/http"
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	// No separate auth import needed, SlackAuth is in the same package
)

// mockSlackGoClient is a mock implementation of slackGoClientInterface for testing.
type mockSlackGoClient struct {
	PostMessageFunc func(channelID string, options ...slack.MsgOption) (string, string, error)
}

func (m *mockSlackGoClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	if m.PostMessageFunc != nil {
		return m.PostMessageFunc(channelID, options...)
	}
	return "", "", errors.New("PostMessageFunc not implemented in mock")
}

func TestNewSlackClient(t *testing.T) {
	// Explicit setup for the test cases
	const validTestToken1 = "xoxb-01234567890123456789012345678901234567890123"
	const validTestToken2 = "xoxb-abcdefghijklmnopqrstuvwxyzabcdefghijklmnopq"

	validAuthForTest, authErr := NewSlackAuth(validTestToken1)
	if authErr != nil {
		t.Fatalf("Test setup failed for 'Valid auth, no options': NewSlackAuth error: %v", authErr)
	}
	if validAuthForTest == nil {
		t.Fatal("Test setup failed for 'Valid auth, no options': NewSlackAuth returned nil auth")
	}

	validAuthForHttpTest, httpAuthErr := NewSlackAuth(validTestToken2)
	if httpAuthErr != nil {
		t.Fatalf("Test setup failed for 'Valid auth, with HTTP': NewSlackAuth error: %v", httpAuthErr)
	}
	if validAuthForHttpTest == nil { // Added nil check for consistency
		t.Fatal("Test setup failed for 'Valid auth, with HTTP': NewSlackAuth returned nil auth")
	}

	tests := []struct {
		name        string
		auth        *SlackAuth
		opts        []ClientOption
		expectError bool
		errMessage  string
	}{
		{
			name:        "Valid auth, no options",
			auth:        validAuthForTest, // Use explicitly created auth object
			expectError: false,
		},
		{
			name:        "Nil auth",
			auth:        nil,
			expectError: true,
			errMessage:  "cannot create SlackClient: SlackAuth is nil or token is empty",
		},
		{
			name:        "Auth with empty token",
			auth:        &SlackAuth{Token: ""},
			expectError: true,
			errMessage:  "cannot create SlackClient: SlackAuth is nil or token is empty",
		},
		{
			name:        "Valid auth, with HTTP client option",
			auth:        validAuthForHttpTest,
			opts:        []ClientOption{WithHTTPClient(&http.Client{})},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Logf("Running test: %s, auth provided: %+v", tt.name, tt.auth) // For local debugging
			// if tt.auth != nil {
			// 	t.Logf("Token in auth for test %s: '%s'", tt.name, tt.auth.Token) // For local debugging
			// }
			client, err := NewSlackClient(tt.auth, tt.opts...)

			if tt.expectError {
				assert.Error(t, err)
				require.NotNil(t, err, "err should not be nil when expectError is true")
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, client)
			} else {
				// if err != nil {
				// 	t.Logf("Test %s expected no error, but got: %v", tt.name, err) // For local debugging
				// }
				assert.NoError(t, err)
				assert.NotNil(t, client, "client should not be nil when expectError is false")
				if client != nil { // Guard access to client.slackGoClient
					assert.NotNil(t, client.slackGoClient, "client.slackGoClient should not be nil")
					if len(tt.opts) > 0 && client.httpClient == nil {
						t.Errorf("httpClient should be set when options are provided")
					}
				}
			}
		})
	}
}

func TestSlackClient_PostMessage(t *testing.T) {
	const validPostMessageTestToken = "xoxb-98765432109876543210987654321098765432109876"
	validAuth, err := NewSlackAuth(validPostMessageTestToken)
	if err != nil {
		t.Fatalf("Error creating validAuth for PostMessage tests: %v", err)
	}
	if validAuth == nil {
		t.Fatal("validAuth is nil for PostMessage tests after NewSlackAuth")
	}

	tests := []struct {
		name             string
		setupClient      func() *SlackClient
		mockSetup        func(*mockSlackGoClient)
		channelID        string
		options          []slack.MsgOption
		expectChannelID  string
		expectTimestamp  string
		expectError      bool
		expectErrMessage string
	}{
		{
			name: "Successful PostMessage",
			setupClient: func() *SlackClient {
				mockGoClient := &mockSlackGoClient{}
				client, _ := NewSlackClient(validAuth)
				client.slackGoClient = mockGoClient // Inject mock
				return client
			},
			mockSetup: func(m *mockSlackGoClient) {
				m.PostMessageFunc = func(channelID string, options ...slack.MsgOption) (string, string, error) {
					return "C123", "123.456", nil
				}
			},
			channelID:       "C12345",
			options:         []slack.MsgOption{slack.MsgOptionText("Hello", false)},
			expectChannelID: "C123",
			expectTimestamp: "123.456",
			expectError:     false,
		},
		{
			name: "Slack API returns error",
			setupClient: func() *SlackClient {
				mockGoClient := &mockSlackGoClient{}
				client, _ := NewSlackClient(validAuth)
				client.slackGoClient = mockGoClient
				return client
			},
			mockSetup: func(m *mockSlackGoClient) {
				m.PostMessageFunc = func(channelID string, options ...slack.MsgOption) (string, string, error) {
					return "", "", errors.New("slack-api-error")
				}
			},
			channelID:        "C12345",
			options:          []slack.MsgOption{slack.MsgOptionText("Hello", false)},
			expectError:      true,
			expectErrMessage: "failed to post message to Slack channel C12345: slack-api-error",
		},
		{
			name: "Uninitialized slackGoClient (internally nil)",
			setupClient: func() *SlackClient {
				// Create client but explicitly set internal client to nil
				// This simulates a state that might occur if NewSlackClient logic changes or is bypassed
				client := &SlackClient{auth: validAuth} // slackGoClient will be nil
				return client
			},
			mockSetup:        func(m *mockSlackGoClient) { /* No mock needed, direct path */ },
			channelID:        "C12345",
			options:          []slack.MsgOption{slack.MsgOptionText("Hello", false)},
			expectError:      true,
			expectErrMessage: "SlackClient not initialized or slackGoClient is nil",
		},
		{
			name:             "Empty channelID",
			setupClient:      func() *SlackClient { client, _ := NewSlackClient(validAuth); return client },
			mockSetup:        func(m *mockSlackGoClient) { /* No mock needed for validation path */ },
			channelID:        "",
			options:          []slack.MsgOption{slack.MsgOptionText("Hello", false)},
			expectError:      true,
			expectErrMessage: "channelID cannot be empty",
		},
		{
			name:             "No message options",
			setupClient:      func() *SlackClient { client, _ := NewSlackClient(validAuth); return client },
			mockSetup:        func(m *mockSlackGoClient) { /* No mock needed for validation path */ },
			channelID:        "C12345",
			options:          []slack.MsgOption{},
			expectError:      true,
			expectErrMessage: "at least one MsgOption (e.g., text or blocks) must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()

			// If the client has a mockable slackGoClient, set it up
			if mockable, ok := client.slackGoClient.(*mockSlackGoClient); ok {
				tt.mockSetup(mockable)
			} else if tt.name == "Successful PostMessage" || tt.name == "Slack API returns error" {
				// For tests that are SUPPOSED to use the mock, if it's not a mock, something is wrong in test setup
				t.Fatal("Expected client.slackGoClient to be a *mockSlackGoClient for this test case")
			}

			respChannelID, respTimestamp, err := client.PostMessage(tt.channelID, tt.options...)

			if tt.expectError {
				assert.Error(t, err)
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectErrMessage)
				assert.Empty(t, respChannelID)
				assert.Empty(t, respTimestamp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectChannelID, respChannelID)
				assert.Equal(t, tt.expectTimestamp, respTimestamp)
			}
		})
	}
}
