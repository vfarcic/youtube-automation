package ai

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// MockProvider implements AIProvider for testing
type MockProvider struct {
	response string
	err      error
}

func (m *MockProvider) GenerateContent(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestSuggestTitles(t *testing.T) {
	ctx := context.Background()
	validManuscript := "This is a test manuscript about AI-powered video title generation. It needs multiple suggestions."

	tests := []struct {
		name              string
		mockResponse      string
		mockError         error
		manuscript        string
		wantTitles        []string
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name:         "Successful titles suggestion - direct array",
			mockResponse: `["AI Title 1", "Catchy Title 2"]`,
			manuscript:   validManuscript,
			wantTitles:   []string{"AI Title 1", "Catchy Title 2"},
			wantErr:      false,
		},
		{
			name:         "Successful titles suggestion - with JSON code fence",
			mockResponse: "```json\n[\"Fenced Title A\"]\n```",
			manuscript:   validManuscript,
			wantTitles:   []string{"Fenced Title A"},
			wantErr:      false,
		},
		{
			name:              "AI returns empty response for titles",
			mockResponse:      "",
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI returned an empty response for titles",
		},
		{
			name:              "AI returns malformed JSON for titles",
			mockResponse:      `["Unterminated title array]`,
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "failed to parse JSON response",
		},
		{
			name:              "AI generation error for titles",
			mockError:         fmt.Errorf("AI titles service unavailable"),
			manuscript:        validManuscript,
			wantTitles:        nil,
			wantErr:           true,
			expectedErrSubstr: "AI titles service unavailable",
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

			gotTitles, err := SuggestTitles(ctx, tt.manuscript)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SuggestTitles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("SuggestTitles() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("SuggestTitles() unexpected error = %v", err)
					return
				}
			}

			if !reflect.DeepEqual(gotTitles, tt.wantTitles) {
				t.Errorf("SuggestTitles() gotTitles = %v, want %v", gotTitles, tt.wantTitles)
			}
		})
	}
}

func TestGetAIConfig(t *testing.T) {
	// Test the compatibility function
	config, err := GetAIConfig()
	if err != nil {
		t.Errorf("GetAIConfig() unexpected error = %v", err)
	}
	if config == nil {
		t.Errorf("GetAIConfig() returned nil config")
	}
}

func TestTitlesTemplateExecution(t *testing.T) {
	// Test that the template executes correctly with various manuscript inputs
	ctx := context.Background()

	tests := []struct {
		name       string
		manuscript string
		wantErr    bool
	}{
		{
			name:       "Normal manuscript",
			manuscript: "This is a test manuscript",
			wantErr:    false,
		},
		{
			name:       "Manuscript with special characters",
			manuscript: "Test <script>alert('xss')</script> & \"quotes\" 'single'",
			wantErr:    false,
		},
		{
			name:       "Empty manuscript",
			manuscript: "",
			wantErr:    false,
		},
		{
			name:       "Very long manuscript",
			manuscript: strings.Repeat("A", 10000),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: `["Test Title 1", "Test Title 2"]`,
				err:      nil,
			}

			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			titles, err := SuggestTitles(ctx, tt.manuscript)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && len(titles) != 2 {
				t.Errorf("Expected 2 titles, got %d", len(titles))
			}
		})
	}
}
