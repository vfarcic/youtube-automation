package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// mockLLM for description_tags_test.go - ensuring it's compatible with other tests in package ai
// It seems the other tests (e.g. highlights_test.go) define this. If running tests for the whole package,
// this might lead to a redefinition error if not handled carefully (e.g. build tags or one shared mock file).
// For isolated testing of this file, it would be needed.
// Let's assume for now it's defined in highlights_test.go or similar and accessible.

func TestSuggestDescriptionTags(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video description tag generation. It needs three specific #tags separated by spaces."
	validAIConfig := AITitleGeneratorConfig{
		Endpoint:       "https://fake-desc-tags-endpoint.openai.azure.com/",
		DeploymentName: "fake-desc-tags-deployment",
		APIKey:         "fake-desc-tags-api-key",
		APIVersion:     "2023-07-01-preview",
	}

	originalNewLLMDescTags := newLLMClientFuncForDescriptionTags
	defer func() { newLLMClientFuncForDescriptionTags = originalNewLLMDescTags }()

	tests := []struct {
		name              string
		mockLLMResponse   string
		mockLLMError      error
		aiConfig          AITitleGeneratorConfig
		manuscript        string
		wantTags          string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:            "Successful description tag generation",
			mockLLMResponse: "#go #programming #ai",
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTags:        "#go #programming #ai",
			wantErr:         false,
		},
		{
			name:            "AI response with leading/trailing spaces",
			mockLLMResponse: "  #go #programming #ai  ",
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTags:        "#go #programming #ai",
			wantErr:         false,
		},
		{
			name:              "AI returns incorrect number of tags (too few hashtags)",
			mockLLMResponse:   "#go #programming", // Missing one #
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI did not return exactly three space-separated tags starting with '#'",
		},
		{
			name:              "AI returns incorrect number of tags (too many hashtags)",
			mockLLMResponse:   "#go #programming #ai #extra",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI did not return exactly three space-separated tags starting with '#'",
		},
		{
			name:              "AI returns incorrect format (wrong separator)",
			mockLLMResponse:   "#go,#programming,#ai", // Commas instead of spaces
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI did not return exactly three space-separated tags starting with '#'",
		},
		{
			name:              "AI returns tags not starting with #",
			mockLLMResponse:   "go programming ai",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI did not return exactly three space-separated tags starting with '#'",
		},
		{
			name:              "AI returns empty response for description tags",
			mockLLMResponse:   "",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for description tags",
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
			mockLLMError:      fmt.Errorf("mock LLM generation error for desc tags"),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "mock LLM generation error for desc tags",
		},
		{
			name:              "LLM creation fails",
			mockLLMError:      fmt.Errorf("mock LLM creation error for desc tags"),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "failed to create LangChainGo client for description tags",
		},
		{
			name:              "Incomplete AI config - no API key for description tags",
			aiConfig:          AITitleGeneratorConfig{Endpoint: "e", DeploymentName: "d", APIVersion: "v"},
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set for description tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLLM{
				ResponseContent: tt.mockLLMResponse,
				ErrorToReturn:   tt.mockLLMError,
			}

			newLLMClientFuncForDescriptionTags = func(options ...openai.Option) (llms.Model, error) {
				if strings.Contains(tt.name, "LLM creation fails") {
					return nil, tt.mockLLMError
				}
				return mock, nil
			}

			gotTags, err := SuggestDescriptionTags(ctx, tt.manuscript, tt.aiConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestDescriptionTags() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.mockLLMError != nil && (strings.Contains(tt.name, "LLM generation fails") || strings.Contains(tt.name, "LLM creation fails")) {
					if !strings.Contains(err.Error(), tt.mockLLMError.Error()) {
						t.Errorf("SuggestDescriptionTags() error = %q, want to contain underlying LLM error %q", err.Error(), tt.mockLLMError.Error())
					}
				} else if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("SuggestDescriptionTags() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("SuggestDescriptionTags() unexpected error = %v", err)
				}
				if gotTags != tt.wantTags {
					t.Errorf("SuggestDescriptionTags() gotTags = %q, want %q", gotTags, tt.wantTags)
				}
			}
		})
	}
}
