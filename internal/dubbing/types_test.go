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
	jsonStr := `{"detail":{"status":"error","message":"Something went wrong"}}`

	var resp errorResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Detail.Status != "error" {
		t.Errorf("expected status 'error', got %q", resp.Detail.Status)
	}
	if resp.Detail.Message != "Something went wrong" {
		t.Errorf("expected message 'Something went wrong', got %q", resp.Detail.Message)
	}
}
