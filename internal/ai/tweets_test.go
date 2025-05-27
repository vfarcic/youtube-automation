package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// mockLLM is assumed to be defined in another test file in this package (e.g., highlights_test.go)
// and accessible here. If not, it would need to be defined or imported.

func TestSuggestTweets(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered tweet generation. It should result in 5 tweet suggestions."
	validAIConfig := AITitleGeneratorConfig{ // Assuming AITitleGeneratorConfig is accessible from titles.go
		Endpoint:       "https://fake-tweets-endpoint.openai.azure.com/",
		DeploymentName: "fake-tweets-deployment",
		APIKey:         "fake-tweets-api-key",
		APIVersion:     "2023-07-01-preview",
	}

	// Backup and restore original newLLMClientFuncForTweets
	originalNewLLMTweets := newLLMClientFuncForTweets
	defer func() { newLLMClientFuncForTweets = originalNewLLMTweets }()

	tests := []struct {
		name              string
		mockLLMResponse   string
		mockLLMError      error
		aiConfig          AITitleGeneratorConfig
		manuscript        string
		wantTweets        []string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:            "Successful tweet generation - 5 tweets",
			mockLLMResponse: `["Tweet 1: Exciting new video! 🤩 #Exciting #NewVideo\n\n[YOUTUBE]", "Tweet 2: Check this out. 👀 #CheckItOut #Tech\n\n[YOUTUBE]", "Tweet 3: AI powered content. 🤖 #AI #Content\n\n[YOUTUBE]", "Tweet 4: Don't miss it! 🚀 #MustWatch #YouTube\n\n[YOUTUBE]", "Tweet 5: Watch now! 🎬 #Video #Premier\n\n[YOUTUBE]"]`,
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTweets:      []string{"Tweet 1: Exciting new video! 🤩 #Exciting #NewVideo\n\n[YOUTUBE]", "Tweet 2: Check this out. 👀 #CheckItOut #Tech\n\n[YOUTUBE]", "Tweet 3: AI powered content. 🤖 #AI #Content\n\n[YOUTUBE]", "Tweet 4: Don't miss it! 🚀 #MustWatch #YouTube\n\n[YOUTUBE]", "Tweet 5: Watch now! 🎬 #Video #Premier\n\n[YOUTUBE]"},
			wantErr:         false,
		},
		{
			name:            "Successful tweet generation - 3 tweets (flexible count)",
			mockLLMResponse: `["Tweet 1: Short and sweet. 👍 #Short #Sweet\n\n[YOUTUBE]", "Tweet 2: Another one. 🎉 #AnotherOne #Fun\n\n[YOUTUBE]", "Tweet 3: Final idea. ✨ #Final #Idea\n\n[YOUTUBE]"]`,
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTweets:      []string{"Tweet 1: Short and sweet. 👍 #Short #Sweet\n\n[YOUTUBE]", "Tweet 2: Another one. 🎉 #AnotherOne #Fun\n\n[YOUTUBE]", "Tweet 3: Final idea. ✨ #Final #Idea\n\n[YOUTUBE]"},
			wantErr:         false, // Function should be flexible if it doesn't get exactly 5
		},
		{
			name:            "AI response with markdown code fences",
			mockLLMResponse: "```json\n[\"Tweet A 👍 #Emoji #Tag\\n\\n[YOUTUBE]\", \"Tweet B 🎉 #Another #Example\\n\\n[YOUTUBE]\"]\n```",
			aiConfig:        validAIConfig,
			manuscript:      validManuscript,
			wantTweets:      []string{"Tweet A 👍 #Emoji #Tag\n\n[YOUTUBE]", "Tweet B 🎉 #Another #Example\n\n[YOUTUBE]"},
			wantErr:         false,
		},
		{
			name:              "AI returns empty JSON array",
			mockLLMResponse:   `[]`,
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty list of tweets",
		},
		{
			name:              "AI returns malformed JSON",
			mockLLMResponse:   `["Tweet 1", "Tweet 2"`, // Missing closing bracket
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "failed to parse AI response for tweets as JSON",
		},
		{
			name:              "AI returns a tweet exceeding 280 characters",
			mockLLMResponse:   fmt.Sprintf(`["Normal tweet 👍 #Normal #Tag\n\n[YOUTUBE]", "%s 👍 #Long #Tweet\n\n[YOUTUBE]"]`, strings.Repeat("a", 260)),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned a tweet exceeding 280 characters",
		},
		{
			name:              "AI returns empty response string",
			mockLLMResponse:   "",
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for tweets",
		},
		{
			name:              "Manuscript is empty",
			mockLLMResponse:   "", // LLM won't be called
			aiConfig:          validAIConfig,
			manuscript:        "  ", // Empty after trim
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "LLM generation fails",
			mockLLMError:      fmt.Errorf("mock LLM generation error for tweets"),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "mock LLM generation error for tweets",
		},
		{
			name:              "LLM client creation fails",
			mockLLMError:      fmt.Errorf("mock LLM client creation error for tweets"),
			aiConfig:          validAIConfig,
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "failed to create LangChainGo client for tweets",
		},
		{
			name:              "Incomplete AI config - no API key",
			aiConfig:          AITitleGeneratorConfig{Endpoint: "e", DeploymentName: "d", APIVersion: "v"},
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI configuration is not fully set for tweets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockLLM{ // Assumes mockLLM struct has ResponseContent and ErrorToReturn fields
				ResponseContent: tt.mockLLMResponse,
				ErrorToReturn:   tt.mockLLMError,
			}

			newLLMClientFuncForTweets = func(options ...openai.Option) (llms.Model, error) {
				if strings.Contains(tt.name, "LLM client creation fails") {
					return nil, tt.mockLLMError
				}
				return mock, nil
			}

			gotTweets, err := SuggestTweets(ctx, tt.manuscript, tt.aiConfig)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestTweets() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("SuggestTweets() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("SuggestTweets() unexpected error = %v", err)
				}
				if !equalStringSlices(gotTweets, tt.wantTweets) {
					t.Errorf("SuggestTweets() gotTweets = %v, want %v", gotTweets, tt.wantTweets)
				}
			}
		})
	}
}

// equalStringSlices is a helper to compare two []string.
// Consider moving to a shared test utility if used across multiple test files.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
