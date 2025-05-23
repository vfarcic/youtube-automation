package slack

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestSlackError_Error(t *testing.T) {
	originalErr := errors.New("original issue")
	tests := []struct {
		name     string
		sErr     *SlackError
		expected string
	}{
		{
			name: "Error with original error",
			sErr: &SlackError{
				Type:          ErrorTypeNetwork,
				Message:       "Connection failed",
				OriginalError: originalErr,
			},
			expected: "Slack error [network]: Connection failed (original: original issue)",
		},
		{
			name: "Error without original error",
			sErr: &SlackError{
				Type:    ErrorTypeAuth,
				Message: "Invalid token",
			},
			expected: "Slack error [auth]: Invalid token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.sErr.Error())
		})
	}
}

func TestSlackError_Unwrap(t *testing.T) {
	originalErr := errors.New("root cause")
	wrappedErr := &SlackError{
		Type:          ErrorTypeUnknown,
		Message:       "Something went wrong",
		OriginalError: originalErr,
	}

	assert.Equal(t, originalErr, errors.Unwrap(wrappedErr))
	assert.Nil(t, errors.Unwrap(&SlackError{Message: "No original"})) // Test case where OriginalError is nil
}

func TestCategorizeError(t *testing.T) {
	tests := []struct {
		name               string
		inputError         error
		expectedSlackError *SlackError // Only Type, Retryable, and presence of OriginalError are checked for simplicity
	}{
		{
			name:               "Nil error",
			inputError:         nil,
			expectedSlackError: nil,
		},
		{
			name:       "Slack RateLimitedError",
			inputError: &slack.RateLimitedError{RetryAfter: 30 * time.Second},
			expectedSlackError: &SlackError{
				Type:      ErrorTypeRateLimit,
				Retryable: true, // slack.RateLimitedError is always retryable
			},
		},
		{
			name:               "Auth error - invalid_auth",
			inputError:         errors.New("slack api error: invalid_auth"),
			expectedSlackError: &SlackError{Type: ErrorTypeAuth, Retryable: false},
		},
		{
			name:               "Auth error - not_authed",
			inputError:         errors.New("not_authed by slack rules"),
			expectedSlackError: &SlackError{Type: ErrorTypeAuth, Retryable: false},
		},
		{
			name:               "Auth error - account_inactive",
			inputError:         errors.New("user account_inactive"),
			expectedSlackError: &SlackError{Type: ErrorTypeAuth, Retryable: false},
		},
		{
			name:               "Auth error - token_revoked",
			inputError:         errors.New("bad token_revoked"),
			expectedSlackError: &SlackError{Type: ErrorTypeAuth, Retryable: false},
		},
		{
			name:               "Auth error - missing_scope",
			inputError:         errors.New("missing_scope for this action"),
			expectedSlackError: &SlackError{Type: ErrorTypeAuth, Retryable: false},
		},
		{
			name:               "Channel error - channel_not_found",
			inputError:         errors.New("oops, channel_not_found there"),
			expectedSlackError: &SlackError{Type: ErrorTypeChannel, Retryable: false},
		},
		{
			name:               "Channel error - is_archived",
			inputError:         errors.New("channel is_archived, cannot post"),
			expectedSlackError: &SlackError{Type: ErrorTypeChannel, Retryable: false},
		},
		{
			name:               "Network error - connection refused",
			inputError:         errors.New("dial tcp: connection refused by server"),
			expectedSlackError: &SlackError{Type: ErrorTypeNetwork, Retryable: true},
		},
		{
			name:               "Network error - connection reset by peer",
			inputError:         errors.New("oops connection reset by peer failure"),
			expectedSlackError: &SlackError{Type: ErrorTypeNetwork, Retryable: true},
		},
		{
			name:               "Network error - timeout",
			inputError:         errors.New("client transport: i/o timeout"),
			expectedSlackError: &SlackError{Type: ErrorTypeNetwork, Retryable: true},
		},
		{
			name:               "Network error - no such host",
			inputError:         errors.New("lookup api.slack.com: no such host"),
			expectedSlackError: &SlackError{Type: ErrorTypeNetwork, Retryable: true},
		},
		{
			name:               "Network error - network is unreachable",
			inputError:         errors.New("dial: network is unreachable for slack"),
			expectedSlackError: &SlackError{Type: ErrorTypeNetwork, Retryable: true},
		},
		{
			name:               "Invalid request error - invalid_payload",
			inputError:         errors.New("request contained invalid_payload according to slack"),
			expectedSlackError: &SlackError{Type: ErrorTypeInvalid, Retryable: false},
		},
		{
			name:               "Invalid request error - invalid_arg_name",
			inputError:         errors.New("invalid_arg_name for your request"),
			expectedSlackError: &SlackError{Type: ErrorTypeInvalid, Retryable: false},
		},
		{
			name:               "Invalid request error - bad_request",
			inputError:         errors.New("slack says bad_request to you"),
			expectedSlackError: &SlackError{Type: ErrorTypeInvalid, Retryable: false},
		},
		{
			name:               "Server error - fatal_error",
			inputError:         errors.New("slack had an fatal_error, oh no"),
			expectedSlackError: &SlackError{Type: ErrorTypeServer, Retryable: true},
		},
		{
			name:               "Server error - internal_error",
			inputError:         errors.New("an internal_error on slack side"),
			expectedSlackError: &SlackError{Type: ErrorTypeServer, Retryable: true},
		},
		{
			name:               "Unknown error",
			inputError:         errors.New("a very mysterious problem occurred"),
			expectedSlackError: &SlackError{Type: ErrorTypeUnknown, Retryable: false},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := CategorizeError(tc.inputError)
			if tc.expectedSlackError == nil {
				assert.Nil(t, actual)
				return
			}
			assert.NotNil(t, actual)
			assert.Equal(t, tc.expectedSlackError.Type, actual.Type, "ErrorType mismatch")
			assert.Equal(t, tc.expectedSlackError.Retryable, actual.Retryable, "Retryable mismatch")
			if tc.inputError != nil { // If there was an input error, it should be wrapped
				assert.Equal(t, tc.inputError, actual.OriginalError, "OriginalError mismatch")
				// Check that the message is not empty and gives some context
				assert.NotEmpty(t, actual.Message, "Message should not be empty for non-nil categorized error")
				if rateLimitedErr, ok := tc.inputError.(*slack.RateLimitedError); ok {
					assert.Contains(t, actual.Message, fmt.Sprintf("%v", rateLimitedErr.RetryAfter), "RateLimitError message should contain RetryAfter duration")
				}
			} else if actual != nil { // Should only happen if expectedSlackError was nil but actual was not
				assert.Fail(t, "Actual error was not nil when expected was nil")
			}
		})
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name       string
		inputError error
		expected   bool
	}{
		{
			name:       "Nil error",
			inputError: nil,
			expected:   false,
		},
		{
			name:       "Slack RateLimitedError",
			inputError: &slack.RateLimitedError{RetryAfter: 5 * time.Second},
			expected:   true,
		},
		{
			name:       "Network error (categorized as retryable)",
			inputError: errors.New("some network timeout issue"), // This will be categorized
			expected:   true,
		},
		{
			name:       "Auth error (categorized as not retryable)",
			inputError: errors.New("invalid_auth token problem"), // This will be categorized
			expected:   false,
		},
		{
			name:       "Unknown error (categorized as not retryable)",
			inputError: errors.New("completely unexpected error"), // This will be categorized
			expected:   false,
		},
		{
			// Test case for when CategorizeError itself might return nil for a non-nil error (should not happen with current logic)
			name:       "Error that might lead to nil categorized error (defensive)",
			inputError: &customNonSlackError{}, // Needs a custom error type that won't be categorized
			expected:   false,                  // Default to not retryable if categorization is odd
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsRetryableError(tc.inputError))
		})
	}
}

// customNonSlackError is a helper for testing IsRetryableError robustness.
type customNonSlackError struct{}

func (e *customNonSlackError) Error() string { return "custom non-slacky error" }
