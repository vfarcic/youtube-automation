package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestSuggestDescription(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video description generation."

	tests := []struct {
		name              string
		mockResponse      string
		mockError         error
		manuscript        string
		wantDescription   string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:            "Successful description suggestion",
			mockResponse:    "This is a great AI-generated description.",
			manuscript:      validManuscript,
			wantDescription: "This is a great AI-generated description.",
			wantErr:         false,
		},
		{
			name:              "AI returns empty response for description",
			mockResponse:      "",
			manuscript:        validManuscript,
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty description",
		},
		{
			name:              "AI generation error for description",
			mockError:         fmt.Errorf("AI description service unavailable"),
			manuscript:        validManuscript,
			wantDescription:   "",
			wantErr:           true,
			expectedErrSubstr: "AI description service unavailable",
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

			gotDescription, err := SuggestDescription(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestDescription() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
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
		})
	}
}
