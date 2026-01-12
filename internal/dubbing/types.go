package dubbing

import "encoding/json"

// Config holds ElevenLabs configuration
type Config struct {
	APIKey              string
	TestMode            bool // true = watermark + lower resolution (saves credits)
	StartTime           int  // Start time in seconds (0 = beginning)
	EndTime             int  // End time in seconds (0 = full video)
	NumSpeakers         int  // Number of speakers (default: 1)
	DropBackgroundAudio bool // Whether to drop background audio (default: false)
}

// DubbingJob represents a dubbing job response from ElevenLabs
type DubbingJob struct {
	DubbingID        string   `json:"dubbing_id"`
	Name             string   `json:"name,omitempty"`
	Status           string   `json:"status"` // "dubbing", "dubbed", "failed"
	TargetLanguages  []string `json:"target_languages,omitempty"`
	Error            string   `json:"error,omitempty"`
	ExpectedDuration float64  `json:"expected_duration_sec,omitempty"`
}

// DubbingStatus constants
const (
	StatusDubbing = "dubbing" // Job is in progress
	StatusDubbed  = "dubbed"  // Job completed successfully
	StatusFailed  = "failed"  // Job failed
)

// createDubbingResponse represents the response from creating a dubbing job
type createDubbingResponse struct {
	DubbingID        string `json:"dubbing_id"`
	ExpectedDuration float64 `json:"expected_duration_sec,omitempty"`
}

// errorResponse represents an error response from the ElevenLabs API
// ElevenLabs can return errors in different formats:
// - {"detail": {"status": "...", "message": "..."}}
// - {"detail": "string message"}
// - {"error": "...", "message": "..."}
type errorResponse struct {
	Detail      errorDetail `json:"detail,omitempty"`
	Error       string      `json:"error,omitempty"`
	Message     string      `json:"message,omitempty"`
	StatusCode  int         `json:"status_code,omitempty"`
	Description string      `json:"description,omitempty"`
}

// errorDetail handles both object and string formats for the detail field
type errorDetail struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Raw     string // For when detail is a plain string
}

// UnmarshalJSON handles both {"detail": "string"} and {"detail": {...}} formats
func (e *errorDetail) UnmarshalJSON(data []byte) error {
	// Try as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.Raw = s
		return nil
	}

	// Try as object
	type detailObj struct {
		Status  string `json:"status,omitempty"`
		Message string `json:"message,omitempty"`
	}
	var obj detailObj
	if err := json.Unmarshal(data, &obj); err == nil {
		e.Status = obj.Status
		e.Message = obj.Message
		return nil
	}

	return nil // Ignore parse errors, fall back to raw response
}

// GetMessage returns the best available error message from the response
func (e *errorResponse) GetMessage() string {
	// Priority: detail.message > detail.raw > message > error > description
	if e.Detail.Message != "" {
		if e.Detail.Status != "" {
			return e.Detail.Status + ": " + e.Detail.Message
		}
		return e.Detail.Message
	}
	if e.Detail.Raw != "" {
		return e.Detail.Raw
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Error != "" {
		return e.Error
	}
	if e.Description != "" {
		return e.Description
	}
	return ""
}
