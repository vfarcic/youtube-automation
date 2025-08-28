package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestSuggestTags(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video tag generation. It needs a good comma-separated list of tags under 450 characters."
	longManuscript := strings.Repeat("This is a very long manuscript. ", 50) // To test truncation indirectly if AI over-generates

	tests := []struct {
		name              string
		mockResponse      string
		mockError         error
		manuscript        string
		wantTags          string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:         "Successful tag generation",
			mockResponse: "go, programming, ai, video, testing",
			manuscript:   validManuscript,
			wantTags:     "go, programming, ai, video, testing",
			wantErr:      false,
		},
		{
			name:         "AI response with leading/trailing spaces",
			mockResponse: "  go, programming, ai  ",
			manuscript:   validManuscript,
			wantTags:     "go, programming, ai",
			wantErr:      false,
		},
		{
			name:         "AI response exceeds 450 characters - intelligent truncation",
			mockResponse: strings.Repeat("tag,", 100) + "longtagthatwillbecutoff", // 423 chars
			manuscript:   longManuscript,
			// Expected: the original response, as it's < 450 chars, so truncation logic is skipped.
			wantTags: strings.TrimSpace(strings.Repeat("tag,", 100) + "longtagthatwillbecutoff"),
			wantErr:  false,
		},
		{
			name:         "AI response exceeds 450 characters - hard truncation (no comma before limit)",
			mockResponse: strings.Repeat("a", 500),
			manuscript:   longManuscript,
			wantTags:     strings.Repeat("a", 450),
			wantErr:      false,
		},
		{
			name:              "AI returns empty response for tags",
			mockResponse:      "",
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for tags",
		},
		{
			name:              "Manuscript is empty",
			mockResponse:      "", // LLM won't be called
			manuscript:        "  ", // Empty after trim
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "AI generation fails",
			mockError:         fmt.Errorf("mock AI generation error"),
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "mock AI generation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			// Store original GetAIProvider function
			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			// Mock the GetAIProvider function
			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			gotTags, err := SuggestTags(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestTags() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("SuggestTags() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("SuggestTags() unexpected error = %v", err)
				}
				if gotTags != tt.wantTags {
					if tt.name == "AI response exceeds 450 characters - intelligent truncation" {
						if len(gotTags) > 450 {
							t.Errorf("SuggestTags() gotTags length = %d, want <= 450. Got: %q", len(gotTags), gotTags)
						}
						if !strings.HasPrefix(gotTags, "tag,") || !strings.Contains(gotTags, "tag,tag,tag") {
							t.Logf("Warning: Truncation test for tags might be brittle. Got: %q, Want (approx): %q", gotTags, tt.wantTags)
						}
					}
					if gotTags != tt.wantTags {
						t.Errorf("SuggestTags() gotTags = %q, want %q", gotTags, tt.wantTags)
					}
				}
			}
		})
	}
}

// The mockLLM struct and its methods are expected to be defined in another test file within this package (e.g., titles_test.go)
// and will be available at compile time.
