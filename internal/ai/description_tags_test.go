package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
)


func TestSuggestDescriptionTags(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video description tag generation. It needs three specific #tags separated by spaces."

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
			name:         "Successful description tag generation",
			mockResponse: "#go #programming #ai",
			manuscript:   validManuscript,
			wantTags:     "#go #programming #ai",
			wantErr:      false,
		},
		{
			name:         "AI response with leading/trailing spaces",
			mockResponse: "  #go #programming #ai  ",
			manuscript:   validManuscript,
			wantTags:     "#go #programming #ai",
			wantErr:      false,
		},
		{
			name:         "AI returns too few tags (function handles it gracefully)",
			mockResponse: "#go #programming", // Missing one # - but function will only use what it gets
			manuscript:   validManuscript,
			wantTags:     "#go #programming", // Function returns what it gets, up to 3
			wantErr:      false,
		},
		{
			name:              "AI returns empty response for description tags",
			mockResponse:      "",
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for description tags",
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
			mockError:         fmt.Errorf("mock AI generation error for desc tags"),
			manuscript:        validManuscript,
			wantTags:          "",
			wantErr:           true,
			expectedErrSubstr: "mock AI generation error for desc tags",
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

			gotTags, err := SuggestDescriptionTags(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestDescriptionTags() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
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
