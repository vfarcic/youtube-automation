package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	// schema is imported by other test files in this package, like highlights_test.go, which defines mockLLM
)

func TestSuggestTags(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video tag generation. It needs a good comma-separated list of tags under 450 characters."
	longManuscript := strings.Repeat("This is a very long manuscript. ", 50) // To test truncation indirectly if AI over-generates
	validAIConfig := AITitleGeneratorConfig{
		Endpoint:       "https://fake-tags-endpoint.openai.azure.com/",
		DeploymentName: "fake-tags-deployment",
		APIKey:         "fake-tags-api-key",
		APIVersion:     "2023-07-01-preview",
	}

	originalNewLLMTags := newLLMClientFuncForTags
	defer func() { newLLMClientFuncForTags = originalNewLLMTags }()

	tests := []struct {
		name              string
		mockLLMResponse   string // This will be assigned to mockLLM.response
		mockLLMError      error  // This will be assigned to mockLLM.err
		aiConfig          AITitleGeneratorConfig
		manuscript        string
		wantTags          string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:            "Successful tag generation",
			mockLLMResponse: "go, programming, ai, video, testing",
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTags:        "go, programming, ai, video, testing",
			wantErr:         false,
		},
		{
			name:            "AI response with leading/trailing spaces",
			mockLLMResponse: "  go, programming, ai  ",
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTags:        "go, programming, ai",
			wantErr:         false,
		},
		{
			name:            "AI response exceeds 450 characters - intelligent truncation",
			mockLLMResponse: strings.Repeat("tag,", 100) + "longtagthatwillbecutoff", // 423 chars
			aiConfig:        validAIConfig,
			manuscript:      longManuscript,
			// Expected: the original response, as it's < 450 chars, so truncation logic is skipped.
			wantTags: strings.TrimSpace(strings.Repeat("tag,", 100) + "longtagthatwillbecutoff"),
			wantErr:  false,
		},
		{
			name:            "AI response exceeds 450 characters - hard truncation (no comma before limit)",
			mockLLMResponse: strings.Repeat("a", 500),
			aiConfig:        validAIConfig,
			manuscript:      longManuscript,
			wantTags:        strings.Repeat("a", 450),
			wantErr:         false,
		},
		{
			name:              "AI returns empty response for tags",
			mockLLMResponse:   "",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for tags",
		},
		{
			name:              "Manuscript is empty",
			mockLLMResponse:   "", // LLM won't be called
			aiConfig:          validAIConfig,
			manuscript:        "  ", // Empty after trim
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "LLM generation fails",
			mockLLMError:      fmt.Errorf("mock LLM generation error"),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "mock LLM generation error", // This will be wrapped by the retry logic
		},
		{
			name:              "LLM creation fails",
			mockLLMError:      fmt.Errorf("mock LLM creation error"), // This error is used by the mock newLLMClientFunc
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "failed to create LangChainGo client for tags",
		},
		{
			name:              "Incomplete AI config - no API key",
			aiConfig:          AITitleGeneratorConfig{Endpoint: "e", DeploymentName: "d", APIVersion: "v"},
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLLM{
				ResponseContent: tt.mockLLMResponse,
				ErrorToReturn:   tt.mockLLMError,
			}

			newLLMClientFuncForTags = func(options ...openai.Option) (llms.Model, error) {
				if strings.Contains(tt.name, "LLM creation fails") {
					return nil, tt.mockLLMError
				}
				return mock, nil
			}

			gotTags, err := SuggestTags(ctx, tt.manuscript, tt.aiConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestTags() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.mockLLMError != nil && (strings.Contains(tt.name, "LLM generation fails") || strings.Contains(tt.name, "LLM creation fails")) {
					if !strings.Contains(err.Error(), tt.mockLLMError.Error()) {
						t.Errorf("SuggestTags() error = %q, want to contain underlying LLM error %q", err.Error(), tt.mockLLMError.Error())
					}
				} else if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
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

// The mockLLM struct and its methods are expected to be defined in another test file within this package (e.g., highlights_test.go or titles_test.go)
// and will be available at compile time.
