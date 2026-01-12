package dubbing

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
type errorResponse struct {
	Detail struct {
		Status  string `json:"status,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"detail,omitempty"`
}
