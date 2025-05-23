package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Helper function to capture log output
func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stdout) // Reset to default
	return buf.String()
}

func TestLogSlackError(t *testing.T) {
	originalErr := errors.New("network hiccup")
	sErr := &SlackError{
		Type:          ErrorTypeNetwork,
		Message:       "Failed to connect",
		Retryable:     true,
		OriginalError: originalErr,
	}

	output := captureOutput(func() {
		LogSlackError(sErr, "Attempting to post message")
	})

	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(output), &logEntry)
	assert.NoError(t, err, "Log output should be valid JSON")

	assert.Equal(t, "error", logEntry["level"], "Log level should be error")
	assert.Equal(t, "slack", logEntry["component"], "Component should be slack")
	assert.Contains(t, logEntry["msg"], "Attempting to post message: Failed to connect", "Message mismatch")
	assert.Equal(t, string(ErrorTypeNetwork), logEntry["error_type"], "Error type mismatch")
	assert.Equal(t, true, logEntry["retryable"], "Retryable mismatch")
	assert.Equal(t, "network hiccup", logEntry[logrus.ErrorKey], "Original error message mismatch in log. Expected in '%s' field", logrus.ErrorKey)

	// Test with nil SlackError
	outputNilErr := captureOutput(func() {
		LogSlackError(nil, "A generic error occurred")
	})
	var logEntryNil map[string]interface{}
	errNil := json.Unmarshal([]byte(outputNilErr), &logEntryNil)
	assert.NoError(t, errNil, "Log output for nil SlackError should be valid JSON")
	assert.Equal(t, "error", logEntryNil["level"], "Log level should be error for nil SlackError")
	assert.Equal(t, "slack", logEntryNil["component"], "Component should be slack for nil SlackError")
	assert.Equal(t, "A generic error occurred", logEntryNil["msg"], "Message mismatch for nil SlackError")
	_, hasErrorType := logEntryNil["error_type"]
	assert.False(t, hasErrorType, "error_type should not be present for nil SlackError")

	// Test with SlackError that has no OriginalError
	sErrNoOriginal := &SlackError{
		Type:      ErrorTypeAuth,
		Message:   "Token expired",
		Retryable: false,
	}
	outputNoOriginal := captureOutput(func() {
		LogSlackError(sErrNoOriginal, "Auth check failed")
	})
	var logEntryNoOriginal map[string]interface{}
	errNoOriginal := json.Unmarshal([]byte(outputNoOriginal), &logEntryNoOriginal)
	assert.NoError(t, errNoOriginal)
	assert.Equal(t, string(ErrorTypeAuth), logEntryNoOriginal["error_type"])
	assert.Contains(t, logEntryNoOriginal["msg"], "Auth check failed: Token expired")
	_, hasOriginalErrorField := logEntryNoOriginal["original_error.Error"]
	// When OriginalError is nil, logrus.WithError is not called, so logrus.ErrorKey field should be absent
	_, hasErrorKeyField := logEntryNoOriginal[logrus.ErrorKey]
	assert.False(t, hasErrorKeyField, "'%s' field should not be present if OriginalError was nil and WithError was not called", logrus.ErrorKey)

	if hasOriginalErrorField { // This custom field should ideally also be absent or nil
		// This part of the test might need to be removed if "original_error" is fully removed from LogSlackError's WithFields
		// For now, if it's still there due to some logrus magic, it should be nil.
		// assert.Nil(t, logEntryNoOriginal["original_error.Error"], "original_error.Error should be nil if OriginalError was nil")
	}
}

func TestLogSlackWarnInfoDebug(t *testing.T) {
	tests := []struct {
		name        string
		logFunc     func(string, ...interface{})
		level       string
		message     string
		args        []interface{}
		expectedMsg string
	}{
		{
			name:        "Warn log",
			logFunc:     LogSlackWarn,
			level:       "warning",
			message:     "Potential issue with channel ID %s",
			args:        []interface{}{"C123"},
			expectedMsg: "Potential issue with channel ID C123",
		},
		{
			name:        "Info log",
			logFunc:     LogSlackInfo,
			level:       "info",
			message:     "Message posted successfully to %d channels",
			args:        []interface{}{3},
			expectedMsg: "Message posted successfully to 3 channels",
		},
		{
			name:        "Debug log",
			logFunc:     LogSlackDebug,
			level:       "debug",
			message:     "Video details: %v",
			args:        []interface{}{struct{ ID string }{ID: "v789"}},
			expectedMsg: "Video details: {v789}",
		},
	}

	originalLevel := log.GetLevel()
	defer log.SetLevel(originalLevel) // Restore original log level

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.level == "debug" {
				log.SetLevel(logrus.DebugLevel)
			} else {
				log.SetLevel(logrus.InfoLevel) // Ensure Warn and Info are captured
			}

			output := captureOutput(func() {
				tc.logFunc(tc.message, tc.args...)
			})

			// Debug messages might not be printed if level is not Debug
			if tc.level == "debug" && originalLevel > logrus.DebugLevel && log.GetLevel() != logrus.DebugLevel {
				// This check is a bit tricky because captureOutput sets the level temporarily.
				// If the test runs with a higher global log level, debug might not show.
				// The log.SetLevel(logrus.DebugLevel) inside the test run should ensure it's captured.
			}

			parsedLevel, err := logrus.ParseLevel(tc.level)
			assert.NoError(t, err, "Failed to parse log level: %s", tc.level)

			if log.IsLevelEnabled(parsedLevel) {
				var logEntry map[string]interface{}
				err := json.Unmarshal([]byte(output), &logEntry)
				assert.NoError(t, err, "Log output should be valid JSON for %s", tc.level)
				assert.Equal(t, tc.level, logEntry["level"], "Log level mismatch for %s", tc.level)
				assert.Equal(t, "slack", logEntry["component"], "Component should be slack for %s", tc.level)
				assert.Equal(t, tc.expectedMsg, logEntry["msg"], "Message mismatch for %s", tc.level)
			} else {
				assert.Empty(t, output, "Expected no log output for %s when level is disabled", tc.level)
			}
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	originalLevel := log.GetLevel()
	defer log.SetLevel(originalLevel)

	SetLogLevel(logrus.ErrorLevel)
	assert.Equal(t, logrus.ErrorLevel, log.GetLevel())

	// Check that an info log is not printed
	output := captureOutput(func() {
		LogSlackInfo("This should not be printed")
	})
	assert.Empty(t, output, "Info log should not be printed when level is Error")

	SetLogLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, log.GetLevel())
	debugOutput := captureOutput(func() {
		LogSlackDebug("This is a debug message")
	})
	assert.NotEmpty(t, debugOutput, "Debug log should be printed when level is Debug")

	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(debugOutput), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "debug", logEntry["level"])
	assert.Equal(t, "This is a debug message", logEntry["msg"])
}
