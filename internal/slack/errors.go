package slack

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

// ErrorType defines the category of a Slack-related error.
// This helps in deciding how to handle the error (e.g., retry, log level).
// Ensure all fields that need to be accessed from other packages are exported (start with a capital letter).
type ErrorType string

// Slack API Error Categories
const (
	ErrorTypeAuth      ErrorType = "auth"            // Authentication or permission issue
	ErrorTypeRateLimit ErrorType = "rate_limit"      // Rate limit exceeded
	ErrorTypeNetwork   ErrorType = "network"         // Network connectivity problem
	ErrorTypeInvalid   ErrorType = "invalid_request" // Malformed or invalid request (e.g., bad channel ID)
	ErrorTypeServer    ErrorType = "server_error"    // Slack server-side issue (5xx errors)
	ErrorTypeChannel   ErrorType = "channel_error"   // Issues specific to a channel (e.g., not found, archived)
	ErrorTypeUnknown   ErrorType = "unknown"         // Error that doesn't fit other categories
	ErrorTypeInternal  ErrorType = "internal"        // Errors originating from within this application
)

// SlackError is a custom error structure to wrap and categorize errors from Slack operations.
// Ensure all fields that need to be accessed from other packages are exported.
type SlackError struct {
	Type          ErrorType // Category of the error
	Message       string    // Human-readable error message
	Retryable     bool      // Indicates if the operation that caused this error can be retried
	OriginalError error     // The original error object, if any
}

// Error implements the error interface for SlackError.
func (e *SlackError) Error() string {
	if e.OriginalError != nil {
		return fmt.Sprintf("Slack error [%s]: %s (original: %v)", e.Type, e.Message, e.OriginalError)
	}
	return fmt.Sprintf("Slack error [%s]: %s", e.Type, e.Message)
}

// Unwrap provides compatibility for Go 1.13+ error chains.
func (e *SlackError) Unwrap() error {
	return e.OriginalError
}

// CategorizeError inspects an error and returns a structured SlackError.
// It attempts to identify specific error types from the slack-go library first,
// then falls back to string matching for common error messages.
func CategorizeError(err error) *SlackError {
	if err == nil {
		return nil
	}

	// Check for specific typed errors from slack-go library
	if slackRateLimitedError, ok := err.(*slack.RateLimitedError); ok {
		return &SlackError{
			Type:          ErrorTypeRateLimit,
			Message:       fmt.Sprintf("Rate limited by Slack. Retry after %v.", slackRateLimitedError.RetryAfter),
			Retryable:     slackRateLimitedError.Retryable(), // Use the method from the error itself
			OriginalError: err,
		}
	}
	// TODO: Add checks for other specific slack-go typed errors if they exist and are relevant (e.g., channel not found)
	// For example, slack.SlackErrorResponse might contain more details if the error is of that type.
	// if sre, ok := err.(slack.SlackErrorResponse); ok {
	//    // sre.Err contains the string error like "channel_not_found"
	// }

	// Fallback to string matching for common error patterns
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "invalid_auth"),
		strings.Contains(errStr, "not_authed"),
		strings.Contains(errStr, "account_inactive"),
		strings.Contains(errStr, "token_revoked"),
		strings.Contains(errStr, "missing_scope"):
		return &SlackError{
			Type:          ErrorTypeAuth,
			Message:       "Slack authentication or permission error.",
			Retryable:     false,
			OriginalError: err,
		}
	case strings.Contains(errStr, "channel_not_found"),
		strings.Contains(errStr, "is_archived"):
		return &SlackError{
			Type:          ErrorTypeChannel,
			Message:       "Slack channel error (e.g., not found, archived).",
			Retryable:     false, // Usually not retryable without intervention
			OriginalError: err,
		}
	case strings.Contains(errStr, "connection refused"), // Go's net errors
		strings.Contains(errStr, "connection reset by peer"),
		strings.Contains(errStr, "timeout"),      // Covers i/o timeout, client timeout
		strings.Contains(errStr, "no such host"), // DNS resolution issue
		strings.Contains(errStr, "network is unreachable"):
		return &SlackError{
			Type:          ErrorTypeNetwork,
			Message:       "Network error occurred while communicating with Slack.",
			Retryable:     true,
			OriginalError: err,
		}
	case strings.Contains(errStr, "invalid_payload"), // Common for bad requests
		strings.Contains(errStr, "invalid_arg_name"),
		strings.Contains(errStr, "bad_request"):
		return &SlackError{
			Type:          ErrorTypeInvalid,
			Message:       "Invalid request sent to Slack (e.g., malformed payload).",
			Retryable:     false,
			OriginalError: err,
		}
	case strings.Contains(errStr, "fatal_error"), // Generic Slack server errors
		strings.Contains(errStr, "internal_error"):
		return &SlackError{
			Type:          ErrorTypeServer,
			Message:       "Slack reported a server-side error.",
			Retryable:     true, // Often transient
			OriginalError: err,
		}
	}

	// Default if no other category matches
	return &SlackError{
		Type:          ErrorTypeUnknown,
		Message:       "An unknown error occurred with Slack.",
		Retryable:     false,
		OriginalError: err,
	}
}

// IsRetryableError checks if an error, after categorization, is marked as retryable.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// First, check if the error itself is a slack.RateLimitedError, as it has a Retryable method.
	if rateLimitedErr, ok := err.(*slack.RateLimitedError); ok {
		return rateLimitedErr.Retryable()
	}
	// Otherwise, categorize and check our custom SlackError's Retryable field.
	categorizedErr := CategorizeError(err)
	return categorizedErr != nil && categorizedErr.Retryable
}
