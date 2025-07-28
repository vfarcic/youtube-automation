package ai

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)


func TestSuggestHighlights(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript with important keywords and phrases to highlight."

	tests := []struct {
		name              string
		mockResponse      string
		mockError         error
		manuscript        string
		wantHighlights    []string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:           "Successful suggestion - direct array",
			mockResponse:   `["important keywords", "phrases to highlight"]`,
			manuscript:     validManuscript,
			wantHighlights: []string{"important keywords", "phrases to highlight"},
			wantErr:        false,
		},
		{
			name:           "Successful suggestion - object-wrapped array",
			mockResponse:   `{"suggested_highlights": ["keywords", "phrases"]}`,
			manuscript:     validManuscript,
			wantHighlights: []string{"keywords", "phrases"},
			wantErr:        false,
		},
		{
			name:           "Successful suggestion - with JSON code fence",
			mockResponse:   "```json\n{\"suggested_highlights\": [\"fenced keyword\"]}\n```",
			manuscript:     validManuscript,
			wantHighlights: []string{"fenced keyword"},
			wantErr:        false,
		},
		{
			name:              "AI returns empty response",
			mockResponse:      "",
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response",
		},
		{
			name:              "AI returns malformed JSON",
			mockResponse:      `{"suggested_highlights": ["unterminated string]}`,
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "failed to parse JSON response from AI",
		},
		{
			name:              "AI generation error",
			mockError:         fmt.Errorf("AI service unavailable"),
			manuscript:        validManuscript,
			wantHighlights:    nil,
			wantErr:           true,
			expectedErrSubstr: "AI service unavailable",
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

			gotHighlights, err := SuggestHighlights(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestHighlights() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
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
		})
	}
}
