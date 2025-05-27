package ai

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// mockLLM is a mock implementation of llms.Model for testing.
// It allows configuring the response or error to be returned.
type mockLLM struct {
	ResponseContent string
	ErrorToReturn   error
	CallCount       int
}

// GenerateContent implements the llms.Model interface.
// It's the method usually called by functions like llms.GenerateFromSinglePrompt.
func (m *mockLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	m.CallCount++
	if m.ErrorToReturn != nil {
		return nil, m.ErrorToReturn
	}
	// Assuming the first message prompt is what we care about for this mock
	// and the response is a single text choice.
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: m.ResponseContent,
			},
		},
	}, nil
}

// Call is another method that might be required by the llms.Model interface or used internally.
func (m *mockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	m.CallCount++
	if m.ErrorToReturn != nil {
		return "", m.ErrorToReturn
	}
	return m.ResponseContent, nil
}

// GetNumTokens is a dummy implementation for the interface.
func (m *mockLLM) GetNumTokens(text string) int {
	return len(strings.Fields(text))
}

func TestSuggestHighlights(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript with important keywords and phrases to highlight."
	validAIConfig := AITitleGeneratorConfig{
		Endpoint:       "https://fake-endpoint.openai.azure.com/",
		DeploymentName: "fake-deployment",
		APIKey:         "fake-api-key",
		APIVersion:     "2023-07-01-preview",
	}

	originalNewLLMHighlights := newLLMClientFuncForHighlights
	defer func() { newLLMClientFuncForHighlights = originalNewLLMHighlights }()

	tests := []struct {
		name              string
		mockLLMResponse   string
		mockLLMError      error
		aiConfig          AITitleGeneratorConfig
		manuscript        string
		wantHighlights    []string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:            "Successful suggestion - direct array",
			mockLLMResponse: `["important keywords", "phrases to highlight"]`,
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantHighlights:  []string{"important keywords", "phrases to highlight"},
			wantErr:         false,
		},
		{
			name:            "Successful suggestion - object-wrapped array",
			mockLLMResponse: `{"suggested_highlights": ["keywords", "phrases"]}`,
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantHighlights:  []string{"keywords", "phrases"},
			wantErr:         false,
		},
		{
			name:            "Successful suggestion - with JSON code fence",
			mockLLMResponse: "```json\n{\"suggested_highlights\": [\"fenced keyword\"]}\n```",
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantHighlights:  []string{"fenced keyword"},
			wantErr:         false,
		},
		{
			name:              "AI returns empty response",
			mockLLMResponse:   "",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response",
		},
		{
			name:              "AI returns malformed JSON",
			mockLLMResponse:   `{"suggested_highlights": ["unterminated string]}`,
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "failed to unmarshal highlights JSON",
		},
		{
			name:              "AI generation error",
			mockLLMError:      fmt.Errorf("AI service unavailable"),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "AI service unavailable",
		},
		{
			name:              "Incomplete AI config - no endpoint",
			aiConfig:          AITitleGeneratorConfig{DeploymentName: "d", APIKey: "k", APIVersion: "v"},
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set",
		},
		{
			name:              "Incomplete AI config - no deployment name",
			aiConfig:          AITitleGeneratorConfig{Endpoint: "e", APIKey: "k", APIVersion: "v"},
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set",
		},
		{
			name:              "Incomplete AI config - no API key",
			aiConfig:          AITitleGeneratorConfig{Endpoint: "e", DeploymentName: "d", APIVersion: "v"},
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set",
		},
		{
			name:              "Incomplete AI config - no API version",
			aiConfig:          AITitleGeneratorConfig{Endpoint: "e", DeploymentName: "d", APIKey: "k"},
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set",
		},
		{
			name:              "LLM creation fails (mocked by returning error from newLLMClientFuncForHighlights)",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			mockLLMError:      fmt.Errorf("mock llm creation failed"), // This error will be returned by our mocked newLLMClientFuncForHighlights
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "mock llm creation failed", // This should match the error from newLLMClientFuncForHighlights
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLLM{
				ResponseContent: tt.mockLLMResponse,
				ErrorToReturn:   tt.mockLLMError,
			}

			newLLMClientFuncForHighlights = func(options ...openai.Option) (llms.Model, error) {
				if strings.Contains(tt.name, "LLM creation fails") {
					return nil, tt.mockLLMError
				}
				return mock, nil
			}

			gotHighlights, err := SuggestHighlights(ctx, tt.manuscript, tt.aiConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestHighlights() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				// Check if the underlying error is what we expect, even if wrapped by the retry logic
				if tt.mockLLMError != nil { // If we are testing an LLM error scenario
					if !strings.Contains(err.Error(), tt.mockLLMError.Error()) {
						t.Errorf("SuggestHighlights() error = %q, want to contain underlying error %q", err.Error(), tt.mockLLMError.Error())
					}
				} else if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					// For non-LLM errors (e.g., config errors, empty/malformed JSON response after successful LLM call)
					t.Errorf("SuggestHighlights() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("SuggestHighlights() unexpected error = %v", err)
					return
				}
			}

			if !reflect.DeepEqual(gotHighlights, tt.wantHighlights) {
				t.Errorf("SuggestHighlights() gotHighlights = %v, want %v", gotHighlights, tt.wantHighlights)
			}

			// For tests that are expected to call the LLM, check if it was called.
			// (i.e., not config error tests or llm creation fail tests before the call point)
			if !tt.wantErr && tt.aiConfig.Endpoint != "" && tt.aiConfig.DeploymentName != "" && tt.aiConfig.APIKey != "" && tt.aiConfig.APIVersion != "" && tt.name != "LLM creation fails (mocked by returning error from newLLMClientFuncForHighlights)" {
				if mock.CallCount == 0 {
					t.Errorf("Expected LLM to be called, but CallCount is 0")
				}
			}
		})
	}
}
