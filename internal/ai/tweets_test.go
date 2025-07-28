package ai

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// mockLLM is assumed to be defined in another test file in this package (e.g., highlights_test.go)
// and accessible here. If not, it would need to be defined or imported.

func TestSuggestTweets(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered tweet generation. It should result in 5 tweet suggestions."

	tests := []struct {
		name              string
		mockResponse      string
		mockError         error
		manuscript        string
		wantTweets        []string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:         "Successful tweet generation - 5 tweets",
			mockResponse: `["Tweet 1: Exciting new video! 🤩 #Exciting #NewVideo\n\n[YOUTUBE]", "Tweet 2: Check this out. 👀 #CheckItOut #Tech\n\n[YOUTUBE]", "Tweet 3: AI powered content. 🤖 #AI #Content\n\n[YOUTUBE]", "Tweet 4: Don't miss it! 🚀 #MustWatch #YouTube\n\n[YOUTUBE]", "Tweet 5: Watch now! 🎬 #Video #Premier\n\n[YOUTUBE]"]`,
			manuscript:   validManuscript,
			wantTweets:   []string{"Tweet 1: Exciting new video! 🤩 #Exciting #NewVideo\n\n[YOUTUBE]", "Tweet 2: Check this out. 👀 #CheckItOut #Tech\n\n[YOUTUBE]", "Tweet 3: AI powered content. 🤖 #AI #Content\n\n[YOUTUBE]", "Tweet 4: Don't miss it! 🚀 #MustWatch #YouTube\n\n[YOUTUBE]", "Tweet 5: Watch now! 🎬 #Video #Premier\n\n[YOUTUBE]"},
			wantErr:      false,
		},
		{
			name:         "AI response with markdown code fences",
			mockResponse: "```json\n[\"Tweet A 👍 #Emoji #Tag\\n\\n[YOUTUBE]\", \"Tweet B 🎉 #Another #Example\\n\\n[YOUTUBE]\"]\n```",
			manuscript:   validManuscript,
			wantTweets:   []string{"Tweet A 👍 #Emoji #Tag\n\n[YOUTUBE]", "Tweet B 🎉 #Another #Example\n\n[YOUTUBE]"},
			wantErr:      false,
		},
		{
			name:              "AI returns malformed JSON",
			mockResponse:      `["Tweet 1", "Tweet 2"`, // Missing closing bracket
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "failed to parse JSON response",
		},
		{
			name:              "AI returns empty response string",
			mockResponse:      "",
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for tweets",
		},
		{
			name:              "Manuscript is empty",
			mockResponse:      "", // LLM won't be called
			manuscript:        "  ", // Empty after trim
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "manuscript content is empty",
		},
		{
			name:              "AI generation fails",
			mockError:         fmt.Errorf("mock AI generation error for tweets"),
			manuscript:        validManuscript,
			wantTweets:        nil,
			wantErr:           true,
			expectedErrSubstr: "mock AI generation error for tweets",
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

			gotTweets, err := SuggestTweets(ctx, tt.manuscript)

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
				if !reflect.DeepEqual(gotTweets, tt.wantTweets) {
					t.Errorf("SuggestTweets() gotTweets = %v, want %v", gotTweets, tt.wantTweets)
				}
			}
		})
	}
}

