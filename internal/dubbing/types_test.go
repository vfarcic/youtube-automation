package dubbing

import (
	"encoding/json"
	"testing"
)

func TestDubbingJob_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		wantID   string
		wantStatus string
		wantError string
	}{
		{
			name:     "complete job",
			json:     `{"dubbing_id":"dub_123","status":"dubbed","target_languages":["es"],"expected_duration_sec":120.5}`,
			wantID:   "dub_123",
			wantStatus: "dubbed",
		},
		{
			name:     "failed job with error",
			json:     `{"dubbing_id":"dub_456","status":"failed","error":"Processing failed"}`,
			wantID:   "dub_456",
			wantStatus: "failed",
			wantError: "Processing failed",
		},
		{
			name:     "in progress job",
			json:     `{"dubbing_id":"dub_789","status":"dubbing"}`,
			wantID:   "dub_789",
			wantStatus: "dubbing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var job DubbingJob
			if err := json.Unmarshal([]byte(tt.json), &job); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if job.DubbingID != tt.wantID {
				t.Errorf("expected DubbingID %q, got %q", tt.wantID, job.DubbingID)
			}
			if job.Status != tt.wantStatus {
				t.Errorf("expected Status %q, got %q", tt.wantStatus, job.Status)
			}
			if job.Error != tt.wantError {
				t.Errorf("expected Error %q, got %q", tt.wantError, job.Error)
			}
		})
	}
}

func TestConfig_Defaults(t *testing.T) {
	config := Config{}

	if config.TestMode != false {
		t.Error("expected TestMode to default to false")
	}
	if config.NumSpeakers != 0 {
		t.Error("expected NumSpeakers to default to 0")
	}
	if config.DropBackgroundAudio != false {
		t.Error("expected DropBackgroundAudio to default to false")
	}
	if config.StartTime != 0 {
		t.Error("expected StartTime to default to 0")
	}
	if config.EndTime != 0 {
		t.Error("expected EndTime to default to 0")
	}
}

func TestCreateDubbingResponse_JSONUnmarshal(t *testing.T) {
	jsonStr := `{"dubbing_id":"dub_test","expected_duration_sec":45.5}`

	var resp createDubbingResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.DubbingID != "dub_test" {
		t.Errorf("expected DubbingID 'dub_test', got %q", resp.DubbingID)
	}
	if resp.ExpectedDuration != 45.5 {
		t.Errorf("expected ExpectedDuration 45.5, got %f", resp.ExpectedDuration)
	}
}

func TestErrorResponse_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantMessage string
	}{
		{
			name:        "detail object with status and message",
			json:        `{"detail":{"status":"error","message":"Something went wrong"}}`,
			wantMessage: "error: Something went wrong",
		},
		{
			name:        "detail object with message only",
			json:        `{"detail":{"message":"Video not found"}}`,
			wantMessage: "Video not found",
		},
		{
			name:        "detail as string",
			json:        `{"detail":"Access denied"}`,
			wantMessage: "Access denied",
		},
		{
			name:        "error field",
			json:        `{"error":"Invalid URL format"}`,
			wantMessage: "Invalid URL format",
		},
		{
			name:        "message field",
			json:        `{"message":"Rate limit exceeded"}`,
			wantMessage: "Rate limit exceeded",
		},
		{
			name:        "description field",
			json:        `{"description":"Service unavailable"}`,
			wantMessage: "Service unavailable",
		},
		{
			name:        "status_code with message",
			json:        `{"status_code":422,"message":"Unprocessable entity"}`,
			wantMessage: "Unprocessable entity",
		},
		{
			name:        "empty response",
			json:        `{}`,
			wantMessage: "",
		},
		{
			name:        "priority: detail.message over error",
			json:        `{"detail":{"message":"Primary error"},"error":"Secondary error"}`,
			wantMessage: "Primary error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp errorResponse
			if err := json.Unmarshal([]byte(tt.json), &resp); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			got := resp.GetMessage()
			if got != tt.wantMessage {
				t.Errorf("GetMessage() = %q, want %q", got, tt.wantMessage)
			}
		})
	}
}

func TestErrorDetail_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		wantStatus string
		wantMsg    string
		wantRaw    string
	}{
		{
			name:       "object format",
			json:       `{"status":"forbidden","message":"Access denied"}`,
			wantStatus: "forbidden",
			wantMsg:    "Access denied",
			wantRaw:    "",
		},
		{
			name:       "string format",
			json:       `"Simple error message"`,
			wantStatus: "",
			wantMsg:    "",
			wantRaw:    "Simple error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var detail errorDetail
			if err := json.Unmarshal([]byte(tt.json), &detail); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if detail.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", detail.Status, tt.wantStatus)
			}
			if detail.Message != tt.wantMsg {
				t.Errorf("Message = %q, want %q", detail.Message, tt.wantMsg)
			}
			if detail.Raw != tt.wantRaw {
				t.Errorf("Raw = %q, want %q", detail.Raw, tt.wantRaw)
			}
		})
	}
}
