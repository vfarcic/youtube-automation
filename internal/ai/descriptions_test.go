package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestSuggestDescription(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video description generation."
	validAIConfig := AITitleGeneratorConfig{ // Reusing AITitleGeneratorConfig as it has the same fields
		Endpoint:       "https://fake-desc-endpoint.openai.azure.com/",
		DeploymentName: "fake-desc-deployment",
		APIKey:         "fake-desc-api-key",
		APIVersion:     "2023-07-01-preview",
	}

	originalNewLLMDesc := newLLMClientFuncForDescriptions
	defer func() { newLLMClientFuncForDescriptions = originalNewLLMDesc }()

	tests := []struct {
		name              string
		mockLLMResponse   string
		mockLLMError      error
		aiConfig          AITitleGeneratorConfig
		manuscript        string
		wantDescription   string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:            "Successful description suggestion",
			mockLLMResponse: "This is a great AI-generated description.",
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantDescription: "This is a great AI-generated description.",
			wantErr:         false,
		},
		{
			name:              "AI returns empty response for description",
			mockLLMResponse:   "",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty description",
		},
		{
			name:              "AI generation error for description",
			mockLLMError:      fmt.Errorf("AI description service unavailable"),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "AI description service unavailable",
		},
		{
			name:              "Incomplete AI config for description - no endpoint",
			aiConfig:          AITitleGeneratorConfig{DeploymentName: "d", APIKey: "k", APIVersion: "v"},
			manuscript:        validManuscript,
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set",
		},
		{
			name:              "LLM creation fails for description",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			mockLLMError:      fmt.Errorf("mock llm creation failed for description"),
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "mock llm creation failed for description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLLM{ // mockLLM is from highlights_test.go (same package)
				ResponseContent: tt.mockLLMResponse,
				ErrorToReturn:   tt.mockLLMError,
			}

			newLLMClientFuncForDescriptions = func(options ...openai.Option) (llms.Model, error) {
				if strings.Contains(tt.name, "LLM creation fails") {
					return nil, tt.mockLLMError
				}
				return mock, nil
			}

			gotDescription, err := SuggestDescription(ctx, tt.manuscript, tt.aiConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestDescription() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				// Check if the underlying error is what we expect, even if wrapped by the retry logic
				if tt.mockLLMError != nil { // If we are testing an LLM error scenario
					if !strings.Contains(err.Error(), tt.mockLLMError.Error()) {
						t.Errorf("SuggestDescription() error = %q, want to contain underlying error %q", err.Error(), tt.mockLLMError.Error())
					}
				} else if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					// For non-LLM errors (e.g., config errors, empty response after successful LLM call)
					t.Errorf("SuggestDescription() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("SuggestDescription() unexpected error = %v", err)
					return
				}
			}

			if gotDescription != tt.wantDescription {
				t.Errorf("SuggestDescription() gotDescription = %v, want %v", gotDescription, tt.wantDescription)
			}

			// Check LLM call count for relevant tests
			if !tt.wantErr && tt.aiConfig.Endpoint != "" && tt.aiConfig.APIKey != "" && !strings.Contains(tt.name, "LLM creation fails") {
				if mock.CallCount == 0 {
					t.Errorf("Expected LLM to be called for SuggestDescription, but CallCount is 0")
				}
			}
		})
	}
}
