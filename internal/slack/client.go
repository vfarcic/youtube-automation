package slack

import (
	"errors" // For NewSlackClient error return
	"fmt"    // For error wrapping in PostMessage
	"net/http"

	"github.com/slack-go/slack"
	// "devopstoolkit/youtube-automation/internal/ratelimiter" // Placeholder for now
)

// slackGoClientInterface defines the methods we use from the slack-go client.
// This allows for easier mocking in tests.
type slackGoClientInterface interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	// Add other methods from slack.Client if SlackClient starts using them directly.
}

// SlackClient wraps the slack-go client and adds custom functionality
// like rate limiting and retries.
type SlackClient struct {
	auth          *SlackAuth
	slackGoClient slackGoClientInterface // Changed from *slack.Client
	// rateLimiter   *ratelimiter.RateLimiter // To be added in a later subtask
	httpClient *http.Client // Optional: if we need to customize the underlying client
}

// ClientOption defines a function signature for options to configure the SlackClient.
type ClientOption func(*SlackClient)

// WithHTTPClient allows providing a custom http.Client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *SlackClient) {
		c.httpClient = hc
	}
}

// newSlackClient_actual creates a new Slack API client.
// It requires a valid SlackAuth.
// Additional options can be provided to customize the client, e.g., a custom HTTP client.
func newSlackClient_actual(auth *SlackAuth, opts ...ClientOption) (*SlackClient, error) {
	if auth == nil || auth.Token == "" {
		return nil, errors.New("cannot create SlackClient: SlackAuth is nil or token is empty")
	}

	sc := &SlackClient{
		auth: auth,
	}

	for _, opt := range opts {
		opt(sc)
	}

	var slackGoClientOpts []slack.Option
	if sc.httpClient != nil {
		slackGoClientOpts = append(slackGoClientOpts, slack.OptionHTTPClient(sc.httpClient))
	}
	// We can add other options like slack.OptionDebug(true), slack.OptionAPIURL("customURL") if needed.

	sc.slackGoClient = slack.New(auth.GetToken(), slackGoClientOpts...)

	// Initialize RateLimiter in a later subtask
	// sc.rateLimiter = ratelimiter.New(...)

	return sc, nil
}

// NewSlackClient is a function variable for creating a new Slack API client.
// This allows it to be replaced for mocking in tests.
var NewSlackClient = newSlackClient_actual

// PostMessage sends a message to a Slack channel.
// It wraps the underlying slack-go client's PostMessage method.
// Parameters:
//   - channelID: The ID of the channel to post to (e.g., "C1234567890").
//   - options: A variadic list of slack.MsgOption to define the message content
//     (e.g., slack.MsgOptionText("Hello world", false), slack.MsgOptionBlocks(...)).
//
// Returns:
//   - string: The channel ID where the message was posted.
//   - string: The timestamp of the posted message.
//   - error: An error if the message could not be posted.
func (c *SlackClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	if c.slackGoClient == nil {
		return "", "", errors.New("SlackClient not initialized or slackGoClient is nil")
	}
	if channelID == "" {
		return "", "", errors.New("channelID cannot be empty")
	}
	if len(options) == 0 {
		return "", "", errors.New("at least one MsgOption (e.g., text or blocks) must be provided")
	}

	// For now, direct call. Rate limiting and retries will be added in later subtasks.
	respChannelID, respTimestamp, err := c.slackGoClient.PostMessage(channelID, options...)
	if err != nil {
		// Basic error wrapping. More sophisticated error handling can be added later.
		return "", "", fmt.Errorf("failed to post message to Slack channel %s: %w", channelID, err)
	}

	return respChannelID, respTimestamp, nil
}
