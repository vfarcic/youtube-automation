package slack

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	// Default to Info level, can be made configurable later if needed
	log.SetLevel(logrus.InfoLevel)
	log.SetOutput(os.Stdout)
}

// SetLogLevel allows changing the global log level for Slack operations.
// This could be called from a central configuration loading point in the application.
func SetLogLevel(level logrus.Level) {
	log.SetLevel(level)
}

func baseEntry() *logrus.Entry {
	return log.WithField("component", "slack")
}

// LogSlackError logs a categorized Slack error with structured fields.
func LogSlackError(sErr *SlackError, message string) {
	if sErr == nil {
		baseEntry().Error(message)
		return
	}

	fields := logrus.Fields{
		"error_type": sErr.Type,
		"retryable":  sErr.Retryable,
	}
	// Do not add sErr.OriginalError to fields directly if using .WithError()

	entry := baseEntry().WithFields(fields)

	if sErr.OriginalError != nil {
		entry.WithError(sErr.OriginalError).Error(fmt.Sprintf("%s: %s", message, sErr.Message))
	} else {
		entry.Error(fmt.Sprintf("%s: %s", message, sErr.Message))
	}
}

// LogSlackWarn logs a warning message related to Slack operations.
func LogSlackWarn(message string, args ...interface{}) {
	baseEntry().Warnf(message, args...)
}

// LogSlackInfo logs an informational message related to Slack operations.
func LogSlackInfo(message string, args ...interface{}) {
	baseEntry().Infof(message, args...)
}

// LogSlackDebug logs a debug message related to Slack operations.
// Note: Debug logs will only be output if log level is set to DebugLevel.
func LogSlackDebug(message string, args ...interface{}) {
	baseEntry().Debugf(message, args...)
}
